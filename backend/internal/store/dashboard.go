package store

import (
	"context"
	"fmt"

	"github.com/samfiallos/owenone/backend/internal/domain"
)

func (store *Store) LoadLedgerSnapshot(ctx context.Context, userID string) (domain.LedgerSnapshot, error) {
	currentUser, err := store.loadCurrentUser(ctx, userID)
	if err != nil {
		return domain.LedgerSnapshot{}, err
	}

	snapshot := domain.LedgerSnapshot{CurrentUser: currentUser}
	if snapshot.Groups, err = store.loadGroups(ctx, userID); err != nil {
		return domain.LedgerSnapshot{}, err
	}
	if snapshot.Friends, err = store.loadFriends(ctx, userID); err != nil {
		return domain.LedgerSnapshot{}, err
	}
	if snapshot.Expenses, err = store.loadExpenses(ctx, userID, currentUser.PreferredCurrency); err != nil {
		return domain.LedgerSnapshot{}, err
	}
	if snapshot.Settlements, err = store.loadSettlements(ctx, userID, currentUser.PreferredCurrency); err != nil {
		return domain.LedgerSnapshot{}, err
	}
	return snapshot, nil
}

func (store *Store) loadCurrentUser(ctx context.Context, userID string) (domain.User, error) {
	var user domain.User
	err := store.pool.QueryRow(ctx, `
		SELECT id::text, email, display_name, preferred_currency
		FROM users
		WHERE id = $1::uuid`, userID,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PreferredCurrency)
	if err != nil {
		return domain.User{}, fmt.Errorf("load current user: %w", mapNotFound(err))
	}
	return user, nil
}

func (store *Store) loadGroups(ctx context.Context, userID string) ([]domain.Group, error) {
	rows, err := store.pool.Query(ctx, `
		SELECT g.id::text, g.name, g.icon, g.created_at, member.id::text, member.display_name
		FROM groups g
		JOIN group_members current_membership ON current_membership.group_id = g.id
		JOIN group_members membership ON membership.group_id = g.id
		JOIN users member ON member.id = membership.user_id
		WHERE current_membership.user_id = $1::uuid
		ORDER BY g.created_at DESC, membership.joined_at, member.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("query groups: %w", err)
	}
	defer rows.Close()

	groups := make([]domain.Group, 0)
	groupIndexes := make(map[string]int)
	for rows.Next() {
		var group domain.Group
		var member domain.User
		if err := rows.Scan(&group.ID, &group.Name, &group.Icon, &group.CreatedAt, &member.ID, &member.DisplayName); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		index, exists := groupIndexes[group.ID]
		if !exists {
			index = len(groups)
			groupIndexes[group.ID] = index
			groups = append(groups, group)
		}
		groups[index].Members = append(groups[index].Members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate groups: %w", err)
	}
	return groups, nil
}

func (store *Store) loadFriends(ctx context.Context, userID string) ([]domain.User, error) {
	rows, err := store.pool.Query(ctx, `
		SELECT friend.id::text, friend.display_name
		FROM friendships friendship
		JOIN users friend ON friend.id = CASE
			WHEN friendship.user_id = $1::uuid THEN friendship.friend_id
			ELSE friendship.user_id
		END
		WHERE friendship.user_id = $1::uuid OR friendship.friend_id = $1::uuid
		ORDER BY friend.display_name, friend.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("query friends: %w", err)
	}
	defer rows.Close()

	friends := make([]domain.User, 0)
	for rows.Next() {
		var friend domain.User
		if err := rows.Scan(&friend.ID, &friend.DisplayName); err != nil {
			return nil, fmt.Errorf("scan friend: %w", err)
		}
		friends = append(friends, friend)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate friends: %w", err)
	}
	return friends, nil
}

func (store *Store) loadExpenses(ctx context.Context, userID, currency string) ([]domain.Expense, error) {
	rows, err := store.pool.Query(ctx, `
		SELECT expense.id::text, expense.group_id::text, expense.description, expense.category,
			expense.paid_by::text, expense.created_by::text, expense.amount_minor, expense.currency,
			expense.split_method, expense.created_at, split.user_id::text, split.amount_minor
		FROM expenses expense
		JOIN group_members current_membership ON current_membership.group_id = expense.group_id
		JOIN expense_splits split ON split.expense_id = expense.id
		WHERE current_membership.user_id = $1::uuid AND expense.currency = $2
		ORDER BY expense.created_at DESC, expense.id, split.user_id`, userID, currency)
	if err != nil {
		return nil, fmt.Errorf("query expenses: %w", err)
	}
	defer rows.Close()

	expenses := make([]domain.Expense, 0)
	expenseIndexes := make(map[string]int)
	for rows.Next() {
		var expense domain.Expense
		var split domain.ExpenseSplit
		if err := rows.Scan(
			&expense.ID, &expense.GroupID, &expense.Description, &expense.Category,
			&expense.PayerID, &expense.CreatedByID, &expense.Money.AmountMinor, &expense.Money.Currency,
			&expense.SplitMethod, &expense.CreatedAt, &split.UserID, &split.AmountMinor,
		); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		index, exists := expenseIndexes[expense.ID]
		if !exists {
			index = len(expenses)
			expenseIndexes[expense.ID] = index
			expenses = append(expenses, expense)
		}
		expenses[index].Splits = append(expenses[index].Splits, split)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expenses: %w", err)
	}
	return expenses, nil
}

func (store *Store) loadSettlements(ctx context.Context, userID, currency string) ([]domain.Settlement, error) {
	rows, err := store.pool.Query(ctx, `
		WITH network_users AS (
			SELECT DISTINCT network_membership.user_id
			FROM group_members current_membership
			JOIN group_members network_membership ON network_membership.group_id = current_membership.group_id
			WHERE current_membership.user_id = $1::uuid
		)
		SELECT settlement.id::text, settlement.group_id::text, settlement.from_user_id::text,
			settlement.to_user_id::text, settlement.amount_minor, settlement.currency, settlement.created_at
		FROM settlements settlement
		WHERE settlement.status = 'completed' AND settlement.currency = $2
		AND (
			EXISTS (
				SELECT 1 FROM group_members current_membership
				WHERE current_membership.group_id = settlement.group_id
				AND current_membership.user_id = $1::uuid
			)
			OR (
				settlement.group_id IS NULL
				AND settlement.from_user_id IN (SELECT user_id FROM network_users)
				AND settlement.to_user_id IN (SELECT user_id FROM network_users)
			)
		)
		ORDER BY settlement.created_at DESC, settlement.id`, userID, currency)
	if err != nil {
		return nil, fmt.Errorf("query settlements: %w", err)
	}
	defer rows.Close()

	settlements := make([]domain.Settlement, 0)
	for rows.Next() {
		var settlement domain.Settlement
		if err := rows.Scan(
			&settlement.ID, &settlement.GroupID, &settlement.FromUserID, &settlement.ToUserID,
			&settlement.Money.AmountMinor, &settlement.Money.Currency, &settlement.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan settlement: %w", err)
		}
		settlements = append(settlements, settlement)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate settlements: %w", err)
	}
	return settlements, nil
}
