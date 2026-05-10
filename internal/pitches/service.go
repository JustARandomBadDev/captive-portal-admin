package pitches

import (
	"context"
	"errors"
	"strings"
)

var ErrPitchCodeRequired = errors.New("pitch code is required")

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, input PitchCreateInput) (Pitch, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Label = strings.TrimSpace(input.Label)
	if input.Code == "" {
		return Pitch{}, ErrPitchCodeRequired
	}

	return s.repository.Create(ctx, input)
}

func (s *Service) ListAll(ctx context.Context) ([]Pitch, error) {
	return s.repository.ListAll(ctx)
}

func (s *Service) ListActive(ctx context.Context) ([]Pitch, error) {
	return s.repository.ListActive(ctx)
}

func (s *Service) Disable(ctx context.Context, id string) (Pitch, error) {
	return s.repository.Disable(ctx, id)
}

func (s *Service) Enable(ctx context.Context, id string) (Pitch, error) {
	return s.repository.Enable(ctx, id)
}

func (s *Service) Repository() Repository {
	return s.repository
}
