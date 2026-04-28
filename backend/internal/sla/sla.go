// Package sla computes shipment health (on track / at risk / breached)
// from a delivery SLA per carrier+service combination plus the shipment's
// own state (created date, last event, delivered date).
//
// Today the SLAs are hard-coded — these are the values our ops team
// observes for the BR carriers we use. A future iteration moves them to
// a `delivery_slas` table editable from the dashboard.
package sla

import (
	"strings"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// Health categorises a shipment for the dashboard.
type Health string

const (
	HealthUnknown   Health = "unknown"
	HealthOnTrack   Health = "on_track"
	HealthAtRisk    Health = "at_risk"
	HealthBreached  Health = "breached"
	HealthDelivered Health = "delivered"
)

// Result is the full output of Compute — useful for both the persisted
// shipment row and analytics.
type Result struct {
	Health            Health
	EstimatedDelivery time.Time
	IdleSince         time.Time
	IdleDays          int
	DaysOverSLA       int
	RiskScore         int16
}

// Days returns the carrier+service SLA in business-ish days. Generous on
// purpose — we'd rather flag at_risk a day too late than panic the team.
func Days(carrier, service string) int {
	c := normalize(carrier)
	s := normalize(service)

	switch {
	case strings.Contains(c, "correios") && strings.Contains(s, "sedex"):
		return 4
	case strings.Contains(c, "correios") && strings.Contains(s, "pac"):
		return 8
	case strings.Contains(c, "correios"):
		return 6
	case strings.Contains(c, "jadlog"):
		return 5
	case strings.Contains(c, "loggi"):
		return 3
	case strings.Contains(c, "azul"):
		return 4
	case strings.Contains(c, "total"):
		return 5
	case strings.Contains(c, "braspress"):
		return 6
	case strings.Contains(c, "dhl"):
		return 5
	case strings.Contains(c, "fedex"):
		return 5
	}
	return 7
}

// Compute classifies the shipment given the current time. Pure function:
// no I/O, no globals — easy to unit test as we add more rules.
func Compute(ship *store.Shipment, now time.Time) Result {
	r := Result{Health: HealthUnknown}

	// Already delivered? Nothing else to compute.
	if ship.DeliveredAt != nil && !ship.DeliveredAt.IsZero() {
		r.Health = HealthDelivered
		return r
	}

	// Estimated delivery = createdAt + SLA days. We use the shipment's
	// created_at because that's when we know the carrier accepted the
	// label (or at least when we became aware of it).
	days := Days(ship.Carrier, ship.Service)
	base := ship.CreatedAt
	if base.IsZero() {
		base = now
	}
	r.EstimatedDelivery = base.AddDate(0, 0, days)

	// Idle window — how long since the carrier last reported anything.
	idleAnchor := base
	if ship.LastEventAt != nil && ship.LastEventAt.After(idleAnchor) {
		idleAnchor = *ship.LastEventAt
	}
	r.IdleSince = idleAnchor
	r.IdleDays = max(0, daysBetween(idleAnchor, now))

	if now.After(r.EstimatedDelivery) {
		r.DaysOverSLA = max(0, daysBetween(r.EstimatedDelivery, now))
	}

	// Risk score (0-100).
	score := 0
	if r.DaysOverSLA > 0 {
		score += clamp(r.DaysOverSLA*15, 0, 60)
	}
	if r.IdleDays >= 4 {
		score += clamp((r.IdleDays-3)*12, 0, 40)
	}
	if score > 100 {
		score = 100
	}
	r.RiskScore = int16(score)

	// Categorize. Order matters.
	switch {
	case r.DaysOverSLA >= 2:
		r.Health = HealthBreached
	case r.DaysOverSLA >= 1:
		r.Health = HealthAtRisk
	case r.IdleDays >= 7:
		r.Health = HealthBreached
	case r.IdleDays >= 4:
		r.Health = HealthAtRisk
	case daysBetween(now, r.EstimatedDelivery) <= 1:
		r.Health = HealthAtRisk
	default:
		r.Health = HealthOnTrack
	}

	return r
}

// Apply writes the computed Health into the Shipment fields.
func Apply(ship *store.Shipment, r Result) {
	ship.Health = string(r.Health)
	ship.RiskScore = r.RiskScore
	if !r.IdleSince.IsZero() {
		t := r.IdleSince
		ship.IdleSince = &t
	}
	if !r.EstimatedDelivery.IsZero() && ship.EstimatedDelivery == nil {
		t := r.EstimatedDelivery
		ship.EstimatedDelivery = &t
	}
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func daysBetween(a, b time.Time) int {
	return int(b.Sub(a).Hours() / 24)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
