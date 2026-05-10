package database

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMissingDatabaseURL = errors.New("DATABASE_URL is required")

type Config struct {
	URL string
}

type Handle struct {
	config Config
	pool   *pgxpool.Pool
}

func Connect(ctx context.Context, config Config) (*Handle, error) {
	if config.URL == "" {
		return nil, ErrMissingDatabaseURL
	}

	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = 5
	poolConfig.MinConns = 0
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	handle := &Handle{
		config: config,
		pool:   pool,
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := handle.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return handle, nil
}

func (h *Handle) Config() Config {
	return h.config
}

func (h *Handle) Pool() *pgxpool.Pool {
	return h.pool
}

func (h *Handle) Ping(ctx context.Context) error {
	return h.pool.Ping(ctx)
}

func (h *Handle) Close() {
	h.pool.Close()
}

func NewUUID() (string, error) {
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
