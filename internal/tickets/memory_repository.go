package tickets

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

var ErrTicketNotFound = errors.New("ticket not found")

type MemoryRepository struct {
	mu      sync.RWMutex
	tickets map[string]Ticket
	now     func() time.Time
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tickets: make(map[string]Ticket),
		now:     time.Now,
	}
}

func (r *MemoryRepository) Create(ctx context.Context, input TicketCreateInput) (Ticket, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, err := newUUID()
	if err != nil {
		return Ticket{}, err
	}

	ticket := Ticket{
		ID:                id,
		Username:          input.Username,
		CleartextPassword: input.CleartextPassword,
		PitchID:           input.PitchID,
		Status:            TicketStatusActive,
		ValidFrom:         input.ValidFrom,
		ValidUntil:        input.ValidUntil,
		CreatedBy:         input.CreatedBy,
		CreatedAt:         r.now(),
	}
	r.tickets[ticket.ID] = ticket

	return ticket, nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ticket, ok := r.tickets[id]
	if !ok {
		return Ticket{}, ErrTicketNotFound
	}

	return ticket, nil
}

func (r *MemoryRepository) ListActive(ctx context.Context, now time.Time) ([]Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tickets := make([]Ticket, 0)
	for _, ticket := range r.tickets {
		if ticket.Status == TicketStatusActive && !ticket.ValidFrom.After(now) && ticket.ValidUntil.After(now) {
			tickets = append(tickets, ticket)
		}
	}
	sortTickets(tickets)

	return tickets, nil
}

func (r *MemoryRepository) ListAll(ctx context.Context) ([]Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tickets := make([]Ticket, 0, len(r.tickets))
	for _, ticket := range r.tickets {
		tickets = append(tickets, ticket)
	}
	sortTickets(tickets)

	return tickets, nil
}

func (r *MemoryRepository) Revoke(ctx context.Context, input TicketRevokeInput) (Ticket, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ticket, ok := r.tickets[input.ID]
	if !ok {
		return Ticket{}, ErrTicketNotFound
	}

	ticket.Status = TicketStatusRevoked
	ticket.RevokedAt = &input.RevokedAt
	ticket.RevokedBy = &input.RevokedBy
	r.tickets[ticket.ID] = ticket

	return ticket, nil
}

func (r *MemoryRepository) MarkExpired(ctx context.Context, now time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for id, ticket := range r.tickets {
		if ticket.Status == TicketStatusActive && !ticket.ValidUntil.After(now) {
			ticket.Status = TicketStatusExpired
			r.tickets[id] = ticket
			count++
		}
	}

	return count, nil
}

func (r *MemoryRepository) DeleteOldExpired(ctx context.Context, before time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for id, ticket := range r.tickets {
		if ticket.Status == TicketStatusExpired && ticket.ValidUntil.Before(before) {
			delete(r.tickets, id)
			count++
		}
	}

	return count, nil
}

func sortTickets(tickets []Ticket) {
	sort.Slice(tickets, func(i, j int) bool {
		return tickets[i].CreatedAt.Before(tickets[j].CreatedAt)
	})
}

func newUUID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}

	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16],
	), nil
}
