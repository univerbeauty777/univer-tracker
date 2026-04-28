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
	Total30d         int     `json:"total_30d"`
	Delivered30d     int     `json:"delivered_30d"`
	OnTime30d        int     `json:"on_time_30d"`
	OnTimeRate       float64 `json:"on_time_rate"`
	AtRisk           int     `json:"at_risk"`
	Breached         int     `json:"breached"`
	AvgDeliveryDays  float64 `json:"avg_delivery_days"`
	IdleAlarms       int     `json:"idle_alarms"`
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
	const q = `
SELECT
    COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '30 days')                                          AS total_30d,
    COUNT(*) FILTER (WHERE delivered_at IS NOT NULL AND delivered_at >= NOW() - INTERVAL '30 days')           AS delivered_30d,
    COUNT(*) FILTER (
        WHERE delivered_at IS NOT NULL
          AND delivered_at >= NOW() - INTERVAL '30 days'
          AND estimated_delivery IS NOT NULL
          AND delivered_at::date <= estimated_delivery
    )                                                                                                          AS on_time_30d,
    COUNT(*) FILTER (WHERE health = 'at_risk')                                                                 AS at_risk,
    COUNT(*) FILTER (WHERE health = 'breached')                                                                AS breached,
    COALESCE(
        AVG(EXTRACT(EPOCH FROM (delivered_at - created_at)) / 86400.0)
        FILTER (WHERE delivered_at IS NOT NULL AND delivered_at >= NOW() - INTERVAL '30 days'),
        0
    )                                                                                                          AS avg_delivery_days,
    COUNT(*) FILTER (
        WHERE idle_since IS NOT NULL
          AND idle_since < NOW() - INTERVAL '4 days'
          AND delivered_at IS NULL
          AND last_event_at IS NOT NULL
    )                                                                                                          AS idle_alarms
FROM shipments
WHERE tracking_code <> ''`

	var o Overview
	err := a.Pool.QueryRow(ctx, q).Scan(
		&o.Total30d, &o.Delivered30d, &o.OnTime30d,
		&o.AtRisk, &o.Breached,
		&o.AvgDeliveryDays, &o.IdleAlarms,
	)
	if err != nil {
		return nil, fmt.Errorf("overview: %w", err)
	}
	if o.Delivered30d > 0 {
		o.OnTimeRate = float64(o.OnTime30d) / float64(o.Delivered30d)
	}
	return &o, nil
}

// FetchCarriers returns the top carriers by volume over the last 30 days.
func (a *Analytics) FetchCarriers(ctx context.Context, limit int) ([]CarrierStats, error) {
	if limit <= 0 || limit > 50 {
		limit = 5
	}
	const q = `
SELECT
    COALESCE(NULLIF(carrier, ''), 'desconhecida') AS carrier,
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE health = 'breached') AS breached,
    COALESCE(
        AVG(EXTRACT(EPOCH FROM (delivered_at - created_at)) / 86400.0) FILTER (WHERE delivered_at IS NOT NULL),
        0
    ) AS avg_days
FROM shipments
WHERE created_at >= NOW() - INTERVAL '30 days'
  AND tracking_code <> ''
  AND carrier <> ''
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
