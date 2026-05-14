package adminauth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAdminNotFound          = errors.New("admin user not found")
	ErrSessionNotFound        = errors.New("admin session not found")
	ErrDuplicateSessionToken  = errors.New("admin session token already exists")
	ErrInvalidSessionCreation = errors.New("invalid admin session creation")
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(db *database.Handle) *PostgresRepository {
	return &PostgresRepository{pool: db.Pool()}
}

func (r *PostgresRepository) GetAdminByUsername(ctx context.Context, username string) (AdminUser, error) {
	row := r.pool.QueryRow(ctx, adminUserSelectSQL()+` WHERE username = $1`, username)

	admin, err := scanAdminUser(row)
	if err != nil {
		return AdminUser{}, mapAdminPostgresError(err)
	}

	return admin, nil
}

func (r *PostgresRepository) GetAdminByID(ctx context.Context, id string) (AdminUser, error) {
	row := r.pool.QueryRow(ctx, adminUserSelectSQL()+` WHERE id = $1`, id)

	admin, err := scanAdminUser(row)
	if err != nil {
		return AdminUser{}, mapAdminPostgresError(err)
	}

	return admin, nil
}

func (r *PostgresRepository) CreateSession(ctx context.Context, session AdminSession) (AdminSession, error) {
	if session.ID == "" {
		id, err := database.NewUUID()
		if err != nil {
			return AdminSession{}, err
		}
		session.ID = id
	}
	if session.AdminUserID == "" || session.TokenHash == "" || session.ExpiresAt.IsZero() {
		return AdminSession{}, ErrInvalidSessionCreation
	}

	row := r.pool.QueryRow(ctx, `
INSERT INTO admin_sessions (
    id, admin_user_id, token_hash, expires_at, last_seen_at, revoked_at
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    id::text, admin_user_id::text, token_hash, created_at,
    expires_at, last_seen_at, revoked_at
`,
		session.ID,
		session.AdminUserID,
		session.TokenHash,
		session.ExpiresAt,
		nullableTime(session.LastSeenAt),
		nullableTime(session.RevokedAt),
	)

	created, err := scanAdminSession(row)
	if err != nil {
		return AdminSession{}, mapSessionPostgresError(err)
	}

	return created, nil
}

func (r *PostgresRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (SessionWithAdmin, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
    s.id::text, s.admin_user_id::text, s.token_hash, s.created_at,
    s.expires_at, s.last_seen_at, s.revoked_at,
    u.id::text, u.username, u.password_hash, u.display_name,
    u.is_active, u.created_at, u.updated_at
FROM admin_sessions s
JOIN admin_users u ON u.id = s.admin_user_id
WHERE s.token_hash = $1
`, tokenHash)

	result, err := scanSessionWithAdmin(row)
	if err != nil {
		return SessionWithAdmin{}, mapSessionPostgresError(err)
	}

	return result, nil
}

func (r *PostgresRepository) RevokeSession(ctx context.Context, sessionID string) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE admin_sessions
SET revoked_at = now()
WHERE id = $1
  AND revoked_at IS NULL
`, sessionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (r *PostgresRepository) TouchSession(ctx context.Context, sessionID string) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE admin_sessions
SET last_seen_at = now()
WHERE id = $1
  AND revoked_at IS NULL
`, sessionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func adminUserSelectSQL() string {
	return `
SELECT
    id::text, username, password_hash, display_name,
    is_active, created_at, updated_at
FROM admin_users
`
}

type adminScanner interface {
	Scan(dest ...any) error
}

func scanAdminUser(scanner adminScanner) (AdminUser, error) {
	var admin AdminUser
	var passwordHash sql.NullString
	var displayName sql.NullString

	if err := scanner.Scan(
		&admin.ID,
		&admin.Username,
		&passwordHash,
		&displayName,
		&admin.IsActive,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	); err != nil {
		return AdminUser{}, err
	}

	if passwordHash.Valid {
		admin.PasswordHash = passwordHash.String
	}
	if displayName.Valid {
		admin.DisplayName = displayName.String
	}

	return admin, nil
}

func scanAdminSession(scanner adminScanner) (AdminSession, error) {
	var session AdminSession
	var lastSeenAt sql.NullTime
	var revokedAt sql.NullTime

	if err := scanner.Scan(
		&session.ID,
		&session.AdminUserID,
		&session.TokenHash,
		&session.CreatedAt,
		&session.ExpiresAt,
		&lastSeenAt,
		&revokedAt,
	); err != nil {
		return AdminSession{}, err
	}

	if lastSeenAt.Valid {
		session.LastSeenAt = &lastSeenAt.Time
	}
	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}

	return session, nil
}

func scanSessionWithAdmin(scanner adminScanner) (SessionWithAdmin, error) {
	var result SessionWithAdmin
	var lastSeenAt sql.NullTime
	var revokedAt sql.NullTime
	var passwordHash sql.NullString
	var displayName sql.NullString

	if err := scanner.Scan(
		&result.Session.ID,
		&result.Session.AdminUserID,
		&result.Session.TokenHash,
		&result.Session.CreatedAt,
		&result.Session.ExpiresAt,
		&lastSeenAt,
		&revokedAt,
		&result.Admin.ID,
		&result.Admin.Username,
		&passwordHash,
		&displayName,
		&result.Admin.IsActive,
		&result.Admin.CreatedAt,
		&result.Admin.UpdatedAt,
	); err != nil {
		return SessionWithAdmin{}, err
	}

	if lastSeenAt.Valid {
		result.Session.LastSeenAt = &lastSeenAt.Time
	}
	if revokedAt.Valid {
		result.Session.RevokedAt = &revokedAt.Time
	}
	if passwordHash.Valid {
		result.Admin.PasswordHash = passwordHash.String
	}
	if displayName.Valid {
		result.Admin.DisplayName = displayName.String
	}

	return result, nil
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func mapAdminPostgresError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAdminNotFound
	}

	return err
}

func mapSessionPostgresError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrSessionNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" && pgErr.ConstraintName == "admin_sessions_token_hash_key" {
			return ErrDuplicateSessionToken
		}
	}

	return err
}
