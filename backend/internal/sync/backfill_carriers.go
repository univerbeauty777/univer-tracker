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

// canonicalCarrier is the set of values inferCarrierFromMethod produces.
// Shipments whose carrier is anything else are treated as legacy and
// re-derived from the order's shipping_method.
var canonicalCarrier = map[string]bool{
	"Correios":          true,
	"Correios - PAC":    true,
	"Correios - SEDEX":  true,
	"Jadlog":            true,
	"Loggi":             true,
	"Azul Cargo":        true,
	"Total Express":     true,
	"Braspress":         true,
	"DHL":               true,
	"FedEx":             true,
	"Motoboy":           true,
}

// Run rewrites carrier where the current value is empty or doesn't
// belong to the canonical set produced by inferCarrierFromMethod
// (covers shipping_method titles like "Frete grátis", "Mini Envios..."
// AND case mismatches like "Correios - Sedex"). Idempotent.
func (b *BackfillCarriers) Run(ctx context.Context) (int, error) {
	const q = `
SELECT s.id, s.carrier, COALESCE(o.shipping_method, '')
FROM shipments s
JOIN orders o ON o.id = s.order_id
WHERE s.tracking_code <> ''`
	rows, err := b.Pool.Query(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("list shipments: %w", err)
	}
	defer rows.Close()

	type todo struct {
		id      int64
		carrier string
		method  string
	}
	var pending []todo
	for rows.Next() {
		var t todo
		if err := rows.Scan(&t.id, &t.carrier, &t.method); err != nil {
			return 0, err
		}
		if !canonicalCarrier[t.carrier] {
			pending = append(pending, t)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	updated := 0
	for _, t := range pending {
		carrier := inferCarrierFromMethod(t.method)
		if carrier == "" || carrier == t.carrier {
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
