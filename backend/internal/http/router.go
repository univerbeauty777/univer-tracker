// Package http wires the HTTP router and registers handlers.
package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/univerbeauty777/univer-tracker/backend/internal/config"
	"github.com/univerbeauty777/univer-tracker/backend/internal/http/handler"
	"github.com/univerbeauty777/univer-tracker/backend/internal/integrations"
	"github.com/univerbeauty777/univer-tracker/backend/internal/notifier"
	"github.com/univerbeauty777/univer-tracker/backend/internal/settings"
	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
	syncpkg "github.com/univerbeauty777/univer-tracker/backend/internal/sync"
)

const defaultStoreID = int64(1)

// NewRouter creates the application's main HTTP handler.
func NewRouter(cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", readyHandler(pool))
	mux.HandleFunc("GET /api/v1/version", versionHandler)

	settingsStore := settings.New(pool)
	resolver := integrations.New(settingsStore, cfg)

	auditRepo := &store.Audit{Pool: pool}
	wahaNotifier := notifier.New(resolver, "")

	orders := &handler.Orders{
		StoreID:      defaultStoreID,
		Orders:       &store.Orders{Pool: pool},
		Shipments:    &store.Shipments{Pool: pool},
		Events:       &store.Events{Pool: pool},
		Facets:       &store.Facets{Pool: pool},
		Audit:        auditRepo,
		Notifier:     wahaNotifier,
		Integrations: resolver,
		Log:          log,
	}
	mux.HandleFunc("GET /api/v1/orders", orders.List)
	mux.HandleFunc("GET /api/v1/orders/facets", orders.FacetsList)
	mux.HandleFunc("GET /api/v1/orders/export.csv", orders.ExportCSV)
	mux.HandleFunc("GET /api/v1/orders/{id}", orders.Get)
	mux.HandleFunc("GET /api/v1/orders/{id}/history", orders.History)
	mux.HandleFunc("PATCH /api/v1/orders/{id}/status", orders.UpdateStatus)
	mux.HandleFunc("POST /api/v1/orders/{id}/notify", orders.Notify)

	analytics := &handler.Analytics{
		Repo: &store.Analytics{Pool: pool},
		Log:  log,
	}
	mux.HandleFunc("GET /api/v1/analytics/overview", analytics.Overview)

	stateRepo := &store.SyncStates{Pool: pool}
	wcSync := &syncpkg.WooCommerce{
		Store:        &store.Orders{Pool: pool},
		Shipments:    &store.Shipments{Pool: pool},
		State:        stateRepo,
		Integrations: resolver,
		StoreID:      defaultStoreID,
		Log:          log,
	}
	frenetSync := &syncpkg.Frenet{
		Shipments:    &store.Shipments{Pool: pool},
		Events:       &store.Events{Pool: pool},
		Integrations: resolver,
		BatchSize:    50,
		Log:          log,
	}
	syncH := &handler.SyncH{State: stateRepo, WC: wcSync, Frenet: frenetSync, Log: log}
	mux.HandleFunc("GET /api/v1/sync/status", syncH.Status)
	mux.HandleFunc("POST /api/v1/sync/run", syncH.Trigger)

	settingsH := &handler.Settings{
		Store:    settingsStore,
		Resolver: resolver,
		Log:      log,
	}
	mux.HandleFunc("GET /api/v1/settings/integrations", settingsH.Get)
	mux.HandleFunc("PATCH /api/v1/settings/integrations/woocommerce", settingsH.UpdateWooCommerce)
	mux.HandleFunc("PATCH /api/v1/settings/integrations/frenet", settingsH.UpdateFrenet)
	mux.HandleFunc("PATCH /api/v1/settings/integrations/waha", settingsH.UpdateWAHA)
	mux.HandleFunc("POST /api/v1/settings/integrations/woocommerce/test", settingsH.TestWooCommerce)
	mux.HandleFunc("POST /api/v1/settings/integrations/frenet/test", settingsH.TestFrenet)
	mux.HandleFunc("POST /api/v1/settings/integrations/waha/test", settingsH.TestWAHA)

	return loggingMiddleware(log)(corsMiddleware(cfg.App.PublicURL)(mux))
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func readyHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status": "db_unavailable",
				"error":  err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
	}
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    "univer-tracker",
		"version": "0.3.0",
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
