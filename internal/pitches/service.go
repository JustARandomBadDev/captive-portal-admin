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
	if strings.TrimSpace(input.Code) == "" {
		return Pitch{}, ErrPitchCodeRequired
	}

	return s.repository.Create(ctx, input)
}

func (s *Service) Repository() Repository {
	return s.repository
}
