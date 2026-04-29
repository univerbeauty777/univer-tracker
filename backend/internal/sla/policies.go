package sla

import "strings"

// Policy is the per-stage SLA in cumulative hours from order creation.
// Values match the rastreiaki reference (data.js SLA_POLICIES) so the UI
// renders identical bands.
type Policy struct {
	Label     int // pedido → etiqueta emitida
	Prep      int // → preparando
	Ready     int // → pronto p/ coleta
	Posted    int // → postado
	OFD       int // → saiu p/ entrega
	Delivered int // → entregue
}

// PolicyFor returns the policy for a carrier slug. Falls back to Correios
// when nothing matches — operators rarely use a carrier outside this list.
func PolicyFor(carrier, service string) Policy {
	c := strings.ToLower(carrier)
	s := strings.ToLower(service)

	switch {
	case strings.Contains(c, "motoboy"):
		return Policy{Label: 1, Prep: 2, Ready: 3, Posted: 4, OFD: 6, Delivered: 8}
	case strings.Contains(c, "loggi"):
		return Policy{Label: 4, Prep: 12, Ready: 24, Posted: 36, OFD: 60, Delivered: 72}
	case strings.Contains(c, "dhl"):
		return Policy{Label: 2, Prep: 6, Ready: 12, Posted: 24, OFD: 48, Delivered: 72}
	case strings.Contains(c, "fedex"):
		return Policy{Label: 2, Prep: 6, Ready: 12, Posted: 24, OFD: 48, Delivered: 72}
	case strings.Contains(c, "jadlog"):
		return Policy{Label: 4, Prep: 12, Ready: 24, Posted: 36, OFD: 72, Delivered: 96}
	case strings.Contains(c, "j&t"), strings.Contains(c, "jt express"):
		return Policy{Label: 4, Prep: 12, Ready: 24, Posted: 48, OFD: 96, Delivered: 120}
	case strings.Contains(c, "correios") && strings.Contains(s, "sedex"):
		return Policy{Label: 2, Prep: 8, Ready: 16, Posted: 24, OFD: 72, Delivered: 96}
	}
	// Correios PAC default
	return Policy{Label: 4, Prep: 12, Ready: 24, Posted: 48, OFD: 120, Delivered: 168}
}

// Stage represents a milestone ordered by cumulative SLA.
type Stage struct {
	Field      string // store column name
	JSONField  string // JSON field in API responses
	Label      string // PT-BR display label
	CumHours   int    // cumulative SLA hours
}

// Stages returns the milestones in chronological order for a given carrier,
// each with its cumulative SLA in hours.
func Stages(carrier, service string) []Stage {
	p := PolicyFor(carrier, service)
	return []Stage{
		{Field: "label_issued_at", JSONField: "label_issued_at", Label: "Etiqueta emitida", CumHours: p.Label},
		{Field: "preparing_at", JSONField: "preparing_at", Label: "Em preparação", CumHours: p.Prep},
		{Field: "ready_for_pickup_at", JSONField: "ready_for_pickup_at", Label: "Pronto p/ coleta", CumHours: p.Ready},
		{Field: "posted_at", JSONField: "posted_at", Label: "Postado", CumHours: p.Posted},
		{Field: "out_for_delivery_at", JSONField: "out_for_delivery_at", Label: "Saiu p/ entrega", CumHours: p.OFD},
		{Field: "delivered_at", JSONField: "delivered_at", Label: "Entregue", CumHours: p.Delivered},
	}
}
