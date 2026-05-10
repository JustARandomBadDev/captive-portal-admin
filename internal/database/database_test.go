package database

import (
	"context"
	"errors"
	"testing"
)

func TestConnectRequiresDatabaseURL(t *testing.T) {
	_, err := Connect(context.Background(), Config{})
	if !errors.Is(err, ErrMissingDatabaseURL) {
		t.Fatalf("expected ErrMissingDatabaseURL, got %v", err)
	}
}
