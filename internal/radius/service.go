package radius

import (
	"context"
	"log/slog"
	"time"
)

// Ticket describes the admin-side data needed to provision future FreeRADIUS
// credentials without coupling the tickets package to radius SQL tables.
type Ticket struct {
	ID                string
	Username          string
	CleartextPassword string
	PitchID           string
	ValidFrom         time.Time
	ValidUntil        time.Time
}

// Syncer is the boundary between the admin database and the future
// FreeRADIUS database writer. Implementations must not write legal portal logs.
type Syncer interface {
	ProvisionTicket(ctx context.Context, ticket Ticket) error
	RevokeTicket(ctx context.Context, ticket Ticket) error
	DeleteExpiredTicket(ctx context.Context, ticket Ticket) error
}

// Service keeps RadiusSync orchestration separate from ticket persistence.
// It is intentionally synchronous for now; the same boundary can later be
// moved behind a worker or retry queue without changing ticket repositories.
type Service struct {
	syncer Syncer
}

func NewService(syncer Syncer) *Service {
	if syncer == nil {
		syncer = NoopSyncer{}
	}
	return &Service{syncer: syncer}
}

func (s *Service) ProvisionTicket(ctx context.Context, ticket Ticket) error {
	return s.syncer.ProvisionTicket(ctx, ticket)
}

func (s *Service) RevokeTicket(ctx context.Context, ticket Ticket) error {
	return s.syncer.RevokeTicket(ctx, ticket)
}

func (s *Service) DeleteExpiredTicket(ctx context.Context, ticket Ticket) error {
	return s.syncer.DeleteExpiredTicket(ctx, ticket)
}

// NoopSyncer is the current adapter: it records the intended action but never
// opens or writes to radius_db. The real implementation will target FreeRADIUS
// tables such as radcheck/radreply/radusergroup.
type NoopSyncer struct{}

func (NoopSyncer) ProvisionTicket(ctx context.Context, ticket Ticket) error {
	slog.DebugContext(ctx, "radius sync provision skipped", "ticket_id", ticket.ID, "username", ticket.Username)
	return nil
}

func (NoopSyncer) RevokeTicket(ctx context.Context, ticket Ticket) error {
	slog.DebugContext(ctx, "radius sync revoke skipped", "ticket_id", ticket.ID, "username", ticket.Username)
	return nil
}

func (NoopSyncer) DeleteExpiredTicket(ctx context.Context, ticket Ticket) error {
	slog.DebugContext(ctx, "radius sync delete expired skipped", "ticket_id", ticket.ID, "username", ticket.Username)
	return nil
}
