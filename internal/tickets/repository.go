package tickets

import (
	"context"
	"time"
)

// Repository defines ticket persistence without binding services to PostgreSQL.
type Repository interface {
	Create(ctx context.Context, input TicketCreateInput) (Ticket, error)
	GetByID(ctx context.Context, id string) (Ticket, error)
	ListActive(ctx context.Context, now time.Time) ([]Ticket, error)
	ListAll(ctx context.Context) ([]Ticket, error)
	Revoke(ctx context.Context, input TicketRevokeInput) (Ticket, error)
	MarkExpired(ctx context.Context, now time.Time) (int, error)
	DeleteOldExpired(ctx context.Context, before time.Time) (int, error)
}
