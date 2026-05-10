package pitches

import (
	"context"
	"errors"
	"testing"
)

func TestCreatePitchRequiresCode(t *testing.T) {
	service := NewService(NewMemoryRepository())

	_, err := service.Create(context.Background(), PitchCreateInput{
		Code:  " ",
		Label: "Emplacement sans code",
	})
	if !errors.Is(err, ErrPitchCodeRequired) {
		t.Fatalf("expected ErrPitchCodeRequired, got %v", err)
	}
}

func TestCreatePitch(t *testing.T) {
	service := NewService(NewMemoryRepository())

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
