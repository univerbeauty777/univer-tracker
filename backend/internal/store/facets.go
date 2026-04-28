package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Facets surfaces the distinct values currently in use so the dashboard
// can populate filter dropdowns with what's actually there (no phantom
// options for carriers we no longer ship via).
type Facets struct {
	Pool *pgxpool.Pool
}

// FacetCounts is what the UI receives.
type FacetCounts struct {
	Carriers []FacetValue `json:"carriers"`
	UFs      []FacetValue `json:"ufs"`
	Statuses []FacetValue `json:"statuses"`
	Health   []FacetValue `json:"health"`
}

// FacetValue is a label/count pair.
type FacetValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// All returns every facet in a single round trip.
func (f *Facets) All(ctx context.Context, storeID int64) (FacetCounts, error) {
	out := FacetCounts{}

	if err := f.fetch(ctx, &out.Carriers, `
SELECT s.carrier, COUNT(DISTINCT o.id)
FROM orders o
LEFT JOIN shipments s ON s.order_id = o.id
WHERE o.store_id = $1 AND s.carrier IS NOT NULL AND s.carrier <> ''
GROUP BY s.carrier
ORDER BY COUNT(DISTINCT o.id) DESC, s.carrier`, storeID); err != nil {
		return out, fmt.Errorf("carriers: %w", err)
	}

	if err := f.fetch(ctx, &out.UFs, `
SELECT customer_uf, COUNT(*)
FROM orders
WHERE store_id = $1 AND customer_uf <> ''
GROUP BY customer_uf
ORDER BY COUNT(*) DESC, customer_uf`, storeID); err != nil {
		return out, fmt.Errorf("ufs: %w", err)
	}

	if err := f.fetch(ctx, &out.Statuses, `
SELECT status, COUNT(*)
FROM orders
WHERE store_id = $1 AND status <> ''
GROUP BY status
ORDER BY COUNT(*) DESC, status`, storeID); err != nil {
		return out, fmt.Errorf("statuses: %w", err)
	}

	if err := f.fetch(ctx, &out.Health, `
SELECT s.health, COUNT(*)
FROM orders o
LEFT JOIN shipments s ON s.order_id = o.id
WHERE o.store_id = $1 AND s.health IS NOT NULL
GROUP BY s.health
ORDER BY COUNT(*) DESC, s.health`, storeID); err != nil {
		return out, fmt.Errorf("health: %w", err)
	}

	return out, nil
}

func (f *Facets) fetch(ctx context.Context, dst *[]FacetValue, q string, args ...any) error {
	rows, err := f.Pool.Query(ctx, q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var v FacetValue
		if err := rows.Scan(&v.Value, &v.Count); err != nil {
			return err
		}
		*dst = append(*dst, v)
	}
	return rows.Err()
}
