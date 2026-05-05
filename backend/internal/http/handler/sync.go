package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
	syncpkg "github.com/univerbeauty777/univer-tracker/backend/internal/sync"
)

// SyncH exposes the worker's bookkeeping over HTTP.
type SyncH struct {
	State  *store.SyncStates
	WC     *syncpkg.WooCommerce
	Frenet *syncpkg.Frenet
	Log    *slog.Logger
}

type syncStatusItem struct {
	Entity       string     `json:"entity"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
	SecondsAgo   int64      `json:"seconds_ago"`
}

// Status handles GET /api/v1/sync/status.
func (h *SyncH) Status(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	now := time.Now()
	out := []syncStatusItem{}
	for _, entity := range []string{"wc_orders"} {
		s, err := h.State.Get(ctx, entity)
		item := syncStatusItem{Entity: entity}
		if err == nil && s.LastSyncedAt != nil {
			item.LastSyncedAt = s.LastSyncedAt
			item.SecondsAgo = int64(now.Sub(*s.LastSyncedAt).Seconds())
		} else {
			item.SecondsAgo = -1
		}
		out = append(out, item)
	}

	// Frenet doesn't use a single sync_state row — derive from the most
	// recently refreshed shipment.
	frenetItem := syncStatusItem{Entity: "frenet"}
	var lastFrenet *time.Time
	if err := h.State.Pool.QueryRow(ctx, `SELECT MAX(last_synced_at) FROM shipments WHERE last_synced_at IS NOT NULL`).Scan(&lastFrenet); err == nil && lastFrenet != nil {
		frenetItem.LastSyncedAt = lastFrenet
		frenetItem.SecondsAgo = int64(now.Sub(*lastFrenet).Seconds())
	} else {
		frenetItem.SecondsAgo = -1
	}
	out = append(out, frenetItem)

	writeJSON(w, http.StatusOK, map[string]any{"sources": out})
}

// Trigger handles POST /api/v1/sync/run — kicks both syncs in the background.
func (h *SyncH) Trigger(w http.ResponseWriter, r *http.Request) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if h.WC != nil {
			if _, err := h.WC.Run(ctx); err != nil {
				h.Log.Error("manual wc sync failed", "err", err)
			}
		}
		if h.Frenet != nil {
			if _, err := h.Frenet.Run(ctx); err != nil {
				h.Log.Error("manual frenet sync failed", "err", err)
			}
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "queued"})
}
