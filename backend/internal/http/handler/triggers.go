package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// Triggers serves the per-event automation rules used by the worker to
// fire WhatsApp messages when a shipment crosses Postado / Em trânsito
// / Entregue / Atrasado.
type Triggers struct {
	Store *store.NotificationTriggers
	Log   *slog.Logger
}

// List handles GET /api/v1/settings/triggers.
func (h *Triggers) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.Store.List(ctx)
	if err != nil {
		h.Log.Error("list triggers", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load triggers")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"triggers": rows})
}

// Save handles PUT /api/v1/settings/triggers — receives the full set of
// 4 triggers and upserts each one. Mirrors how the UI edits them all
// together.
func (h *Triggers) Save(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Triggers []store.NotificationTrigger `json:"triggers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if len(body.Triggers) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "triggers is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	for i := range body.Triggers {
		t := &body.Triggers[i]
		t.EventKey = strings.TrimSpace(t.EventKey)
		t.Template = strings.TrimSpace(t.Template)
		t.Session = strings.TrimSpace(t.Session)
		if !store.IsValidEventKey(t.EventKey) {
			writeError(w, http.StatusUnprocessableEntity, "invalid event_key: "+t.EventKey)
			return
		}
		if t.Enabled && t.Template == "" {
			writeError(w, http.StatusUnprocessableEntity, "template required when enabled (event "+t.EventKey+")")
			return
		}
		if err := h.Store.Upsert(ctx, t); err != nil {
			h.Log.Error("upsert trigger", "event", t.EventKey, "err", err)
			writeError(w, http.StatusInternalServerError, "could not save triggers")
			return
		}
	}

	rows, err := h.Store.List(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "saved but reload failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"triggers": rows})
}
