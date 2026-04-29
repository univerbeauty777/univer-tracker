package sync

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BackfillCarriers normalises legacy carrier values in shipments. Earlier
// versions of the WC sync persisted shipping_method titles ("Frete grátis",
// "Expresso", "PAC Padrão") directly into shipments.carrier, which
// pollutes the dashboard with non-carrier names. We re-run the same
// inference rule the live sync uses now, joined against the order's
// stored shipping_method so we don't need to fetch from WC again.
type BackfillCarriers struct {
	Pool *pgxpool.Pool
	Log  *slog.Logger
}

// Run rewrites carrier where the current value is empty or matches a
// shipping-method-ish pattern. Idempotent. Returns the number of rows
// updated.
func (b *BackfillCarriers) Run(ctx context.Context) (int, error) {
	const q = `
SELECT s.id, COALESCE(o.shipping_method, '')
FROM shipments s
JOIN orders o ON o.id = s.order_id
WHERE s.tracking_code <> ''
  AND (
    s.carrier IS NULL OR s.carrier = '' OR
    LOWER(s.carrier) IN ('frete gratis','frete grátis','desconhecida','indefinido','padrao','padrão')
    OR LOWER(s.carrier) LIKE 'expresso%'
    OR LOWER(s.carrier) LIKE 'frete %'
    OR LOWER(s.carrier) LIKE 'econ%'
  )`
	rows, err := b.Pool.Query(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("list legacy carriers: %w", err)
	}
	defer rows.Close()

	type todo struct {
		id     int64
		method string
	}
	var pending []todo
	for rows.Next() {
		var t todo
		if err := rows.Scan(&t.id, &t.method); err != nil {
			return 0, err
		}
		pending = append(pending, t)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	updated := 0
	for _, t := range pending {
		carrier := inferCarrierFromMethod(t.method)
		if carrier == "" {
			continue
		}
		if _, err := b.Pool.Exec(ctx, `
UPDATE shipments
SET carrier = $1, updated_at = NOW()
WHERE id = $2`, carrier, t.id); err != nil {
			b.Log.Warn("backfill carrier", "id", t.id, "err", err)
			continue
		}
		updated++
	}
	return updated, nil
}
