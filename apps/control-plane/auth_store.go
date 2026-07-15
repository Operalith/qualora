package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CountLocalUsers(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT count(*) FROM local_users`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count local users: %w", err)
	}
	return count, nil
}

func (s *Store) CreateFirstLocalAdmin(ctx context.Context, input SetupAdminRequest, passwordHash string) (*LocalUser, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin first admin creation: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `LOCK TABLE local_users IN EXCLUSIVE MODE`); err != nil {
		return nil, fmt.Errorf("lock local users: %w", err)
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM local_users`).Scan(&count); err != nil {
		return nil, fmt.Errorf("count local users: %w", err)
	}
	if count > 0 {
		return nil, ErrAlreadyExists
	}

	user, err := insertLocalUser(ctx, tx, input.Email, input.DisplayName, passwordHash)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit first admin creation: %w", err)
	}
	return user, nil
}

func (s *Store) GetLocalUserByEmail(ctx context.Context, email string) (*LocalUser, error) {
	user, err := scanLocalUser(s.db.QueryRow(ctx, `
SELECT id, email, display_name, password_hash, role, is_active, last_login_at, created_at, updated_at
FROM local_users
WHERE lower(email) = lower($1)
`, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *Store) MarkLocalUserLogin(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `UPDATE local_users SET last_login_at = now(), updated_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark local user login: %w", err)
	}
	return nil
}

func (s *Store) CreateUserSession(ctx context.Context, userID string, tokenHash string, csrfTokenHash string, userAgent string, ipAddress string, expiresAt time.Time) (*UserSession, error) {
	session := &UserSession{}
	err := s.db.QueryRow(ctx, `
INSERT INTO user_sessions (id, user_id, token_hash, csrf_token_hash, user_agent, ip_address, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, token_hash, csrf_token_hash, user_agent, ip_address, expires_at, created_at, last_seen_at, revoked_at
`, uuid.NewString(), userID, tokenHash, csrfTokenHash, userAgent, ipAddress, expiresAt).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.CSRFTokenHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastSeenAt,
		&session.RevokedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user session: %w", err)
	}
	return session, nil
}

func (s *Store) GetActiveSessionByTokenHash(ctx context.Context, tokenHash string, now time.Time) (*UserSession, *LocalUser, error) {
	row := s.db.QueryRow(ctx, `
SELECT
	s.id, s.user_id::text, s.token_hash, COALESCE(s.csrf_token_hash, ''), s.user_agent, s.ip_address,
	s.expires_at, s.created_at, s.last_seen_at, s.revoked_at,
	u.id, u.email, u.display_name, u.password_hash, u.role, u.is_active, u.last_login_at, u.created_at, u.updated_at
FROM user_sessions s
JOIN local_users u ON u.id = s.user_id
WHERE s.token_hash = $1
	AND s.revoked_at IS NULL
	AND s.expires_at > $2
	AND u.is_active = true
`, tokenHash, now)

	session := &UserSession{}
	user := &LocalUser{}
	var revokedAt sql.NullTime
	var lastLoginAt sql.NullTime
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.CSRFTokenHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastSeenAt,
		&revokedAt,
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("query active session: %w", err)
	}
	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	return session, user, nil
}

func (s *Store) TouchUserSession(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `UPDATE user_sessions SET last_seen_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("touch user session: %w", err)
	}
	return nil
}

func (s *Store) RevokeUserSession(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `UPDATE user_sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("revoke user session: %w", err)
	}
	return nil
}

func insertLocalUser(ctx context.Context, tx pgx.Tx, email string, displayName string, passwordHash string) (*LocalUser, error) {
	user, err := scanLocalUser(tx.QueryRow(ctx, `
INSERT INTO local_users (id, email, display_name, password_hash, role, is_active)
VALUES ($1, $2, $3, $4, 'admin', true)
RETURNING id, email, display_name, password_hash, role, is_active, last_login_at, created_at, updated_at
`, uuid.NewString(), email, displayName, passwordHash))
	if err != nil {
		return nil, fmt.Errorf("insert local user: %w", err)
	}
	return user, nil
}

type localUserScanner interface {
	Scan(dest ...any) error
}

func scanLocalUser(row localUserScanner) (*LocalUser, error) {
	user := &LocalUser{}
	var lastLoginAt sql.NullTime
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	return user, nil
}
