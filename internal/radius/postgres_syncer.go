package radius

import (
	"context"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresSyncer struct {
	pool *pgxpool.Pool
}

func NewPostgresSyncer(db *database.Handle) *PostgresSyncer {
	return &PostgresSyncer{pool: db.Pool()}
}

func (s *PostgresSyncer) ProvisionTicket(ctx context.Context, ticket Ticket) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `
INSERT INTO radius_users (id, username, cleartext_password, is_active, expires_at)
VALUES ($1, $2, $3, true, $4)
ON CONFLICT (username) DO UPDATE
SET cleartext_password = EXCLUDED.cleartext_password,
    is_active = true,
    expires_at = EXCLUDED.expires_at,
    updated_at = now()
`, ticket.ID, ticket.Username, ticket.CleartextPassword, ticket.ValidUntil); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
DELETE FROM radcheck
WHERE username = $1
  AND attribute IN ('Cleartext-Password', 'Expiration')
`, ticket.Username); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO radcheck (username, attribute, op, value)
VALUES ($1, 'Cleartext-Password', ':=', $2)
`, ticket.Username, ticket.CleartextPassword); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO radcheck (username, attribute, op, value)
VALUES ($1, 'Expiration', ':=', $2)
`, ticket.Username, formatExpiration(ticket.ValidUntil)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PostgresSyncer) RevokeTicket(ctx context.Context, ticket Ticket) error {
	return s.removeCredentials(ctx, ticket.Username)
}

func (s *PostgresSyncer) DeleteExpiredTicket(ctx context.Context, ticket Ticket) error {
	return s.removeCredentials(ctx, ticket.Username)
}

func (s *PostgresSyncer) removeCredentials(ctx context.Context, username string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `
DELETE FROM radcheck
WHERE username = $1
`, username); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
DELETE FROM radreply
WHERE username = $1
`, username); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
DELETE FROM radusergroup
WHERE username = $1
`, username); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
UPDATE radius_users
SET cleartext_password = NULL,
    is_active = false,
    updated_at = now()
WHERE username = $1
`, username); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func formatExpiration(value time.Time) string {
	return value.UTC().Format("02 Jan 2006 15:04:05 UTC")
}
