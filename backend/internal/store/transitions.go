package store

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

// Transitions runs the rastreiaki gargalos analytics: per-stage averages,
// percentiles and breach rates per carrier.
type Transitions struct {
	Pool *pgxpool.Pool
}

// FunnelStage = quantos envios alcançaram cada etapa.
type FunnelStage struct {
	Field string `json:"field"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// Transition aggregates a stage-to-stage move.
type Transition struct {
	Field      string             `json:"field"`
	Label      string             `json:"label"`
	Count      int                `json:"count"`
	AvgHours   float64            `json:"avg_hours"`
	P50Hours   float64            `json:"p50_hours"`
	P90Hours   float64            `json:"p90_hours"`
	BreachRate float64            `json:"breach_rate"`
	ByCarrier  map[string]CarrCnt `json:"by_carrier"`
}

// CarrCnt is one cell of the heatmap: how a carrier performs on a stage.
type CarrCnt struct {
	Count      int     `json:"count"`
	AvgHours   float64 `json:"avg_hours"`
	BreachRate float64 `json:"breach_rate"`
}

// Each stage anchored from order created_at, in cumulative hours from
// orders.created_at. Mirrors rastreiaki SLA_POLICIES; for the breach
// calculation we pick the policy per carrier in Go below.
var stageDefs = []struct {
	Field, Label, ShipCol string
}{
	{"label_issued_at", "Pedido → Etiqueta emitida", "label_issued_at"},
	{"preparing_at", "Etiqueta → Em preparação", "preparing_at"},
	{"ready_for_pickup_at", "Preparação → Pronto p/ coleta", "ready_for_pickup_at"},
	{"posted_at", "Pronto → Postado", "posted_at"},
	{"out_for_delivery_at", "Postado → Saiu p/ entrega", "out_for_delivery_at"},
	{"delivered_at", "Saiu → Entregue", "delivered_at"},
}

// Funnel returns how many shipments reached each stage in the last
// `windowDays` (default 30).
func (r *Transitions) Funnel(ctx context.Context, windowDays int) ([]FunnelStage, error) {
	if windowDays <= 0 {
		windowDays = 30
	}
	q := fmt.Sprintf(`
SELECT
    COUNT(*) FILTER (WHERE s.created_at >= NOW() - INTERVAL '%d days')                         AS pedido,
    COUNT(*) FILTER (WHERE s.label_issued_at        IS NOT NULL AND s.created_at >= NOW() - INTERVAL '%d days') AS etiqueta,
    COUNT(*) FILTER (WHERE s.preparing_at           IS NOT NULL AND s.created_at >= NOW() - INTERVAL '%d days') AS preparando,
    COUNT(*) FILTER (WHERE s.ready_for_pickup_at    IS NOT NULL AND s.created_at >= NOW() - INTERVAL '%d days') AS coleta,
    COUNT(*) FILTER (WHERE s.posted_at              IS NOT NULL AND s.created_at >= NOW() - INTERVAL '%d days') AS postado,
    COUNT(*) FILTER (WHERE s.out_for_delivery_at    IS NOT NULL AND s.created_at >= NOW() - INTERVAL '%d days') AS saiu,
    COUNT(*) FILTER (WHERE s.delivered_at           IS NOT NULL AND s.created_at >= NOW() - INTERVAL '%d days') AS entregue
FROM shipments s
JOIN orders o ON o.id = s.order_id
WHERE s.tracking_code <> '' AND o.hidden_at IS NULL`,
		windowDays, windowDays, windowDays, windowDays, windowDays, windowDays, windowDays)

	var pedido, etiqueta, prep, coleta, postado, saiu, entregue int
	if err := r.Pool.QueryRow(ctx, q).Scan(&pedido, &etiqueta, &prep, &coleta, &postado, &saiu, &entregue); err != nil {
		return nil, fmt.Errorf("funnel: %w", err)
	}
	return []FunnelStage{
		{Field: "order", Label: "Pedido criado", Count: pedido},
		{Field: "label_issued_at", Label: "Etiqueta emitida", Count: etiqueta},
		{Field: "preparing_at", Label: "Em preparação", Count: prep},
		{Field: "ready_for_pickup_at", Label: "Pronto p/ coleta", Count: coleta},
		{Field: "posted_at", Label: "Postado", Count: postado},
		{Field: "out_for_delivery_at", Label: "Saiu p/ entrega", Count: saiu},
		{Field: "delivered_at", Label: "Entregue", Count: entregue},
	}, nil
}

// All returns the full transitions stat set used by /gargalos. Stages run
// in parallel — they are fully independent queries and 6 sequential
// percentile_cont scans were the dominant cause of the gargalos endpoint
// timing out under load. Concurrency is capped at len(stageDefs) so we
// never exceed the pool's appetite for a single request.
func (r *Transitions) All(ctx context.Context, windowDays int) ([]Transition, error) {
	if windowDays <= 0 {
		windowDays = 30
	}
	out := make([]Transition, len(stageDefs))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(len(stageDefs))
	for i, st := range stageDefs {
		i, st := i, st
		g.Go(func() error {
			t, err := r.computeStage(gctx, st.Field, st.Label, st.ShipCol, windowDays)
			if err != nil {
				return err
			}
			out[i] = t
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

// computeStage runs one query per stage. The breach threshold uses the
// SLA hard-coded as a CASE — keeps the SQL portable without a separate
// policies table.
func (r *Transitions) computeStage(ctx context.Context, field, label, col string, windowDays int) (Transition, error) {
	// SLA cumulative hours per carrier for this stage. Falls back to
	// generous Correios defaults when the carrier slug isn't recognised.
	slaCase := buildSLACaseSQL(col)

	q := fmt.Sprintf(`
WITH base AS (
    SELECT
        s.carrier,
        EXTRACT(EPOCH FROM (s.%[1]s - s.created_at)) / 3600.0 AS hours,
        %[2]s AS sla_hours
    FROM shipments s
    JOIN orders o ON o.id = s.order_id
    WHERE s.%[1]s IS NOT NULL
      AND s.created_at >= NOW() - INTERVAL '%[3]d days'
      AND o.hidden_at IS NULL
)
SELECT
    COALESCE(NULLIF(carrier, ''), 'desconhecida') AS carrier,
    COUNT(*),
    COALESCE(AVG(hours), 0),
    COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY hours), 0),
    COALESCE(percentile_cont(0.9) WITHIN GROUP (ORDER BY hours), 0),
    COALESCE(SUM(CASE WHEN sla_hours IS NOT NULL AND hours > sla_hours THEN 1 ELSE 0 END), 0)
FROM base
GROUP BY 1
ORDER BY 2 DESC`, col, slaCase, windowDays)

	rows, err := r.Pool.Query(ctx, q)
	if err != nil {
		return Transition{}, fmt.Errorf("transition %s: %w", field, err)
	}
	defer rows.Close()

	t := Transition{Field: field, Label: label, ByCarrier: map[string]CarrCnt{}}
	totalCount := 0
	totalBreach := 0
	var sumHours float64
	var sumP50, sumP90 float64
	for rows.Next() {
		var carrier string
		var count int
		var avg, p50, p90 float64
		var breaches int
		if err := rows.Scan(&carrier, &count, &avg, &p50, &p90, &breaches); err != nil {
			return Transition{}, err
		}
		breachRate := 0.0
		if count > 0 {
			breachRate = float64(breaches) / float64(count) * 100
		}
		t.ByCarrier[carrier] = CarrCnt{Count: count, AvgHours: avg, BreachRate: breachRate}
		totalCount += count
		totalBreach += breaches
		sumHours += avg * float64(count)
		sumP50 += p50 * float64(count)
		sumP90 += p90 * float64(count)
	}
	t.Count = totalCount
	if totalCount > 0 {
		t.AvgHours = sumHours / float64(totalCount)
		t.P50Hours = sumP50 / float64(totalCount)
		t.P90Hours = sumP90 / float64(totalCount)
		t.BreachRate = float64(totalBreach) / float64(totalCount) * 100
	}
	if math.IsNaN(t.AvgHours) {
		t.AvgHours = 0
	}
	return t, rows.Err()
}

// buildSLACaseSQL builds a SQL CASE expression that returns the cumulative
// SLA hours for the given column, by carrier — kept inline so we don't need
// a delivery_slas table yet.
func buildSLACaseSQL(col string) string {
	type entry struct {
		Carrier string
		Hours   int
	}
	// Cumulative SLA hours per carrier per stage column.
	policies := map[string]map[string]int{
		"label_issued_at": {
			"Correios - PAC": 4, "Correios - SEDEX": 2, "Correios": 4,
			"Jadlog (Melhor Envio)": 4, "Jadlog": 4,
			"Loggi": 4, "DHL": 2, "FedEx": 2, "Motoboy": 1,
		},
		"preparing_at": {
			"Correios - PAC": 12, "Correios - SEDEX": 8, "Correios": 12,
			"Jadlog (Melhor Envio)": 12, "Jadlog": 12,
			"Loggi": 12, "DHL": 6, "FedEx": 6, "Motoboy": 2,
		},
		"ready_for_pickup_at": {
			"Correios - PAC": 24, "Correios - SEDEX": 16, "Correios": 24,
			"Jadlog (Melhor Envio)": 24, "Jadlog": 24,
			"Loggi": 24, "DHL": 12, "FedEx": 12, "Motoboy": 3,
		},
		"posted_at": {
			"Correios - PAC": 48, "Correios - SEDEX": 24, "Correios": 48,
			"Jadlog (Melhor Envio)": 36, "Jadlog": 36,
			"Loggi": 36, "DHL": 24, "FedEx": 24, "Motoboy": 4,
		},
		"out_for_delivery_at": {
			"Correios - PAC": 120, "Correios - SEDEX": 72, "Correios": 120,
			"Jadlog (Melhor Envio)": 72, "Jadlog": 72,
			"Loggi": 60, "DHL": 48, "FedEx": 48, "Motoboy": 6,
		},
		"delivered_at": {
			"Correios - PAC": 168, "Correios - SEDEX": 96, "Correios": 168,
			"Jadlog (Melhor Envio)": 96, "Jadlog": 96,
			"Loggi": 72, "DHL": 72, "FedEx": 72, "Motoboy": 8,
		},
	}

	cases := ""
	if p, ok := policies[col]; ok {
		for carrier, hours := range p {
			cases += fmt.Sprintf("WHEN s.carrier = '%s' THEN %d ", escape(carrier), hours)
		}
	}
	// Default fallback — Correios PAC numbers above.
	defaults := map[string]int{
		"label_issued_at": 4, "preparing_at": 12, "ready_for_pickup_at": 24,
		"posted_at": 48, "out_for_delivery_at": 120, "delivered_at": 168,
	}
	d := defaults[col]
	return fmt.Sprintf("CASE %s ELSE %d END", cases, d)
}

func escape(s string) string {
	out := ""
	for _, r := range s {
		if r == '\'' {
			out += "''"
		} else {
			out += string(r)
		}
	}
	return out
}
