// Package sync orchestrates incremental fetches from external systems
// (WooCommerce, Frenet) into the local Postgres store.
package sync

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/integrations"
	"github.com/univerbeauty777/univer-tracker/backend/internal/orders"
	"github.com/univerbeauty777/univer-tracker/backend/internal/sla"
	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
	"github.com/univerbeauty777/univer-tracker/backend/internal/woocommerce"
)

// WooCommerce pulls orders incrementally and writes them (plus their
// shipments) into the store. It resolves credentials per-Run so a
// settings change picks up on the next tick — no restart needed.
type WooCommerce struct {
	Store        *store.Orders
	Shipments    *store.Shipments
	State        *store.SyncStates
	Integrations *integrations.Resolver
	StoreID      int64
	Log          *slog.Logger
}

const wcSyncEntity = "wc_orders"

// Run pulls every order modified since the last sync (or 30 days back on
// first run) and upserts it. Returns counts useful for logging.
func (s *WooCommerce) Run(ctx context.Context) (Stats, error) {
	stats := Stats{Started: time.Now()}

	wc, err := s.Integrations.WooCommerce(ctx)
	if err != nil {
		s.Log.Warn("wc sync skipped: not configured", "err", err)
		stats.Finished = time.Now()
		return stats, nil
	}

	since := time.Now().Add(-30 * 24 * time.Hour)
	if state, err := s.State.Get(ctx, wcSyncEntity); err == nil && state.LastSyncedAt != nil {
		// Re-fetch a small overlap window to recover from any race during
		// the previous run.
		since = state.LastSyncedAt.Add(-15 * time.Minute)
	}

	page := 1
	for {
		batch, err := wc.ListOrders(ctx, woocommerce.ListOrdersParams{
			Modified: since,
			PerPage:  100,
			Page:     page,
		})
		if err != nil {
			return stats, fmt.Errorf("list orders page %d: %w", page, err)
		}
		if len(batch) == 0 {
			break
		}
		for i := range batch {
			if err := s.persist(ctx, &batch[i]); err != nil {
				stats.Errors++
				s.Log.Error("persist order failed", "wc_order_id", batch[i].ID, "err", err)
				continue
			}
			stats.Synced++
		}
		if len(batch) < 100 {
			break
		}
		page++
	}

	if err := s.State.Mark(ctx, wcSyncEntity, stats.Started); err != nil {
		s.Log.Warn("mark sync state failed", "err", err)
	}
	stats.Finished = time.Now()
	return stats, nil
}

func (s *WooCommerce) persist(ctx context.Context, w *woocommerce.Order) error {
	dbOrder := mapOrder(w, s.StoreID)
	if _, err := s.Store.Upsert(ctx, dbOrder); err != nil {
		return fmt.Errorf("upsert order: %w", err)
	}

	tracking := orders.FromOrder(w)
	if !tracking.HasTracking() {
		return nil
	}
	ship := &store.Shipment{
		OrderID:      dbOrder.ID,
		TrackingCode: strings.ToUpper(strings.ReplaceAll(tracking.Number, " ", "")),
		Carrier:      tracking.Carrier,
		ServiceCode:  tracking.ServiceCode,
		Status:       "created",
		Health:       "unknown",
		CreatedAt:    dbOrder.CreatedAt,
	}
	sla.Apply(ship, sla.Compute(ship, time.Now().UTC()))
	if _, err := s.Shipments.Upsert(ctx, ship); err != nil {
		return fmt.Errorf("upsert shipment: %w", err)
	}
	return nil
}

// mapOrder projects WooCommerce → store.Order, preferring shipping address
// fields and falling back to billing.
func mapOrder(w *woocommerce.Order, storeID int64) *store.Order {
	name := strings.TrimSpace(w.Shipping.FirstName + " " + w.Shipping.LastName)
	city := w.Shipping.City
	uf := w.Shipping.State
	if name == "" {
		name = strings.TrimSpace(w.Billing.FirstName + " " + w.Billing.LastName)
	}
	if city == "" {
		city = w.Billing.City
		uf = w.Billing.State
	}
	method := ""
	if len(w.ShippingLines) > 0 {
		method = w.ShippingLines[0].MethodTitle
	}

	total, _ := strconv.ParseFloat(w.Total, 64)

	o := &store.Order{
		StoreID:        storeID,
		WCOrderID:      w.ID,
		Status:         w.Status,
		CustomerName:   name,
		CustomerEmail:  w.Billing.Email,
		CustomerPhone:  w.Billing.Phone,
		CustomerCity:   city,
		CustomerUF:     uf,
		ShippingMethod: method,
		TotalBRL:       total,
		CreatedAt:      w.DateCreatedGMT.Time,
	}
	if !w.DatePaidGMT.Time.IsZero() {
		t := w.DatePaidGMT.Time
		o.PaidAt = &t
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC()
	}
	return o
}

// Stats reports the work done by a single sync pass.
type Stats struct {
	Started  time.Time
	Finished time.Time
	Synced   int
	Errors   int
}
