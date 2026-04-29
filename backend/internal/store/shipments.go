package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Shipments persists carrier-side shipment state.
type Shipments struct {
	Pool *pgxpool.Pool
}

// Upsert creates or updates a Shipment keyed on (order_id, tracking_code).
// COALESCE on stamps preserves earliest known value — the Frenet API
// returns events newest-first, but it sometimes drops older ones, and
// we don't want a re-sync to forget that an earlier stage happened.
func (r *Shipments) Upsert(ctx context.Context, s *Shipment) (int64, error) {
	const q = `
INSERT INTO shipments (
    order_id, tracking_code, carrier, service, service_code, tracking_url,
    status, last_event, last_event_at, estimated_delivery,
    label_issued_at, preparing_at, ready_for_pickup_at, posted_at,
    in_transit_at, at_destination_city_at, out_for_delivery_at, delivered_at,
    last_synced_at, health, sla_state, sla_breached_stage,
    idle_since, risk_score, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    COALESCE($7, 'created'), $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18,
    $19, COALESCE($20, 'unknown'), COALESCE($21, 'ON_TRACK'), $22,
    $23, COALESCE($24, 0), NOW(), NOW()
)
ON CONFLICT (order_id, tracking_code) DO UPDATE SET
    carrier               = EXCLUDED.carrier,
    service               = EXCLUDED.service,
    service_code          = EXCLUDED.service_code,
    tracking_url          = EXCLUDED.tracking_url,
    status                = EXCLUDED.status,
    last_event            = EXCLUDED.last_event,
    last_event_at         = EXCLUDED.last_event_at,
    estimated_delivery    = EXCLUDED.estimated_delivery,
    label_issued_at        = COALESCE(shipments.label_issued_at, EXCLUDED.label_issued_at),
    preparing_at           = COALESCE(shipments.preparing_at, EXCLUDED.preparing_at),
    ready_for_pickup_at    = COALESCE(shipments.ready_for_pickup_at, EXCLUDED.ready_for_pickup_at),
    posted_at              = COALESCE(shipments.posted_at, EXCLUDED.posted_at),
    in_transit_at          = COALESCE(shipments.in_transit_at, EXCLUDED.in_transit_at),
    at_destination_city_at = COALESCE(shipments.at_destination_city_at, EXCLUDED.at_destination_city_at),
    out_for_delivery_at    = COALESCE(shipments.out_for_delivery_at, EXCLUDED.out_for_delivery_at),
    delivered_at           = COALESCE(shipments.delivered_at, EXCLUDED.delivered_at),
    last_synced_at        = EXCLUDED.last_synced_at,
    health                = EXCLUDED.health,
    sla_state             = EXCLUDED.sla_state,
    sla_breached_stage    = EXCLUDED.sla_breached_stage,
    idle_since            = EXCLUDED.idle_since,
    risk_score            = EXCLUDED.risk_score,
    updated_at            = NOW()
RETURNING id;`
	var id int64
	err := r.Pool.QueryRow(ctx, q,
		s.OrderID, s.TrackingCode, s.Carrier, s.Service, s.ServiceCode, s.TrackingURL,
		s.Status, s.LastEvent, s.LastEventAt, s.EstimatedDelivery,
		s.LabelIssuedAt, s.PreparingAt, s.ReadyForPickupAt, s.PostedAt,
		s.InTransitAt, s.AtDestinationCityAt, s.OutForDeliveryAt, s.DeliveredAt,
		s.LastSyncedAt, s.Health, s.SLAState, s.SLABreachedStage,
		s.IdleSince, s.RiskScore,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert shipment: %w", err)
	}
	s.ID = id
	return id, nil
}

// GetByID returns the shipment with the given id.
func (r *Shipments) GetByID(ctx context.Context, id int64) (*Shipment, error) {
	const q = `
SELECT id, order_id, tracking_code, carrier, service, service_code, tracking_url,
       status, last_event, last_event_at, estimated_delivery,
       label_issued_at, preparing_at, ready_for_pickup_at, posted_at,
       in_transit_at, at_destination_city_at, out_for_delivery_at, delivered_at,
       last_synced_at, health, sla_state, sla_breached_stage,
       idle_since, risk_score, created_at, updated_at
FROM shipments WHERE id = $1`
	var s Shipment
	err := r.Pool.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.OrderID, &s.TrackingCode, &s.Carrier, &s.Service, &s.ServiceCode, &s.TrackingURL,
		&s.Status, &s.LastEvent, &s.LastEventAt, &s.EstimatedDelivery,
		&s.LabelIssuedAt, &s.PreparingAt, &s.ReadyForPickupAt, &s.PostedAt,
		&s.InTransitAt, &s.AtDestinationCityAt, &s.OutForDeliveryAt, &s.DeliveredAt,
		&s.LastSyncedAt, &s.Health, &s.SLAState, &s.SLABreachedStage,
		&s.IdleSince, &s.RiskScore, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get shipment: %w", err)
	}
	return &s, nil
}

// ListByOrder returns all shipments tied to an order, newest first.
func (r *Shipments) ListByOrder(ctx context.Context, orderID int64) ([]Shipment, error) {
	const q = `
SELECT id, order_id, tracking_code, carrier, service, service_code, tracking_url,
       status, last_event, last_event_at, estimated_delivery,
       label_issued_at, preparing_at, ready_for_pickup_at, posted_at,
       in_transit_at, at_destination_city_at, out_for_delivery_at, delivered_at,
       last_synced_at, health, sla_state, sla_breached_stage,
       idle_since, risk_score, created_at, updated_at
FROM shipments WHERE order_id = $1 ORDER BY updated_at DESC`
	rows, err := r.Pool.Query(ctx, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("list shipments: %w", err)
	}
	defer rows.Close()
	var out []Shipment
	for rows.Next() {
		var s Shipment
		if err := rows.Scan(
			&s.ID, &s.OrderID, &s.TrackingCode, &s.Carrier, &s.Service, &s.ServiceCode, &s.TrackingURL,
			&s.Status, &s.LastEvent, &s.LastEventAt, &s.EstimatedDelivery,
			&s.LabelIssuedAt, &s.PreparingAt, &s.ReadyForPickupAt, &s.PostedAt,
			&s.InTransitAt, &s.AtDestinationCityAt, &s.OutForDeliveryAt, &s.DeliveredAt,
			&s.LastSyncedAt, &s.Health, &s.SLAState, &s.SLABreachedStage,
			&s.IdleSince, &s.RiskScore, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ListActive returns shipments in non-terminal states, oldest sync first
// (so the cron worker fairly rotates through everyone).
func (r *Shipments) ListActive(ctx context.Context, limit int) ([]Shipment, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	const q = `
SELECT id, order_id, tracking_code, carrier, service, service_code, tracking_url,
       status, last_event, last_event_at, estimated_delivery,
       label_issued_at, preparing_at, ready_for_pickup_at, posted_at,
       in_transit_at, at_destination_city_at, out_for_delivery_at, delivered_at,
       last_synced_at, health, sla_state, sla_breached_stage,
       idle_since, risk_score, created_at, updated_at
FROM shipments
WHERE tracking_code <> ''
  AND status NOT IN ('delivered', 'returned')
ORDER BY last_synced_at NULLS FIRST
LIMIT $1`
	rows, err := r.Pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list active shipments: %w", err)
	}
	defer rows.Close()
	var out []Shipment
	for rows.Next() {
		var s Shipment
		if err := rows.Scan(
			&s.ID, &s.OrderID, &s.TrackingCode, &s.Carrier, &s.Service, &s.ServiceCode, &s.TrackingURL,
			&s.Status, &s.LastEvent, &s.LastEventAt, &s.EstimatedDelivery,
			&s.LabelIssuedAt, &s.PreparingAt, &s.ReadyForPickupAt, &s.PostedAt,
			&s.InTransitAt, &s.AtDestinationCityAt, &s.OutForDeliveryAt, &s.DeliveredAt,
			&s.LastSyncedAt, &s.Health, &s.SLAState, &s.SLABreachedStage,
			&s.IdleSince, &s.RiskScore, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
