package tickets

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidTicketStatus = errors.New("invalid ticket status")
	ErrInvalidTicketDates  = errors.New("valid_until must be after valid_from")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, input TicketCreateInput) (Ticket, error) {
	if !input.ValidUntil.After(input.ValidFrom) {
		return Ticket{}, ErrInvalidTicketDates
	}

	return s.repository.Create(ctx, input)
}

func (s *Service) Repository() Repository {
	return s.repository
}

func IsValidStatus(status TicketStatus) bool {
	switch status {
	case TicketStatusActive, TicketStatusExpired, TicketStatusRevoked:
		return true
	default:
		return false
	}
}

func DurationFromDates(validFrom, validUntil time.Time) (time.Duration, error) {
	if !validUntil.After(validFrom) {
		return 0, ErrInvalidTicketDates
	}
	return validUntil.Sub(validFrom), nil
}
