package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Events persists per-shipment carrier events.
type Events struct {
	Pool *pgxpool.Pool
}

// InsertMany adds events to a shipment, ignoring duplicates by the natural
// key (shipment_id, occurred_at, description). Returns rows actually inserted.
func (r *Events) InsertMany(ctx context.Context, evts []Event) (int, error) {
	if len(evts) == 0 {
		return 0, nil
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
INSERT INTO tracking_events (shipment_id, occurred_at, type, description, location, raw)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (shipment_id, occurred_at, description) DO NOTHING`

	count := 0
	for _, e := range evts {
		ct, err := tx.Exec(ctx, q, e.ShipmentID, e.OccurredAt, e.Type, e.Description, e.Location, e.Raw)
		if err != nil {
			return count, fmt.Errorf("insert event: %w", err)
		}
		count += int(ct.RowsAffected())
	}

	if err := tx.Commit(ctx); err != nil {
		return count, fmt.Errorf("commit tx: %w", err)
	}
	return count, nil
}

// ListByShipment returns the carrier timeline newest first.
func (r *Events) ListByShipment(ctx context.Context, shipmentID int64) ([]Event, error) {
	const q = `
SELECT id, shipment_id, occurred_at, type, description, location, raw, created_at
FROM tracking_events
WHERE shipment_id = $1
ORDER BY occurred_at DESC`
	rows, err := r.Pool.Query(ctx, q, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.ShipmentID, &e.OccurredAt, &e.Type, &e.Description, &e.Location, &e.Raw, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

var _ = pgxpool.Pool{}
