package pitches

import "context"

// Repository defines pitch persistence without binding services to PostgreSQL.
type Repository interface {
	Create(ctx context.Context, input PitchCreateInput) (Pitch, error)
	GetByID(ctx context.Context, id string) (Pitch, error)
	ListActive(ctx context.Context) ([]Pitch, error)
	ListAll(ctx context.Context) ([]Pitch, error)
	Update(ctx context.Context, input PitchUpdateInput) (Pitch, error)
	Disable(ctx context.Context, id string) (Pitch, error)
}
