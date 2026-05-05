package sla

import (
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// State is the rastreiaki-style classification.
type State string

const (
	StateOnTrack       State = "ON_TRACK"
	StateAtRisk        State = "AT_RISK"
	StateBreached      State = "BREACHED"
	StateCompleted     State = "COMPLETED"
	StateCompletedLate State = "COMPLETED_LATE"
)

// EvalResult mirrors the rastreiaki shipment SLA fields.
type EvalResult struct {
	State         State
	BreachedStage string // store column name when breached, otherwise ""
	EstimatedAt   time.Time
}

// Evaluate computes sla_state and sla_breached_stage for a shipment given
// its anchor (the moment we linked the tracking) and the carrier policy.
//
// The first stage whose stamp is later than its cumulative SLA marks the
// breach. If no stamp violated but the total deadline has passed without
// delivery, the whole shipment is breached. AT_RISK is anything where
// >= 80% of the total budget has elapsed but no breach yet.
func Evaluate(ship *store.Shipment, anchor time.Time, now time.Time) EvalResult {
	policy := PolicyFor(ship.Carrier, ship.Service)
	deliveredAt := timeOrNil(ship.DeliveredAt)
	estimated := anchor.Add(time.Duration(policy.Delivered) * time.Hour)

	r := EvalResult{State: StateOnTrack, EstimatedAt: estimated}

	type stageCheck struct {
		Field    string
		At       *time.Time
		Deadline time.Time
	}
	checks := []stageCheck{
		{"label_issued_at", ship.LabelIssuedAt, anchor.Add(time.Duration(policy.Label) * time.Hour)},
		{"preparing_at", ship.PreparingAt, anchor.Add(time.Duration(policy.Prep) * time.Hour)},
		{"ready_for_pickup_at", ship.ReadyForPickupAt, anchor.Add(time.Duration(policy.Ready) * time.Hour)},
		{"posted_at", ship.PostedAt, anchor.Add(time.Duration(policy.Posted) * time.Hour)},
		{"out_for_delivery_at", ship.OutForDeliveryAt, anchor.Add(time.Duration(policy.OFD) * time.Hour)},
		{"delivered_at", ship.DeliveredAt, anchor.Add(time.Duration(policy.Delivered) * time.Hour)},
	}

	for _, c := range checks {
		if c.At != nil && c.At.After(c.Deadline) {
			r.State = StateBreached
			r.BreachedStage = c.Field
			break
		}
	}

	if deliveredAt != nil {
		if r.State == StateBreached || deliveredAt.After(estimated) {
			r.State = StateCompletedLate
			if r.BreachedStage == "" {
				r.BreachedStage = "delivered_at"
			}
		} else {
			r.State = StateCompleted
			r.BreachedStage = ""
		}
		return r
	}

	// Not delivered yet — check total deadline + at-risk band.
	if r.State != StateBreached {
		if now.After(estimated) {
			r.State = StateBreached
			r.BreachedStage = "delivered_at"
		} else {
			elapsed := now.Sub(anchor)
			budget := estimated.Sub(anchor)
			if budget > 0 && float64(elapsed)/float64(budget) >= 0.8 {
				r.State = StateAtRisk
			}
		}
	}

	return r
}

func timeOrNil(t *time.Time) *time.Time {
	if t == nil || t.IsZero() {
		return nil
	}
	return t
}
