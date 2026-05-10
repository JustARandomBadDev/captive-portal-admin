package pitches

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

var ErrPitchNotFound = errors.New("pitch not found")

type MemoryRepository struct {
	mu      sync.RWMutex
	pitches map[string]Pitch
	now     func() time.Time
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		pitches: make(map[string]Pitch),
		now:     time.Now,
	}
}

func (r *MemoryRepository) Create(ctx context.Context, input PitchCreateInput) (Pitch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	id, err := newUUID()
	if err != nil {
		return Pitch{}, err
	}

	pitch := Pitch{
		ID:        id,
		Code:      input.Code,
		Label:     input.Label,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.pitches[pitch.ID] = pitch

	return pitch, nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (Pitch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pitch, ok := r.pitches[id]
	if !ok {
		return Pitch{}, ErrPitchNotFound
	}

	return pitch, nil
}

func (r *MemoryRepository) ListActive(ctx context.Context) ([]Pitch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pitches := make([]Pitch, 0)
	for _, pitch := range r.pitches {
		if pitch.IsActive {
			pitches = append(pitches, pitch)
		}
	}
	sortPitches(pitches)

	return pitches, nil
}

func (r *MemoryRepository) ListAll(ctx context.Context) ([]Pitch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pitches := make([]Pitch, 0, len(r.pitches))
	for _, pitch := range r.pitches {
		pitches = append(pitches, pitch)
	}
	sortPitches(pitches)

	return pitches, nil
}

func (r *MemoryRepository) Update(ctx context.Context, input PitchUpdateInput) (Pitch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pitch, ok := r.pitches[input.ID]
	if !ok {
		return Pitch{}, ErrPitchNotFound
	}

	pitch.Code = input.Code
	pitch.Label = input.Label
	pitch.IsActive = input.IsActive
	pitch.UpdatedAt = r.now()
	r.pitches[pitch.ID] = pitch

	return pitch, nil
}

func (r *MemoryRepository) Disable(ctx context.Context, id string) (Pitch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pitch, ok := r.pitches[id]
	if !ok {
		return Pitch{}, ErrPitchNotFound
	}

	pitch.IsActive = false
	pitch.UpdatedAt = r.now()
	r.pitches[pitch.ID] = pitch

	return pitch, nil
}

func sortPitches(pitches []Pitch) {
	sort.Slice(pitches, func(i, j int) bool {
		return pitches[i].Code < pitches[j].Code
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
