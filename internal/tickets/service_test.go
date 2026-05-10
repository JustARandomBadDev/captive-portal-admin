package tickets

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIsValidStatus(t *testing.T) {
	validStatuses := []TicketStatus{
		TicketStatusActive,
		TicketStatusExpired,
		TicketStatusRevoked,
	}

	for _, status := range validStatuses {
		if !IsValidStatus(status) {
			t.Fatalf("expected %q to be valid", status)
		}
	}

	if IsValidStatus(TicketStatus("unknown")) {
		t.Fatal("expected unknown status to be invalid")
	}
}

func TestCreateTicketWithCoherentDates(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	validFrom := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	validUntil := validFrom.Add(24 * time.Hour)

	ticket, err := service.Create(context.Background(), TicketCreateInput{
		Username:          "ticket-001",
		CleartextPassword: "temporary-password",
		PitchID:           "pitch-001",
		ValidFrom:         validFrom,
		ValidUntil:        validUntil,
		CreatedBy:         "admin",
	})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	if ticket.Status != TicketStatusActive {
		t.Fatalf("expected active ticket, got %q", ticket.Status)
	}
	if ticket.ValidUntil.Before(ticket.ValidFrom) {
		t.Fatal("expected valid_until to be after valid_from")
	}
}

func TestCreateTicketRejectsInvalidDates(t *testing.T) {
	service := NewService(&fakeRepository{})
	validFrom := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)

	_, err := service.Create(context.Background(), TicketCreateInput{
		Username:          "ticket-001",
		CleartextPassword: "temporary-password",
		PitchID:           "pitch-001",
		ValidFrom:         validFrom,
		ValidUntil:        validFrom.Add(-time.Hour),
		CreatedBy:         "admin",
	})
	if !errors.Is(err, ErrInvalidTicketDates) {
		t.Fatalf("expected ErrInvalidTicketDates, got %v", err)
	}
}

func TestCreateTicketGeneratesCredentials(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	validFrom := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)

	ticket, err := service.Create(context.Background(), TicketCreateInput{
		PitchID:    "pitch-001",
		ValidFrom:  validFrom,
		ValidUntil: validFrom.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	if ticket.Username == "" {
		t.Fatal("expected generated username")
	}
	if ticket.CleartextPassword == "" {
		t.Fatal("expected generated password")
	}
	if len(ticket.CleartextPassword) != 14 {
		t.Fatalf("expected 14-char password, got %d", len(ticket.CleartextPassword))
	}
}

func TestCreateTicketRequiresPitch(t *testing.T) {
	service := NewService(&fakeRepository{})
	validFrom := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)

	_, err := service.Create(context.Background(), TicketCreateInput{
		ValidFrom:  validFrom,
		ValidUntil: validFrom.Add(24 * time.Hour),
	})
	if !errors.Is(err, ErrPitchRequired) {
		t.Fatalf("expected ErrPitchRequired, got %v", err)
	}
}

func TestListAllMarksExpired(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	_, err := service.ListAll(context.Background())
	if err != nil {
		t.Fatalf("list tickets: %v", err)
	}
	if !repository.markExpiredAt.Equal(now) {
		t.Fatalf("expected MarkExpired at %v, got %v", now, repository.markExpiredAt)
	}
}

func TestRevokeTicket(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	ticket, err := service.Revoke(context.Background(), TicketRevokeInput{ID: "ticket-id"})
	if err != nil {
		t.Fatalf("revoke ticket: %v", err)
	}
	if ticket.Status != TicketStatusRevoked {
		t.Fatalf("expected revoked ticket, got %q", ticket.Status)
	}
	if repository.revoked.ID != "ticket-id" {
		t.Fatalf("expected revoked id ticket-id, got %q", repository.revoked.ID)
	}
	if !repository.revoked.RevokedAt.Equal(now) {
		t.Fatalf("expected revoked_at %v, got %v", now, repository.revoked.RevokedAt)
	}
}

func TestGenerateUsername(t *testing.T) {
	username, err := GenerateUsername(time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("generate username: %v", err)
	}
	if len(username) != len("cp-20260510-AB12") {
		t.Fatalf("unexpected username length: %q", username)
	}
	if username[:12] != "cp-20260510-" {
		t.Fatalf("unexpected username prefix: %q", username)
	}
}

func TestGeneratePassword(t *testing.T) {
	password, err := GeneratePassword(14)
	if err != nil {
		t.Fatalf("generate password: %v", err)
	}
	if len(password) != 14 {
		t.Fatalf("expected 14 chars, got %d", len(password))
	}
}

type fakeRepository struct {
	created       TicketCreateInput
	revoked       TicketRevokeInput
	markExpiredAt time.Time
}

func (r *fakeRepository) Create(ctx context.Context, input TicketCreateInput) (Ticket, error) {
	r.created = input
	return Ticket{
		ID:                "ticket-id",
		Username:          input.Username,
		CleartextPassword: input.CleartextPassword,
		PitchID:           input.PitchID,
		Status:            TicketStatusActive,
		ValidFrom:         input.ValidFrom,
		ValidUntil:        input.ValidUntil,
		CreatedBy:         input.CreatedBy,
		CreatedAt:         input.ValidFrom,
	}, nil
}

func (r *fakeRepository) GetByID(ctx context.Context, id string) (Ticket, error) {
	return Ticket{}, nil
}

func (r *fakeRepository) ListActive(ctx context.Context, now time.Time) ([]Ticket, error) {
	return nil, nil
}

func (r *fakeRepository) ListAll(ctx context.Context) ([]Ticket, error) {
	return nil, nil
}

func (r *fakeRepository) Revoke(ctx context.Context, input TicketRevokeInput) (Ticket, error) {
	r.revoked = input
	return Ticket{ID: input.ID, Status: TicketStatusRevoked, RevokedAt: &input.RevokedAt}, nil
}

func (r *fakeRepository) MarkExpired(ctx context.Context, now time.Time) (int, error) {
	r.markExpiredAt = now
	return 0, nil
}

func (r *fakeRepository) DeleteOldExpired(ctx context.Context, before time.Time) (int, error) {
	return 0, nil
}
