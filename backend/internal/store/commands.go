package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/samfiallos/owenone/backend/internal/app"
	"github.com/samfiallos/owenone/backend/internal/domain"
)

func (store *Store) JoinWaitlist(ctx context.Context, email string) (bool, error) {
	result, err := store.pool.Exec(ctx, `
		INSERT INTO waitlist_entries (email) VALUES ($1)
		ON CONFLICT DO NOTHING`, email)
	if err != nil {
		return false, fmt.Errorf("join waitlist: %w", err)
	}
	return result.RowsAffected() == 1, nil
}

func (store *Store) CreateGroup(ctx context.Context, command app.CreateGroupCommand) error {
	transaction, err := store.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create group: %w", err)
	}
	defer transaction.Rollback(ctx) //nolint:errcheck

	if len(command.MemberIDs) > 0 {
		var friendCount int
		if err := transaction.QueryRow(ctx, `
			SELECT count(*)
			FROM friendships
			WHERE (
				user_id = $1::uuid AND friend_id::text = ANY($2::text[])
			) OR (
				friend_id = $1::uuid AND user_id::text = ANY($2::text[])
			)`, command.CreatorID, command.MemberIDs).Scan(&friendCount); err != nil {
			return fmt.Errorf("check group friends: %w", err)
		}
		if friendCount != len(command.MemberIDs) {
			return fmt.Errorf("%w: group members must be existing friends", app.ErrForbidden)
		}
	}

	if _, err := transaction.Exec(ctx, `
		INSERT INTO groups (id, name, icon, created_by)
		VALUES ($1::uuid, $2, $3, $4::uuid)`, command.ID, command.Name, command.Icon, command.CreatorID); err != nil {
		return fmt.Errorf("insert group: %w", err)
	}
	if _, err := transaction.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, role)
		VALUES ($1::uuid, $2::uuid, 'owner')`, command.ID, command.CreatorID); err != nil {
		return fmt.Errorf("insert group owner: %w", err)
	}
	for _, memberID := range command.MemberIDs {
		if _, err := transaction.Exec(ctx, `
			INSERT INTO group_members (group_id, user_id, role)
			VALUES ($1::uuid, $2::uuid, 'member')`, command.ID, memberID); err != nil {
			return fmt.Errorf("insert group member: %w", err)
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit create group: %w", err)
	}
	return nil
}

func (store *Store) AddFriend(ctx context.Context, actorID, email string) (domain.User, error) {
	var friend domain.User
	err := store.pool.QueryRow(ctx, `
		SELECT id::text, display_name
		FROM users
		WHERE lower(email) = lower($1)`, email).Scan(&friend.ID, &friend.DisplayName)
	if err != nil {
		return domain.User{}, fmt.Errorf("find friend: %w", mapNotFound(err))
	}
	if friend.ID == actorID {
		return domain.User{}, fmt.Errorf("%w: cannot add yourself as a friend", app.ErrInvalidInput)
	}

	firstID, secondID := actorID, friend.ID
	if strings.Compare(firstID, secondID) > 0 {
		firstID, secondID = secondID, firstID
	}
	if _, err := store.pool.Exec(ctx, `
		INSERT INTO friendships (user_id, friend_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT DO NOTHING`, firstID, secondID); err != nil {
		return domain.User{}, fmt.Errorf("add friend: %w", err)
	}
	return friend, nil
}

func (store *Store) CreateExpense(ctx context.Context, command app.CreateExpenseCommand) (string, error) {
	transaction, err := store.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin create expense: %w", err)
	}
	defer transaction.Rollback(ctx) //nolint:errcheck

	memberIDs := []string{command.ActorID, command.PaidByUserID}
	for _, split := range command.Splits {
		memberIDs = append(memberIDs, split.UserID)
	}
	memberIDs = uniqueStrings(memberIDs)
	var memberCount int
	if err := transaction.QueryRow(ctx, `
		SELECT count(*)
		FROM group_members
		WHERE group_id = $1::uuid AND user_id::text = ANY($2::text[])`, command.GroupID, memberIDs).Scan(&memberCount); err != nil {
		return "", fmt.Errorf("check expense group members: %w", err)
	}
	if memberCount != len(memberIDs) {
		return "", fmt.Errorf("%w: every expense participant must belong to the group", app.ErrForbidden)
	}

	var expenseID string
	err = transaction.QueryRow(ctx, `
		INSERT INTO expenses (
			id, group_id, description, category, amount_minor, currency, paid_by,
			split_method, created_by, idempotency_key, request_fingerprint
		) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7::uuid, $8, $9::uuid, $10, $11)
		ON CONFLICT (created_by, idempotency_key) DO NOTHING
		RETURNING id::text`,
		command.ID, command.GroupID, command.Description, command.Category,
		command.Money.AmountMinor, command.Money.Currency, command.PaidByUserID,
		command.SplitMethod, command.ActorID, command.IdempotencyKey, command.RequestFingerprint,
	).Scan(&expenseID)
	if errors.Is(err, pgx.ErrNoRows) {
		var existingFingerprint string
		if err := transaction.QueryRow(ctx, `
			SELECT id::text, request_fingerprint FROM expenses WHERE created_by = $1::uuid AND idempotency_key = $2`,
			command.ActorID, command.IdempotencyKey,
		).Scan(&expenseID, &existingFingerprint); err != nil {
			return "", fmt.Errorf("load idempotent expense: %w", err)
		}
		if existingFingerprint != command.RequestFingerprint {
			return "", fmt.Errorf("%w: Idempotency-Key was already used for a different expense", ErrConflict)
		}
		if err := transaction.Commit(ctx); err != nil {
			return "", fmt.Errorf("commit idempotent expense: %w", err)
		}
		return expenseID, nil
	}
	if err != nil {
		return "", fmt.Errorf("insert expense: %w", err)
	}

	for _, split := range command.Splits {
		if _, err := transaction.Exec(ctx, `
			INSERT INTO expense_splits (expense_id, user_id, amount_minor)
			VALUES ($1::uuid, $2::uuid, $3)`, expenseID, split.UserID, split.AmountMinor); err != nil {
			return "", fmt.Errorf("insert expense split: %w", err)
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit create expense: %w", err)
	}
	return expenseID, nil
}

func (store *Store) CreateSettlement(ctx context.Context, command app.CreateSettlementCommand) (string, error) {
	connected, err := store.usersConnected(ctx, command.FromUserID, command.ToUserID, command.GroupID)
	if err != nil {
		return "", err
	}
	if !connected {
		return "", fmt.Errorf("%w: settlement users are not connected", app.ErrForbidden)
	}

	transaction, err := store.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin create settlement: %w", err)
	}
	defer transaction.Rollback(ctx) //nolint:errcheck

	var settlementID string
	err = transaction.QueryRow(ctx, `
		INSERT INTO settlements (
			id, group_id, from_user_id, to_user_id, amount_minor, currency, status,
			idempotency_key, request_fingerprint
		) VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, 'completed', $7, $8)
		ON CONFLICT (from_user_id, idempotency_key) DO NOTHING
		RETURNING id::text`,
		command.ID, command.GroupID, command.FromUserID, command.ToUserID,
		command.Money.AmountMinor, command.Money.Currency, command.IdempotencyKey, command.RequestFingerprint,
	).Scan(&settlementID)
	if errors.Is(err, pgx.ErrNoRows) {
		var existingFingerprint string
		if err := transaction.QueryRow(ctx, `
			SELECT id::text, request_fingerprint
			FROM settlements
			WHERE from_user_id = $1::uuid AND idempotency_key = $2`,
			command.FromUserID, command.IdempotencyKey,
		).Scan(&settlementID, &existingFingerprint); err != nil {
			return "", fmt.Errorf("load idempotent settlement: %w", err)
		}
		if existingFingerprint != command.RequestFingerprint {
			return "", fmt.Errorf("%w: Idempotency-Key was already used for a different settlement", ErrConflict)
		}
	} else if err != nil {
		return "", fmt.Errorf("create settlement: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit create settlement: %w", err)
	}
	return settlementID, nil
}

func (store *Store) CreateReminder(ctx context.Context, command app.CreateReminderCommand) (string, error) {
	connected, err := store.usersConnected(ctx, command.SenderID, command.RecipientID, nil)
	if err != nil {
		return "", err
	}
	if !connected {
		return "", fmt.Errorf("%w: reminder recipient is not connected", app.ErrForbidden)
	}
	if _, err := store.pool.Exec(ctx, `
		INSERT INTO reminders (id, sender_id, recipient_id, message)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4)`, command.ID, command.SenderID, command.RecipientID, command.Message); err != nil {
		return "", fmt.Errorf("create reminder: %w", err)
	}
	return command.ID, nil
}

func (store *Store) usersConnected(ctx context.Context, firstUserID, secondUserID string, groupID *string) (bool, error) {
	var connected bool
	if groupID != nil {
		err := store.pool.QueryRow(ctx, `
			SELECT count(*) = 2
			FROM group_members
			WHERE group_id = $1::uuid AND user_id IN ($2::uuid, $3::uuid)`,
			*groupID, firstUserID, secondUserID,
		).Scan(&connected)
		if err != nil {
			return false, fmt.Errorf("check shared group: %w", err)
		}
		return connected, nil
	}

	err := store.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM friendships
			WHERE user_id IN ($1::uuid, $2::uuid) AND friend_id IN ($1::uuid, $2::uuid)
		) OR EXISTS (
			SELECT 1
			FROM group_members first_membership
			JOIN group_members second_membership ON second_membership.group_id = first_membership.group_id
			WHERE first_membership.user_id = $1::uuid AND second_membership.user_id = $2::uuid
		)`, firstUserID, secondUserID).Scan(&connected)
	if err != nil {
		return false, fmt.Errorf("check connected users: %w", err)
	}
	return connected, nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
