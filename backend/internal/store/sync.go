package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncStates manages the worker's "last synced at" bookmarks.
type SyncStates struct {
	Pool *pgxpool.Pool
}

// PoolHandle exposes the underlying pool when the handler needs to run a
// custom query (e.g. derive the latest Frenet sync timestamp).
func (r *SyncStates) PoolHandle() *pgxpool.Pool { return r.Pool }

// Get returns the last sync state for an entity (or ErrNotFound).
func (r *SyncStates) Get(ctx context.Context, entity string) (*SyncState, error) {
	const q = `SELECT entity, last_synced_at, extra, updated_at FROM sync_state WHERE entity = $1`
	var s SyncState
	err := r.Pool.QueryRow(ctx, q, entity).Scan(&s.Entity, &s.LastSyncedAt, &s.Extra, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get sync state: %w", err)
	}
	return &s, nil
}

// Mark stores `at` as the last successful sync for the given entity.
func (r *SyncStates) Mark(ctx context.Context, entity string, at time.Time) error {
	const q = `
INSERT INTO sync_state (entity, last_synced_at, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (entity) DO UPDATE SET
    last_synced_at = EXCLUDED.last_synced_at,
    updated_at     = NOW()`
	_, err := r.Pool.Exec(ctx, q, entity, at)
	if err != nil {
		return fmt.Errorf("mark sync state: %w", err)
	}
	return nil
}
