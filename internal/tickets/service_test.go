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
	service := NewService(NewMemoryRepository())
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
	service := NewService(NewMemoryRepository())
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
