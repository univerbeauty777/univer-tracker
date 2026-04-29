package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/univerbeauty777/univer-tracker/backend/internal/frenet"
	"github.com/univerbeauty777/univer-tracker/backend/internal/sla"
	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// BackfillStages walks every shipment, replays the tracking_events table
// to populate per-stage timestamps and re-evaluates sla_state. Safe to
// re-run — it keeps the earliest timestamp per stage and only writes
// when something changes.
type BackfillStages struct {
	Pool      *pgxpool.Pool
	Shipments *store.Shipments
	Log       *slog.Logger
}

// Run replays events for every shipment with a tracking_code. It first
// nulls out the stage stamps so any earlier mapping bug (e.g. "Etiqueta
// emitida - Aguardando postagem" landing on preparing_at) gets fully
// repaired from the events table on every deploy. delivered_at is left
// alone — that's authoritative from WC's "completed" status, not events.
func (b *BackfillStages) Run(ctx context.Context) (int, error) {
	if _, err := b.Pool.Exec(ctx, `
UPDATE shipments SET
    label_issued_at = NULL,
    preparing_at = NULL,
    ready_for_pickup_at = NULL,
    posted_at = NULL,
    in_transit_at = NULL,
    at_destination_city_at = NULL,
    out_for_delivery_at = NULL
WHERE tracking_code <> ''`); err != nil {
		return 0, fmt.Errorf("reset stage stamps: %w", err)
	}

	rows, err := b.Pool.Query(ctx, `
SELECT id FROM shipments WHERE tracking_code <> '' ORDER BY id`)
	if err != nil {
		return 0, fmt.Errorf("list shipments: %w", err)
	}
	defer rows.Close()

	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	updated := 0
	for _, id := range ids {
		ship, err := b.Shipments.GetByID(ctx, id)
		if err != nil {
			b.Log.Warn("backfill: get shipment", "id", id, "err", err)
			continue
		}
		if b.replay(ctx, ship) {
			if _, err := b.Shipments.Upsert(ctx, ship); err != nil {
				b.Log.Warn("backfill: upsert", "id", id, "err", err)
				continue
			}
			updated++
		}
	}
	return updated, nil
}

// replay applies every event row to the shipment and recomputes SLA.
// Returns true when something changed (caller decides whether to upsert).
func (b *BackfillStages) replay(ctx context.Context, ship *store.Shipment) bool {
	const eventsQ = `
SELECT occurred_at, description
FROM tracking_events
WHERE shipment_id = $1
ORDER BY occurred_at ASC`
	rows, err := b.Pool.Query(ctx, eventsQ, ship.ID)
	if err != nil {
		return false
	}
	defer rows.Close()

	changed := false
	for rows.Next() {
		var occ time.Time
		var descr string
		if err := rows.Scan(&occ, &descr); err != nil {
			continue
		}
		for _, stage := range frenet.MapEventToStages(descr) {
			if applyIfEarlier(ship, stage, occ) {
				changed = true
			}
		}
	}

	now := time.Now().UTC()
	anchor := ship.CreatedAt
	if anchor.IsZero() {
		anchor = now
	}
	eval := sla.Evaluate(ship, anchor, now)
	if string(eval.State) != ship.SLAState || eval.BreachedStage != ship.SLABreachedStage {
		ship.SLAState = string(eval.State)
		ship.SLABreachedStage = eval.BreachedStage
		changed = true
	}
	if ship.EstimatedDelivery == nil && !eval.EstimatedAt.IsZero() {
		t := eval.EstimatedAt
		ship.EstimatedDelivery = &t
		changed = true
	}
	return changed
}

func applyIfEarlier(ship *store.Shipment, stage string, t time.Time) bool {
	earlier := func(cur **time.Time, next time.Time) bool {
		if *cur != nil && !((*cur).IsZero()) && !next.Before(**cur) {
			return false
		}
		tt := next
		*cur = &tt
		return true
	}
	switch stage {
	case "label_issued_at":
		return earlier(&ship.LabelIssuedAt, t)
	case "preparing_at":
		return earlier(&ship.PreparingAt, t)
	case "ready_for_pickup_at":
		return earlier(&ship.ReadyForPickupAt, t)
	case "posted_at":
		return earlier(&ship.PostedAt, t)
	case "in_transit_at":
		return earlier(&ship.InTransitAt, t)
	case "at_destination_city_at":
		return earlier(&ship.AtDestinationCityAt, t)
	case "out_for_delivery_at":
		return earlier(&ship.OutForDeliveryAt, t)
	case "delivered_at":
		return earlier(&ship.DeliveredAt, t)
	}
	return false
}
