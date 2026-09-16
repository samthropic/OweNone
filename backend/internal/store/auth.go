package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samfiallos/owenone/backend/internal/app"
	"github.com/samfiallos/owenone/backend/internal/domain"
)

func (store *Store) CreateUser(ctx context.Context, command app.CreateUserCommand) (domain.User, error) {
	var user domain.User
	var venmo, paypal, cashapp, zelle, avatarURL string
	err := store.pool.QueryRow(ctx, `
		INSERT INTO users (id, email, display_name, preferred_currency, password_hash)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING id::text, email, display_name, preferred_currency, `+userAvatarColumn+`, `+userPaymentColumns,
		command.ID, command.Email, command.DisplayName, command.PreferredCurrency, command.PasswordHash,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PreferredCurrency, &avatarURL, &venmo, &paypal, &cashapp, &zelle)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, fmt.Errorf("create user: %w", ErrConflict)
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	applyAvatar(&user, avatarURL)
	user.PaymentApps = paymentAppsFromScan(venmo, paypal, cashapp, zelle)
	return user, nil
}

func (store *Store) UserCredentialsByEmail(ctx context.Context, email string) (app.UserCredentials, error) {
	var creds app.UserCredentials
	var venmo, paypal, cashapp, zelle, avatarURL string
	err := store.pool.QueryRow(ctx, `
		SELECT id::text, email, display_name, preferred_currency, password_hash, `+userAvatarColumn+`, `+userPaymentColumns+`
		FROM users
		WHERE lower(email) = lower($1)`, email,
	).Scan(&creds.User.ID, &creds.User.Email, &creds.User.DisplayName, &creds.User.PreferredCurrency, &creds.PasswordHash, &avatarURL, &venmo, &paypal, &cashapp, &zelle)
	if err != nil {
		return app.UserCredentials{}, fmt.Errorf("user credentials: %w", mapNotFound(err))
	}
	applyAvatar(&creds.User, avatarURL)
	creds.User.PaymentApps = paymentAppsFromScan(venmo, paypal, cashapp, zelle)
	return creds, nil
}

func (store *Store) CreateSession(ctx context.Context, command app.CreateSessionCommand) error {
	if _, err := store.pool.Exec(ctx, `
		INSERT INTO sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2::uuid, $3)`,
		command.TokenHash, command.UserID, command.ExpiresAt,
	); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	// Opportunistically clean up expired sessions.
	_, _ = store.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= now()`)
	return nil
}

func (store *Store) SessionUser(ctx context.Context, tokenHash string) (domain.User, error) {
	var user domain.User
	var venmo, paypal, cashapp, zelle, avatarURL string
	err := store.pool.QueryRow(ctx, `
		SELECT u.id::text, u.email, u.display_name, u.preferred_currency, `+
		`COALESCE(u.avatar_url, ''), COALESCE(u.venmo_handle, ''), COALESCE(u.paypal_handle, ''), COALESCE(u.cashapp_handle, ''), COALESCE(u.zelle_handle, '')
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()`, tokenHash,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PreferredCurrency, &avatarURL, &venmo, &paypal, &cashapp, &zelle)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("session user: %w", app.ErrSessionNotFound)
		}
		return domain.User{}, fmt.Errorf("session user: %w", err)
	}
	applyAvatar(&user, avatarURL)
	user.PaymentApps = paymentAppsFromScan(venmo, paypal, cashapp, zelle)
	return user, nil
}

func (store *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	if _, err := store.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (store *Store) UpdateUser(ctx context.Context, command app.UpdateUserCommand) (domain.User, error) {
	var user domain.User
	var venmo, paypal, cashapp, zelle, avatarURL string
	err := store.pool.QueryRow(ctx, `
		UPDATE users
		SET email = $2, display_name = $3,
		    preferred_currency = COALESCE(NULLIF($4::text, ''), preferred_currency),
		    venmo_handle = $5, paypal_handle = $6, cashapp_handle = $7, zelle_handle = $8
		WHERE id = $1::uuid
		RETURNING id::text, email, display_name, preferred_currency, `+userAvatarColumn+`, `+userPaymentColumns,
		command.ID, command.Email, command.DisplayName, command.PreferredCurrency,
		nullableHandle(command.PaymentApps.Venmo),
		nullableHandle(command.PaymentApps.PayPal),
		nullableHandle(command.PaymentApps.CashApp),
		nullableHandle(command.PaymentApps.Zelle),
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PreferredCurrency, &avatarURL, &venmo, &paypal, &cashapp, &zelle)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, fmt.Errorf("update user: %w", ErrConflict)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("update user: %w", ErrNotFound)
		}
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}
	applyAvatar(&user, avatarURL)
	user.PaymentApps = paymentAppsFromScan(venmo, paypal, cashapp, zelle)
	return user, nil
}

func (store *Store) UpdateAvatarURL(ctx context.Context, userID, avatarURL string) (domain.User, error) {
	var user domain.User
	var venmo, paypal, cashapp, zelle, storedAvatar string
	err := store.pool.QueryRow(ctx, `
		UPDATE users
		SET avatar_url = $2
		WHERE id = $1::uuid
		RETURNING id::text, email, display_name, preferred_currency, `+userAvatarColumn+`, `+userPaymentColumns,
		userID, nullableAvatar(avatarURL),
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PreferredCurrency, &storedAvatar, &venmo, &paypal, &cashapp, &zelle)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("update avatar: %w", ErrNotFound)
		}
		return domain.User{}, fmt.Errorf("update avatar: %w", err)
	}
	applyAvatar(&user, storedAvatar)
	user.PaymentApps = paymentAppsFromScan(venmo, paypal, cashapp, zelle)
	return user, nil
}

func (store *Store) UserCredentialsByID(ctx context.Context, userID string) (app.UserCredentials, error) {
	var creds app.UserCredentials
	var venmo, paypal, cashapp, zelle, avatarURL string
	err := store.pool.QueryRow(ctx, `
		SELECT id::text, email, display_name, preferred_currency, password_hash, `+userAvatarColumn+`, `+userPaymentColumns+`
		FROM users
		WHERE id = $1::uuid`, userID,
	).Scan(&creds.User.ID, &creds.User.Email, &creds.User.DisplayName, &creds.User.PreferredCurrency, &creds.PasswordHash, &avatarURL, &venmo, &paypal, &cashapp, &zelle)
	if err != nil {
		return app.UserCredentials{}, fmt.Errorf("user credentials: %w", mapNotFound(err))
	}
	applyAvatar(&creds.User, avatarURL)
	creds.User.PaymentApps = paymentAppsFromScan(venmo, paypal, cashapp, zelle)
	return creds, nil
}

func (store *Store) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	if _, err := store.pool.Exec(ctx, `
		UPDATE users SET password_hash = $2 WHERE id = $1::uuid`,
		userID, passwordHash,
	); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (store *Store) DeleteSessionsForUserExcept(ctx context.Context, userID, keepTokenHash string) error {
	if _, err := store.pool.Exec(ctx, `
		DELETE FROM sessions WHERE user_id = $1::uuid AND token_hash <> $2`,
		userID, keepTokenHash,
	); err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}
	return nil
}
