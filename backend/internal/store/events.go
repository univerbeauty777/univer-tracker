package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Events persists per-shipment carrier events.
type Events struct {
	Pool *pgxpool.Pool
}

// InsertMany adds events to a shipment, ignoring duplicates by the
// natural key (shipment_id, occurred_at, description). Uses pgx.Batch
// so the whole set ships in a single network round-trip and the server
// plans the INSERT once — important on Frenet sync where a single
// shipment can carry 30+ events and we used to send 30+ separate Exec
// calls inside the transaction.
func (r *Events) InsertMany(ctx context.Context, evts []Event) (int, error) {
	if len(evts) == 0 {
		return 0, nil
	}

	const q = `
INSERT INTO tracking_events (shipment_id, occurred_at, type, description, location, raw)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (shipment_id, occurred_at, description) DO NOTHING`

	batch := &pgx.Batch{}
	for _, e := range evts {
		batch.Queue(q, e.ShipmentID, e.OccurredAt, e.Type, e.Description, e.Location, e.Raw)
	}

	br := r.Pool.SendBatch(ctx, batch)
	defer br.Close()

	count := 0
	for range evts {
		ct, err := br.Exec()
		if err != nil {
			return count, fmt.Errorf("insert event: %w", err)
		}
		count += int(ct.RowsAffected())
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
