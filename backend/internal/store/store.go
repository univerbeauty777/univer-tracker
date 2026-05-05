// Package store wraps the Postgres connection pool and exposes typed
// repositories for the application's persistent state. We deliberately
// keep SQL inline (no ORM, no sqlc yet) — at this stage the queries are
// few enough that readability beats abstraction.
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open creates a Postgres connection pool from a libpq URL. The pool is
// tuned for a small two-container deploy (api + worker, 512MB each)
// behind a managed Postgres: connections recycle hourly so transient
// network drops don't pin a dead conn, idle conns get reaped to free DB
// memory, and a healthcheck weeds out broken sockets before a query
// notices.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)
	}
	cfg.MaxConns = 25
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}
