package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StatusChange is one row in the order audit log.
type StatusChange struct {
	ID         int64     `json:"id"`
	OrderID    int64     `json:"order_id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Source     string    `json:"source"` // manual | sync | webhook
	Note       string    `json:"note"`
	Actor      string    `json:"actor"`
	CreatedAt  time.Time `json:"created_at"`
}

// Notification is one customer notification sent through any channel.
type Notification struct {
	ID       int64           `json:"id"`
	OrderID  int64           `json:"order_id"`
	Channel  string          `json:"channel"`
	Template string          `json:"template"`
	Payload  json.RawMessage `json:"payload,omitempty"`
	Status   string          `json:"status"`
	Error    string          `json:"error,omitempty"`
	SentAt   time.Time       `json:"sent_at"`
}

// Audit persists status changes and customer notifications.
type Audit struct {
	Pool *pgxpool.Pool
}

// RecordStatusChange appends a row. Caller is responsible for the WC update;
// this is the local mirror.
func (a *Audit) RecordStatusChange(ctx context.Context, c StatusChange) error {
	const q = `
INSERT INTO status_changes (order_id, from_status, to_status, source, note, actor)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := a.Pool.Exec(ctx, q, c.OrderID, c.FromStatus, c.ToStatus,
		nonEmpty(c.Source, "manual"), c.Note, nonEmpty(c.Actor, "system"))
	if err != nil {
		return fmt.Errorf("record status change: %w", err)
	}
	return nil
}

// ListStatusChanges returns the order's audit log newest first.
func (a *Audit) ListStatusChanges(ctx context.Context, orderID int64) ([]StatusChange, error) {
	const q = `
SELECT id, order_id, from_status, to_status, source, note, actor, created_at
FROM status_changes
WHERE order_id = $1
ORDER BY created_at DESC
LIMIT 200`
	rows, err := a.Pool.Query(ctx, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("list status changes: %w", err)
	}
	defer rows.Close()
	var out []StatusChange
	for rows.Next() {
		var c StatusChange
		if err := rows.Scan(&c.ID, &c.OrderID, &c.FromStatus, &c.ToStatus, &c.Source, &c.Note, &c.Actor, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// RecordNotification appends a notification row.
func (a *Audit) RecordNotification(ctx context.Context, n Notification) error {
	const q = `
INSERT INTO notifications (order_id, channel, template, payload, status, error)
VALUES ($1, $2, $3, COALESCE($4, '{}'::jsonb), $5, $6)`
	_, err := a.Pool.Exec(ctx, q, n.OrderID, n.Channel, n.Template, n.Payload,
		nonEmpty(n.Status, "sent"), n.Error)
	if err != nil {
		return fmt.Errorf("record notification: %w", err)
	}
	return nil
}

// HasNotificationSince returns true when there's a notifications row
// for (order_id, template) sent at or after `since` with status != 'failed'.
// Used by the trigger dispatcher to enforce per-event cooldowns and
// avoid double-firing when Frenet replays the same milestone event.
//
// A `since` of the zero value matches any prior successful send (any-time
// dedup, useful for cooldown=0 → "send at most once ever per shipment").
func (a *Audit) HasNotificationSince(ctx context.Context, orderID int64, template string, since time.Time) (bool, error) {
	const q = `
SELECT EXISTS (
    SELECT 1 FROM notifications
    WHERE order_id = $1
      AND template = $2
      AND status <> 'failed'
      AND sent_at >= $3
)`
	var exists bool
	if err := a.Pool.QueryRow(ctx, q, orderID, template, since).Scan(&exists); err != nil {
		return false, fmt.Errorf("has notification since: %w", err)
	}
	return exists, nil
}

// ListNotifications returns notifications for an order, newest first.
func (a *Audit) ListNotifications(ctx context.Context, orderID int64) ([]Notification, error) {
	const q = `
SELECT id, order_id, channel, template, payload, status, error, sent_at
FROM notifications
WHERE order_id = $1
ORDER BY sent_at DESC
LIMIT 100`
	rows, err := a.Pool.Query(ctx, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	var out []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.OrderID, &n.Channel, &n.Template, &n.Payload, &n.Status, &n.Error, &n.SentAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
