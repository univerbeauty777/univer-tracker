package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EventKey enumerates the four shipping milestones the dashboard lets
// the user automate. Anything else gets rejected at the API boundary.
const (
	EventPostado   = "postado"
	EventInTransit = "in_transit"
	EventDelivered = "delivered"
	EventBreached  = "breached"
)

// AllEventKeys returns the canonical order the UI renders.
func AllEventKeys() []string {
	return []string{EventPostado, EventInTransit, EventDelivered, EventBreached}
}

// IsValidEventKey reports whether s is one of the four canonical keys.
func IsValidEventKey(s string) bool {
	for _, k := range AllEventKeys() {
		if k == s {
			return true
		}
	}
	return false
}

// NotificationTrigger is the per-event automation row.
type NotificationTrigger struct {
	ID              int64     `json:"id"`
	StoreID         int64     `json:"store_id"`
	EventKey        string    `json:"event_key"`
	Enabled         bool      `json:"enabled"`
	Template        string    `json:"template"`
	Session         string    `json:"session"`
	CooldownMinutes int       `json:"cooldown_minutes"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NotificationTriggers is the repository.
type NotificationTriggers struct {
	Pool    *pgxpool.Pool
	StoreID int64
}

// List returns all triggers for the store, ordered by the canonical
// event sequence (postado → in_transit → delivered → breached). Missing
// events are filled with disabled defaults so the UI always has 4 rows.
func (r *NotificationTriggers) List(ctx context.Context) ([]NotificationTrigger, error) {
	rows, err := r.Pool.Query(ctx, `
SELECT id, store_id, event_key, enabled, template, session, cooldown_minutes, updated_at
FROM notification_triggers
WHERE store_id = $1`, r.StoreID)
	if err != nil {
		return nil, fmt.Errorf("list triggers: %w", err)
	}
	defer rows.Close()

	byKey := map[string]NotificationTrigger{}
	for rows.Next() {
		var t NotificationTrigger
		if err := rows.Scan(
			&t.ID, &t.StoreID, &t.EventKey, &t.Enabled,
			&t.Template, &t.Session, &t.CooldownMinutes, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan trigger: %w", err)
		}
		byKey[t.EventKey] = t
	}

	out := make([]NotificationTrigger, 0, 4)
	for _, key := range AllEventKeys() {
		if t, ok := byKey[key]; ok {
			out = append(out, t)
			continue
		}
		out = append(out, NotificationTrigger{
			StoreID:         r.StoreID,
			EventKey:        key,
			Enabled:         false,
			Template:        "",
			CooldownMinutes: 60,
		})
	}
	return out, rows.Err()
}

// Upsert persists a trigger row keyed on (store_id, event_key).
func (r *NotificationTriggers) Upsert(ctx context.Context, t *NotificationTrigger) error {
	if !IsValidEventKey(t.EventKey) {
		return fmt.Errorf("invalid event_key %q", t.EventKey)
	}
	if t.CooldownMinutes < 0 {
		t.CooldownMinutes = 0
	}
	_, err := r.Pool.Exec(ctx, `
INSERT INTO notification_triggers (
    store_id, event_key, enabled, template, session, cooldown_minutes, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (store_id, event_key) DO UPDATE SET
    enabled          = EXCLUDED.enabled,
    template         = EXCLUDED.template,
    session          = EXCLUDED.session,
    cooldown_minutes = EXCLUDED.cooldown_minutes,
    updated_at       = NOW()`,
		r.StoreID, t.EventKey, t.Enabled, t.Template, t.Session, t.CooldownMinutes)
	if err != nil {
		return fmt.Errorf("upsert trigger: %w", err)
	}
	return nil
}

// EnabledByKey returns the cached enabled triggers indexed by event_key,
// for the worker hot-path. Disabled triggers are filtered out.
func (r *NotificationTriggers) EnabledByKey(ctx context.Context) (map[string]NotificationTrigger, error) {
	rows, err := r.Pool.Query(ctx, `
SELECT id, store_id, event_key, enabled, template, session, cooldown_minutes, updated_at
FROM notification_triggers
WHERE store_id = $1 AND enabled = TRUE AND template <> ''`, r.StoreID)
	if err != nil {
		return nil, fmt.Errorf("enabled triggers: %w", err)
	}
	defer rows.Close()

	out := map[string]NotificationTrigger{}
	for rows.Next() {
		var t NotificationTrigger
		if err := rows.Scan(
			&t.ID, &t.StoreID, &t.EventKey, &t.Enabled,
			&t.Template, &t.Session, &t.CooldownMinutes, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan trigger: %w", err)
		}
		out[t.EventKey] = t
	}
	return out, rows.Err()
}
