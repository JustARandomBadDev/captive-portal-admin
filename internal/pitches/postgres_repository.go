package pitches

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPitchNotFound = errors.New("pitch not found")
	ErrDuplicateCode = errors.New("pitch code already exists")
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(db *database.Handle) *PostgresRepository {
	return &PostgresRepository{pool: db.Pool()}
}

func (r *PostgresRepository) Create(ctx context.Context, input PitchCreateInput) (Pitch, error) {
	id, err := database.NewUUID()
	if err != nil {
		return Pitch{}, err
	}

	row := r.pool.QueryRow(ctx, `
INSERT INTO pitches (id, code, label)
VALUES ($1, $2, $3)
RETURNING id::text, code, label, is_active, created_at, updated_at
`, id, input.Code, nullableString(input.Label))

	pitch, err := scanPitch(row)
	if err != nil {
		return Pitch{}, mapPostgresError(err)
	}

	return pitch, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (Pitch, error) {
	row := r.pool.QueryRow(ctx, pitchSelectSQL()+` WHERE id = $1`, id)

	pitch, err := scanPitch(row)
	if err != nil {
		return Pitch{}, mapPostgresError(err)
	}

	return pitch, nil
}

func (r *PostgresRepository) ListActive(ctx context.Context) ([]Pitch, error) {
	rows, err := r.pool.Query(ctx, pitchSelectSQL()+`
WHERE is_active = true
ORDER BY code ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPitches(rows)
}

func (r *PostgresRepository) ListAll(ctx context.Context) ([]Pitch, error) {
	rows, err := r.pool.Query(ctx, pitchSelectSQL()+`
ORDER BY code ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPitches(rows)
}

func (r *PostgresRepository) Update(ctx context.Context, input PitchUpdateInput) (Pitch, error) {
	row := r.pool.QueryRow(ctx, `
UPDATE pitches
SET code = $2,
    label = $3,
    is_active = $4,
    updated_at = now()
WHERE id = $1
RETURNING id::text, code, label, is_active, created_at, updated_at
`, input.ID, input.Code, nullableString(input.Label), input.IsActive)

	pitch, err := scanPitch(row)
	if err != nil {
		return Pitch{}, mapPostgresError(err)
	}

	return pitch, nil
}

func (r *PostgresRepository) Disable(ctx context.Context, id string) (Pitch, error) {
	row := r.pool.QueryRow(ctx, `
UPDATE pitches
SET is_active = false,
    updated_at = now()
WHERE id = $1
RETURNING id::text, code, label, is_active, created_at, updated_at
`, id)

	pitch, err := scanPitch(row)
	if err != nil {
		return Pitch{}, mapPostgresError(err)
	}

	return pitch, nil
}

func pitchSelectSQL() string {
	return `
SELECT id::text, code, label, is_active, created_at, updated_at
FROM pitches
`
}

func scanPitches(rows pgx.Rows) ([]Pitch, error) {
	pitches := make([]Pitch, 0)
	for rows.Next() {
		pitch, err := scanPitch(rows)
		if err != nil {
			return nil, mapPostgresError(err)
		}
		pitches = append(pitches, pitch)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pitches, nil
}

type pitchScanner interface {
	Scan(dest ...any) error
}

func scanPitch(scanner pitchScanner) (Pitch, error) {
	var pitch Pitch
	var label sql.NullString

	if err := scanner.Scan(
		&pitch.ID,
		&pitch.Code,
		&label,
		&pitch.IsActive,
		&pitch.CreatedAt,
		&pitch.UpdatedAt,
	); err != nil {
		return Pitch{}, err
	}

	if label.Valid {
		pitch.Label = label.String
	}

	return pitch, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func mapPostgresError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPitchNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" && pgErr.ConstraintName == "pitches_code_key" {
			return ErrDuplicateCode
		}
	}

	return err
}
