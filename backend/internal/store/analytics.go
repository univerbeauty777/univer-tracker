package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Analytics serves the dashboard summary queries.
type Analytics struct {
	Pool *pgxpool.Pool
}

// Overview is the headline KPI block shown at the top of the dashboard.
type Overview struct {
	Total30d        int     `json:"total_30d"`
	Delivered30d    int     `json:"delivered_30d"`
	OnTime30d       int     `json:"on_time_30d"`
	OnTimeRate      float64 `json:"on_time_rate"`
	AtRisk          int     `json:"at_risk"`
	Breached        int     `json:"breached"`
	InProgress      int     `json:"in_progress"`
	AvgDeliveryDays float64 `json:"avg_delivery_days"`
	IdleAlarms      int     `json:"idle_alarms"`

	// Per-phase averages in hours (rastreiaki painel).
	AvgPreparingHours float64 `json:"avg_preparing_hours"`
	AvgInTransitHours float64 `json:"avg_in_transit_hours"`
	AvgLastMileHours  float64 `json:"avg_last_mile_hours"`

	// PreviousPeriod is the same KPI block for the prior 30-day window
	// so the dashboard can render up/down deltas.
	PreviousPeriod *PreviousPeriod `json:"previous_period,omitempty"`
}

// PreviousPeriod holds the KPI snapshot for the 30-day window ending
// 30 days ago — comparable to Overview for delta computation.
type PreviousPeriod struct {
	Total30d        int     `json:"total_30d"`
	Delivered30d    int     `json:"delivered_30d"`
	OnTime30d       int     `json:"on_time_30d"`
	OnTimeRate      float64 `json:"on_time_rate"`
	AvgDeliveryDays float64 `json:"avg_delivery_days"`
}

// CarrierStats summarises performance per carrier in the last 30 days.
type CarrierStats struct {
	Carrier         string  `json:"carrier"`
	Total           int     `json:"total"`
	Breached        int     `json:"breached"`
	AvgDeliveryDays float64 `json:"avg_delivery_days"`
}

// FetchOverview computes the headline KPIs in a single query.
//
// Counts only shipments with a tracking_code: orders we can't track are
// not "in the funnel" — counting them as breached/idle would inflate
// the alerts and hide the real problems.
func (a *Analytics) FetchOverview(ctx context.Context) (*Overview, error) {
	// Lead-time uses o.created_at (the actual order date) — s.created_at is
	// when our DB inserted the row, which can be AFTER delivered_at for
	// historical backfills, producing a negative delta that gets filtered
	// out by COALESCE-on-NULL and zeroes the KPI.
	// "Preparação" is order → posted_at (much more reliably populated than
	// ready_for_pickup_at, which depends on a Frenet event that not every
	// carrier emits).
	const q = `
SELECT
    COUNT(*) FILTER (WHERE s.created_at >= NOW() - INTERVAL '30 days')                                        AS total_30d,
    COUNT(*) FILTER (WHERE s.delivered_at IS NOT NULL AND s.delivered_at >= NOW() - INTERVAL '30 days')       AS delivered_30d,
    COUNT(*) FILTER (
        WHERE s.sla_state = 'COMPLETED'
          AND s.delivered_at IS NOT NULL
          AND s.delivered_at >= NOW() - INTERVAL '30 days'
    )                                                                                                          AS on_time_30d,
    COUNT(*) FILTER (WHERE s.sla_state = 'AT_RISK' AND s.delivered_at IS NULL)                                AS at_risk,
    COUNT(*) FILTER (WHERE s.sla_state IN ('BREACHED', 'COMPLETED_LATE'))                                     AS breached,
    COUNT(*) FILTER (WHERE s.delivered_at IS NULL)                                                            AS in_progress,
    COALESCE(
        AVG(EXTRACT(EPOCH FROM (s.delivered_at - o.created_at)) / 86400.0)
        FILTER (
            WHERE s.delivered_at IS NOT NULL
              AND s.delivered_at >= NOW() - INTERVAL '30 days'
              AND s.delivered_at > o.created_at
        ),
        0
    )                                                                                                          AS avg_delivery_days,
    COUNT(*) FILTER (
        WHERE s.idle_since IS NOT NULL
          AND s.idle_since < NOW() - INTERVAL '4 days'
          AND s.delivered_at IS NULL
          AND s.last_event_at IS NOT NULL
    )                                                                                                          AS idle_alarms,
    COALESCE(
        AVG(EXTRACT(EPOCH FROM (s.posted_at - o.created_at)) / 3600.0)
        FILTER (
            WHERE s.posted_at IS NOT NULL
              AND o.created_at >= NOW() - INTERVAL '30 days'
              AND s.posted_at > o.created_at
        ),
        0
    )                                                                                                          AS avg_preparing_hours,
    COALESCE(
        AVG(EXTRACT(EPOCH FROM (s.delivered_at - s.posted_at)) / 3600.0)
        FILTER (
            WHERE s.delivered_at IS NOT NULL
              AND s.posted_at IS NOT NULL
              AND s.delivered_at >= NOW() - INTERVAL '30 days'
              AND s.delivered_at > s.posted_at
        ),
        0
    )                                                                                                          AS avg_in_transit_hours,
    COALESCE(
        AVG(EXTRACT(EPOCH FROM (s.delivered_at - s.out_for_delivery_at)) / 3600.0)
        FILTER (
            WHERE s.delivered_at IS NOT NULL
              AND s.out_for_delivery_at IS NOT NULL
              AND s.delivered_at >= NOW() - INTERVAL '30 days'
              AND s.delivered_at > s.out_for_delivery_at
        ),
        0
    )                                                                                                          AS avg_last_mile_hours
FROM shipments s
JOIN orders o ON o.id = s.order_id
WHERE s.tracking_code <> '' AND o.hidden_at IS NULL`

	var o Overview
	err := a.Pool.QueryRow(ctx, q).Scan(
		&o.Total30d, &o.Delivered30d, &o.OnTime30d,
		&o.AtRisk, &o.Breached, &o.InProgress,
		&o.AvgDeliveryDays, &o.IdleAlarms,
		&o.AvgPreparingHours, &o.AvgInTransitHours, &o.AvgLastMileHours,
	)
	if err != nil {
		return nil, fmt.Errorf("overview: %w", err)
	}
	if o.Delivered30d > 0 {
		o.OnTimeRate = float64(o.OnTime30d) / float64(o.Delivered30d)
	}

	prev, err := a.fetchPrevious(ctx)
	if err == nil {
		o.PreviousPeriod = prev
	}
	return &o, nil
}

func (a *Analytics) fetchPrevious(ctx context.Context) (*PreviousPeriod, error) {
	const q = `
SELECT
    COUNT(*) FILTER (
        WHERE s.created_at >= NOW() - INTERVAL '60 days'
          AND s.created_at < NOW() - INTERVAL '30 days'
    ) AS total,
    COUNT(*) FILTER (
        WHERE s.delivered_at IS NOT NULL
          AND s.delivered_at >= NOW() - INTERVAL '60 days'
          AND s.delivered_at < NOW() - INTERVAL '30 days'
    ) AS delivered,
    COUNT(*) FILTER (
        WHERE s.delivered_at IS NOT NULL
          AND s.delivered_at >= NOW() - INTERVAL '60 days'
          AND s.delivered_at < NOW() - INTERVAL '30 days'
          AND s.estimated_delivery IS NOT NULL
          AND s.delivered_at::date <= s.estimated_delivery
    ) AS on_time,
    COALESCE(
        AVG(EXTRACT(EPOCH FROM (s.delivered_at - s.created_at)) / 86400.0)
        FILTER (
            WHERE s.delivered_at IS NOT NULL
              AND s.delivered_at >= NOW() - INTERVAL '60 days'
              AND s.delivered_at < NOW() - INTERVAL '30 days'
        ),
        0
    ) AS avg_days
FROM shipments s
JOIN orders o ON o.id = s.order_id
WHERE s.tracking_code <> '' AND o.hidden_at IS NULL`

	var p PreviousPeriod
	err := a.Pool.QueryRow(ctx, q).Scan(&p.Total30d, &p.Delivered30d, &p.OnTime30d, &p.AvgDeliveryDays)
	if err != nil {
		return nil, err
	}
	if p.Delivered30d > 0 {
		p.OnTimeRate = float64(p.OnTime30d) / float64(p.Delivered30d)
	}
	return &p, nil
}

// FetchCarriers returns the top carriers by volume over the last 30 days.
func (a *Analytics) FetchCarriers(ctx context.Context, limit int) ([]CarrierStats, error) {
	if limit <= 0 || limit > 50 {
		limit = 5
	}
	const q = `
SELECT
    COALESCE(NULLIF(s.carrier, ''), 'desconhecida') AS carrier,
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE s.health = 'breached') AS breached,
    COALESCE(
        AVG(EXTRACT(EPOCH FROM (s.delivered_at - s.created_at)) / 86400.0) FILTER (WHERE s.delivered_at IS NOT NULL),
        0
    ) AS avg_days
FROM shipments s
JOIN orders o ON o.id = s.order_id
WHERE s.created_at >= NOW() - INTERVAL '30 days'
  AND s.tracking_code <> ''
  AND s.carrier <> ''
  AND o.hidden_at IS NULL
GROUP BY 1
ORDER BY total DESC
LIMIT $1`
	rows, err := a.Pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("carriers: %w", err)
	}
	defer rows.Close()
	var out []CarrierStats
	for rows.Next() {
		var c CarrierStats
		if err := rows.Scan(&c.Carrier, &c.Total, &c.Breached, &c.AvgDeliveryDays); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
