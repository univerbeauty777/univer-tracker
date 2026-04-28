// Package http wires the HTTP router and registers handlers.
package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/config"
	"github.com/univerbeauty777/univer-tracker/backend/internal/frenet"
	"github.com/univerbeauty777/univer-tracker/backend/internal/http/handler"
	"github.com/univerbeauty777/univer-tracker/backend/internal/woocommerce"
)

// NewRouter creates the application's main HTTP handler.
func NewRouter(cfg *config.Config, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", readyHandler)
	mux.HandleFunc("GET /api/v1/version", versionHandler)

	// External clients used by the orders handler.
	wc := woocommerce.New(cfg.WooCommerce.URL, cfg.WooCommerce.ConsumerKey, cfg.WooCommerce.ConsumerSecret)
	fc := frenet.New(cfg.Frenet.APIToken)

	orders := &handler.Orders{WC: wc, Frenet: fc, Log: log}
	mux.HandleFunc("GET /api/v1/orders", orders.List)
	mux.HandleFunc("GET /api/v1/orders/{id}", orders.Get)
	mux.HandleFunc("PATCH /api/v1/orders/{id}/status", orders.UpdateStatus)

	return loggingMiddleware(log)(corsMiddleware(cfg.App.PublicURL)(mux))
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func readyHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ready",
	})
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    "univer-tracker",
		"version": "0.1.0",
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
