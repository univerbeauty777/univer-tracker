package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("not found")

// Orders persists WooCommerce orders.
type Orders struct {
	Pool *pgxpool.Pool
}

// Upsert inserts or updates an Order keyed on (store_id, wc_order_id).
// Returns the persisted row id.
func (r *Orders) Upsert(ctx context.Context, o *Order) (int64, error) {
	tagsJSON, _ := json.Marshal(o.Tags)
	if len(tagsJSON) == 0 {
		tagsJSON = []byte("[]")
	}
	const q = `
INSERT INTO orders (
    store_id, wc_order_id, status,
    customer_name, customer_email, customer_phone, customer_city, customer_uf,
    shipping_method, total_brl, declared_value, tags,
    paid_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb, $13, COALESCE($14, NOW()), NOW()
)
ON CONFLICT (store_id, wc_order_id) DO UPDATE SET
    status          = EXCLUDED.status,
    customer_name   = EXCLUDED.customer_name,
    customer_email  = EXCLUDED.customer_email,
    customer_phone  = EXCLUDED.customer_phone,
    customer_city   = EXCLUDED.customer_city,
    customer_uf     = EXCLUDED.customer_uf,
    shipping_method = EXCLUDED.shipping_method,
    total_brl       = EXCLUDED.total_brl,
    declared_value  = EXCLUDED.declared_value,
    tags            = EXCLUDED.tags,
    paid_at         = COALESCE(EXCLUDED.paid_at, orders.paid_at),
    updated_at      = NOW()
RETURNING id;`
	var id int64
	err := r.Pool.QueryRow(ctx, q,
		o.StoreID, o.WCOrderID, o.Status,
		o.CustomerName, o.CustomerEmail, o.CustomerPhone, o.CustomerCity, o.CustomerUF,
		o.ShippingMethod, o.TotalBRL, o.DeclaredValue, string(tagsJSON),
		o.PaidAt, o.CreatedAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert order: %w", err)
	}
	o.ID = id
	return id, nil
}

// ListFilters narrows the order list query.
type ListFilters struct {
	StoreID  int64
	Statuses []string
	Health   []string
	Carriers []string
	UFs      []string
	Search   string // wc_order_id, tracking_code or customer_name LIKE
	Since    *time.Time
	Until    *time.Time
	Sort     string // created_at | total | customer_name | last_event
	SortDir  string // asc | desc
	Limit    int
	Offset   int
}

// ListResult is what List returns: orders + total matching the filter
// (independent of limit/offset) so the dashboard can paginate.
type ListResult struct {
	Rows  []OrderRow
	Total int
}

// OrderRow is what the list view returns: order + (optional) primary shipment.
type OrderRow struct {
	Order
	Shipment *Shipment `json:"shipment,omitempty"`
}

// List returns orders with their primary shipment (most recently updated)
// plus the total count matching the filter (for pagination).
func (r *Orders) List(ctx context.Context, f ListFilters) (ListResult, error) {
	args := []any{}
	where := []string{"1=1"}

	if f.StoreID > 0 {
		args = append(args, f.StoreID)
		where = append(where, fmt.Sprintf("o.store_id = $%d", len(args)))
	}
	if len(f.Statuses) > 0 {
		args = append(args, f.Statuses)
		where = append(where, fmt.Sprintf("o.status = ANY($%d)", len(args)))
	}
	if len(f.Health) > 0 {
		args = append(args, f.Health)
		where = append(where, fmt.Sprintf("s.health = ANY($%d)", len(args)))
	}
	if len(f.Carriers) > 0 {
		args = append(args, f.Carriers)
		where = append(where, fmt.Sprintf("s.carrier = ANY($%d)", len(args)))
	}
	if len(f.UFs) > 0 {
		args = append(args, f.UFs)
		where = append(where, fmt.Sprintf("o.customer_uf = ANY($%d)", len(args)))
	}
	if f.Since != nil {
		args = append(args, *f.Since)
		where = append(where, fmt.Sprintf("o.created_at >= $%d", len(args)))
	}
	if f.Until != nil {
		args = append(args, *f.Until)
		where = append(where, fmt.Sprintf("o.created_at < $%d", len(args)))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		idx := len(args)
		where = append(where, fmt.Sprintf("(o.customer_name ILIKE $%[1]d OR s.tracking_code ILIKE $%[1]d OR CAST(o.wc_order_id AS TEXT) ILIKE $%[1]d)", idx))
	}

	orderBy := "o.created_at DESC"
	dir := "DESC"
	if strings.EqualFold(f.SortDir, "asc") {
		dir = "ASC"
	}
	switch strings.ToLower(f.Sort) {
	case "total":
		orderBy = "o.total_brl " + dir
	case "customer_name":
		orderBy = "o.customer_name " + dir
	case "last_event":
		orderBy = "s.last_event_at " + dir + " NULLS LAST"
	case "created_at", "":
		orderBy = "o.created_at " + dir
	}

	whereClause := joinAnd(where)

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	// Total count first (without limit/offset).
	countSQL := fmt.Sprintf(`
SELECT COUNT(*)
FROM orders o
LEFT JOIN LATERAL (
    SELECT * FROM shipments sh
    WHERE sh.order_id = o.id
    ORDER BY sh.updated_at DESC
    LIMIT 1
) s ON true
WHERE %s`, whereClause)

	var total int
	if err := r.Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return ListResult{}, fmt.Errorf("count orders: %w", err)
	}

	pagedArgs := append([]any{}, args...)
	pagedArgs = append(pagedArgs, limit, f.Offset)

	listSQL := fmt.Sprintf(`
SELECT
    o.id, o.store_id, o.wc_order_id, o.status,
    o.customer_name, o.customer_email, o.customer_phone, o.customer_city, o.customer_uf,
    o.shipping_method, o.total_brl, o.declared_value, o.tags,
    o.paid_at, o.created_at, o.updated_at,
    s.id, s.tracking_code, s.carrier, s.service, s.service_code, s.tracking_url,
    s.status, s.last_event, s.last_event_at, s.estimated_delivery,
    s.label_issued_at, s.preparing_at, s.ready_for_pickup_at, s.posted_at,
    s.in_transit_at, s.at_destination_city_at, s.out_for_delivery_at, s.delivered_at,
    s.last_synced_at, s.health, s.sla_state, s.sla_breached_stage,
    s.idle_since, s.risk_score,
    s.created_at, s.updated_at
FROM orders o
LEFT JOIN LATERAL (
    SELECT * FROM shipments sh
    WHERE sh.order_id = o.id
    ORDER BY sh.updated_at DESC
    LIMIT 1
) s ON true
WHERE %s
ORDER BY %s
LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, len(pagedArgs)-1, len(pagedArgs))

	rows, err := r.Pool.Query(ctx, listSQL, pagedArgs...)
	if err != nil {
		return ListResult{}, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var out []OrderRow
	for rows.Next() {
		var o Order
		var s Shipment
		var tagsRaw []byte
		var (
			sID                                                                                 *int64
			sTracking, sCarrier, sService, sServiceCode, sURL, sStatus, sLast, sBreachedStage *string
			sLastAt, sETA                                                                     *time.Time
			sLabel, sPrep, sReady, sPosted, sInTransit, sAtCity, sOFD, sDelivered             *time.Time
			sLastSync, sIdle                                                                  *time.Time
			sHealth, sSLAState                                                                *string
			sRisk                                                                             *int16
			sCreated, sUpdated                                                                *time.Time
		)
		if err := rows.Scan(
			&o.ID, &o.StoreID, &o.WCOrderID, &o.Status,
			&o.CustomerName, &o.CustomerEmail, &o.CustomerPhone, &o.CustomerCity, &o.CustomerUF,
			&o.ShippingMethod, &o.TotalBRL, &o.DeclaredValue, &tagsRaw,
			&o.PaidAt, &o.CreatedAt, &o.UpdatedAt,
			&sID, &sTracking, &sCarrier, &sService, &sServiceCode, &sURL,
			&sStatus, &sLast, &sLastAt, &sETA,
			&sLabel, &sPrep, &sReady, &sPosted, &sInTransit, &sAtCity, &sOFD, &sDelivered,
			&sLastSync, &sHealth, &sSLAState, &sBreachedStage,
			&sIdle, &sRisk,
			&sCreated, &sUpdated,
		); err != nil {
			return ListResult{}, fmt.Errorf("scan order row: %w", err)
		}

		if len(tagsRaw) > 0 {
			_ = json.Unmarshal(tagsRaw, &o.Tags)
		}

		row := OrderRow{Order: o}
		if sID != nil {
			s.ID = *sID
			s.OrderID = o.ID
			s.TrackingCode = derefStr(sTracking)
			s.Carrier = derefStr(sCarrier)
			s.Service = derefStr(sService)
			s.ServiceCode = derefStr(sServiceCode)
			s.TrackingURL = derefStr(sURL)
			s.Status = derefStr(sStatus)
			s.LastEvent = derefStr(sLast)
			s.LastEventAt = sLastAt
			s.EstimatedDelivery = sETA
			s.LabelIssuedAt = sLabel
			s.PreparingAt = sPrep
			s.ReadyForPickupAt = sReady
			s.PostedAt = sPosted
			s.InTransitAt = sInTransit
			s.AtDestinationCityAt = sAtCity
			s.OutForDeliveryAt = sOFD
			s.DeliveredAt = sDelivered
			s.LastSyncedAt = sLastSync
			s.Health = derefStr(sHealth)
			s.SLAState = derefStr(sSLAState)
			s.SLABreachedStage = derefStr(sBreachedStage)
			s.IdleSince = sIdle
			if sRisk != nil {
				s.RiskScore = *sRisk
			}
			if sCreated != nil {
				s.CreatedAt = *sCreated
			}
			if sUpdated != nil {
				s.UpdatedAt = *sUpdated
			}
			row.Shipment = &s
		}
		out = append(out, row)
	}
	return ListResult{Rows: out, Total: total}, rows.Err()
}

// GetByWCID fetches a single order by its WooCommerce id.
func (r *Orders) GetByWCID(ctx context.Context, storeID, wcOrderID int64) (*Order, error) {
	const q = `
SELECT id, store_id, wc_order_id, status,
       customer_name, customer_email, customer_phone, customer_city, customer_uf,
       shipping_method, total_brl, declared_value, tags,
       paid_at, created_at, updated_at
FROM orders
WHERE store_id = $1 AND wc_order_id = $2`

	var o Order
	var tagsRaw []byte
	err := r.Pool.QueryRow(ctx, q, storeID, wcOrderID).Scan(
		&o.ID, &o.StoreID, &o.WCOrderID, &o.Status,
		&o.CustomerName, &o.CustomerEmail, &o.CustomerPhone, &o.CustomerCity, &o.CustomerUF,
		&o.ShippingMethod, &o.TotalBRL, &o.DeclaredValue, &tagsRaw,
		&o.PaidAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if len(tagsRaw) > 0 {
		_ = json.Unmarshal(tagsRaw, &o.Tags)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return &o, nil
}

// UpdateStatus persists a new WC status (mirror of the source of truth).
func (r *Orders) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.Pool.Exec(ctx, `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

func joinAnd(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += " AND "
		}
		out += p
	}
	return out
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
