package pitches

import (
	"context"
	"errors"
	"testing"
)

func TestCreatePitchRequiresCode(t *testing.T) {
	service := NewService(fakeRepository{})

	_, err := service.Create(context.Background(), PitchCreateInput{
		Code:  " ",
		Label: "Emplacement sans code",
	})
	if !errors.Is(err, ErrPitchCodeRequired) {
		t.Fatalf("expected ErrPitchCodeRequired, got %v", err)
	}
}

func TestCreatePitch(t *testing.T) {
	service := NewService(fakeRepository{})

	pitch, err := service.Create(context.Background(), PitchCreateInput{
		Code:  "A12",
		Label: "Emplacement A12",
	})
	if err != nil {
		t.Fatalf("create pitch: %v", err)
	}

	if pitch.Code != "A12" {
		t.Fatalf("expected code A12, got %q", pitch.Code)
	}
	if !pitch.IsActive {
		t.Fatal("expected new pitch to be active")
	}
}

type fakeRepository struct{}

func (fakeRepository) Create(ctx context.Context, input PitchCreateInput) (Pitch, error) {
	return Pitch{
		ID:       "pitch-id",
		Code:     input.Code,
		Label:    input.Label,
		IsActive: true,
	}, nil
}

func (fakeRepository) GetByID(ctx context.Context, id string) (Pitch, error) {
	return Pitch{}, nil
}

func (fakeRepository) ListActive(ctx context.Context) ([]Pitch, error) {
	return nil, nil
}

func (fakeRepository) ListAll(ctx context.Context) ([]Pitch, error) {
	return nil, nil
}

func (fakeRepository) Update(ctx context.Context, input PitchUpdateInput) (Pitch, error) {
	return Pitch{}, nil
}

func (fakeRepository) Disable(ctx context.Context, id string) (Pitch, error) {
	return Pitch{}, nil
}
