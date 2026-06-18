package http

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/tickets"
)

func TestParseTicketListFilters(t *testing.T) {
	req := httptest.NewRequest("GET", "/tickets?search=%20A01%20&status=active&duration=7d&created=today", nil)

	filters := parseTicketListFilters(req)

	if filters.Search != "A01" || filters.Filters.Search != "A01" {
		t.Fatalf("search = %q / %q, want A01", filters.Search, filters.Filters.Search)
	}
	if filters.Status != "active" || filters.Filters.Status != tickets.TicketStatusActive {
		t.Fatalf("status = %q / %q, want active", filters.Status, filters.Filters.Status)
	}
	if filters.Duration != "7d" || filters.Filters.Duration != 7*24*time.Hour {
		t.Fatalf("duration = %q / %s, want 7d", filters.Duration, filters.Filters.Duration)
	}
	if filters.Created != "today" || filters.Filters.CreatedSince == nil {
		t.Fatalf("created = %q / %v, want today with CreatedSince", filters.Created, filters.Filters.CreatedSince)
	}
}

func TestParseTicketListFiltersIgnoresUnknownValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/tickets?status=bad&duration=bad&created=bad", nil)

	filters := parseTicketListFilters(req)

	if filters.Status != "" || filters.Filters.Status != "" {
		t.Fatalf("status = %q / %q, want empty", filters.Status, filters.Filters.Status)
	}
	if filters.Duration != "" || filters.Filters.Duration != 0 {
		t.Fatalf("duration = %q / %s, want empty", filters.Duration, filters.Filters.Duration)
	}
	if filters.Created != "" || filters.Filters.CreatedSince != nil {
		t.Fatalf("created = %q / %v, want empty", filters.Created, filters.Filters.CreatedSince)
	}
}
