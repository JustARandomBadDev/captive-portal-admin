package pitches

import (
	"context"
	"errors"
	"testing"
)

func TestCreatePitchRequiresCode(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.Create(context.Background(), PitchCreateInput{
		Code:  " ",
		Label: "Emplacement sans code",
	})
	if !errors.Is(err, ErrPitchCodeRequired) {
		t.Fatalf("expected ErrPitchCodeRequired, got %v", err)
	}
}

func TestCreatePitch(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	pitch, err := service.Create(context.Background(), PitchCreateInput{
		Code:  " A12 ",
		Label: " Emplacement A12 ",
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
	if repository.created.Code != "A12" {
		t.Fatalf("expected trimmed code A12, got %q", repository.created.Code)
	}
	if repository.created.Label != "Emplacement A12" {
		t.Fatalf("expected trimmed label, got %q", repository.created.Label)
	}
}

func TestListAllPitches(t *testing.T) {
	service := NewService(&fakeRepository{
		pitches: []Pitch{{ID: "pitch-id", Code: "A12", IsActive: true}},
	})

	pitches, err := service.ListAll(context.Background())
	if err != nil {
		t.Fatalf("list pitches: %v", err)
	}
	if len(pitches) != 1 {
		t.Fatalf("expected 1 pitch, got %d", len(pitches))
	}
}

func TestDisablePitch(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	pitch, err := service.Disable(context.Background(), "pitch-id")
	if err != nil {
		t.Fatalf("disable pitch: %v", err)
	}
	if pitch.IsActive {
		t.Fatal("expected disabled pitch to be inactive")
	}
	if repository.disabledID != "pitch-id" {
		t.Fatalf("expected disabled id pitch-id, got %q", repository.disabledID)
	}
}

func TestEnablePitch(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	pitch, err := service.Enable(context.Background(), "pitch-id")
	if err != nil {
		t.Fatalf("enable pitch: %v", err)
	}
	if !pitch.IsActive {
		t.Fatal("expected enabled pitch to be active")
	}
	if repository.enabledID != "pitch-id" {
		t.Fatalf("expected enabled id pitch-id, got %q", repository.enabledID)
	}
}

type fakeRepository struct {
	created    PitchCreateInput
	disabledID string
	enabledID  string
	pitches    []Pitch
}

func (r *fakeRepository) Create(ctx context.Context, input PitchCreateInput) (Pitch, error) {
	r.created = input
	return Pitch{
		ID:       "pitch-id",
		Code:     input.Code,
		Label:    input.Label,
		IsActive: true,
	}, nil
}

func (r *fakeRepository) GetByID(ctx context.Context, id string) (Pitch, error) {
	return Pitch{}, nil
}

func (r *fakeRepository) ListActive(ctx context.Context) ([]Pitch, error) {
	return nil, nil
}

func (r *fakeRepository) ListAll(ctx context.Context) ([]Pitch, error) {
	return r.pitches, nil
}

func (r *fakeRepository) Update(ctx context.Context, input PitchUpdateInput) (Pitch, error) {
	return Pitch{}, nil
}

func (r *fakeRepository) Disable(ctx context.Context, id string) (Pitch, error) {
	r.disabledID = id
	return Pitch{ID: id, Code: "A12", IsActive: false}, nil
}

func (r *fakeRepository) Enable(ctx context.Context, id string) (Pitch, error) {
	r.enabledID = id
	return Pitch{ID: id, Code: "A12", IsActive: true}, nil
}
