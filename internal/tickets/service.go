package tickets

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidTicketStatus = errors.New("invalid ticket status")
	ErrInvalidTicketDates  = errors.New("valid_until must be after valid_from")
	ErrPitchRequired       = errors.New("pitch_id is required")
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		now:        time.Now,
	}
}

func (s *Service) Create(ctx context.Context, input TicketCreateInput) (Ticket, error) {
	input.PitchID = strings.TrimSpace(input.PitchID)
	input.Username = strings.TrimSpace(input.Username)
	input.CleartextPassword = strings.TrimSpace(input.CleartextPassword)
	input.CreatedBy = strings.TrimSpace(input.CreatedBy)

	if input.PitchID == "" {
		return Ticket{}, ErrPitchRequired
	}
	if !input.ValidUntil.After(input.ValidFrom) {
		return Ticket{}, ErrInvalidTicketDates
	}

	generateUsername := input.Username == ""
	if input.CleartextPassword == "" {
		password, err := GeneratePassword(14)
		if err != nil {
			return Ticket{}, err
		}
		input.CleartextPassword = password
	}

	for attempt := 0; attempt < 5; attempt++ {
		if generateUsername {
			username, err := GenerateUsername(input.ValidFrom)
			if err != nil {
				return Ticket{}, err
			}
			input.Username = username
		}

		ticket, err := s.repository.Create(ctx, input)
		if err == nil {
			return ticket, nil
		}
		if !generateUsername || !errors.Is(err, ErrDuplicateUsername) {
			return Ticket{}, err
		}
	}

	return Ticket{}, ErrDuplicateUsername
}

func (s *Service) ListAll(ctx context.Context) ([]Ticket, error) {
	if _, err := s.repository.MarkExpired(ctx, s.now()); err != nil {
		return nil, err
	}
	return s.repository.ListAll(ctx)
}

func (s *Service) Revoke(ctx context.Context, input TicketRevokeInput) (Ticket, error) {
	input.ID = strings.TrimSpace(input.ID)
	input.RevokedBy = strings.TrimSpace(input.RevokedBy)
	if input.RevokedAt.IsZero() {
		input.RevokedAt = s.now()
	}

	return s.repository.Revoke(ctx, input)
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

func GenerateUsername(now time.Time) (string, error) {
	suffix, err := randomString("ABCDEFGHJKLMNPQRSTUVWXYZ23456789", 4)
	if err != nil {
		return "", err
	}

	return "cp-" + now.Format("20060102") + "-" + suffix, nil
}

func GeneratePassword(length int) (string, error) {
	return randomString("ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789", length)
}

func randomString(alphabet string, length int) (string, error) {
	result := make([]byte, length)
	random := make([]byte, length)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	for i, value := range random {
		result[i] = alphabet[int(value)%len(alphabet)]
	}

	return string(result), nil
}
