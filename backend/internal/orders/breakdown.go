package orders

import (
	"math"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/sla"
	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// StageBreakdown is one row of the rastreiaki "Análise por etapa" panel.
type StageBreakdown struct {
	Field               string     `json:"field"`
	Label               string     `json:"label"`
	TargetHours         float64    `json:"target_hours"`
	ActualHours         *float64   `json:"actual_hours,omitempty"`
	DelayHours          float64    `json:"delay_hours"`
	HoursToTarget       *float64   `json:"hours_to_target,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	TargetAt            time.Time  `json:"target_at"`
	IsOnTime            bool       `json:"is_on_time"`
	IsPending           bool       `json:"is_pending"`
	CascadeContribution float64    `json:"cascade_contribution"`
}

// Diagnosis summarises the delay story for the alert banner.
type Diagnosis struct {
	FirstDelayField   string  `json:"first_delay_field,omitempty"`
	FirstDelayLabel   string  `json:"first_delay_label,omitempty"`
	FirstDelayHours   float64 `json:"first_delay_hours"`
	WorstDelayField   string  `json:"worst_delay_field,omitempty"`
	WorstDelayLabel   string  `json:"worst_delay_label,omitempty"`
	WorstDelayHours   float64 `json:"worst_delay_hours"`
	TotalCascadeDelay float64 `json:"total_cascade_delay"`
}

// BreakdownResult is the response shape of /api/v1/orders/{id}/breakdown.
type BreakdownResult struct {
	Anchor    time.Time        `json:"anchor"`
	Stages    []StageBreakdown `json:"stages"`
	Diagnosis Diagnosis        `json:"diagnosis"`
}

// ComputeBreakdown replays the SLA policy against the persisted stage
// timestamps and returns the rastreiaki breakdown payload.
func ComputeBreakdown(ship *store.Shipment, now time.Time) BreakdownResult {
	policy := sla.PolicyFor(ship.Carrier, ship.Service)
	anchor := ship.CreatedAt
	if anchor.IsZero() {
		anchor = now
	}

	defs := []struct {
		Field, Label string
		CumHours     int
		Stamp        *time.Time
	}{
		{"label_issued_at", "Emissão de etiqueta", policy.Label, ship.LabelIssuedAt},
		{"preparing_at", "Preparação iniciada", policy.Prep, ship.PreparingAt},
		{"ready_for_pickup_at", "Pronto para coleta", policy.Ready, ship.ReadyForPickupAt},
		{"posted_at", "Postagem", policy.Posted, ship.PostedAt},
		{"out_for_delivery_at", "Saiu para entrega", policy.OFD, ship.OutForDeliveryAt},
		{"delivered_at", "Entregue", policy.Delivered, ship.DeliveredAt},
	}

	out := BreakdownResult{Anchor: anchor, Stages: make([]StageBreakdown, 0, len(defs))}
	cascadeDelay := 0.0

	for _, d := range defs {
		targetAt := anchor.Add(time.Duration(d.CumHours) * time.Hour)
		row := StageBreakdown{
			Field:       d.Field,
			Label:       d.Label,
			TargetHours: float64(d.CumHours),
			TargetAt:    targetAt,
		}

		if d.Stamp != nil && !d.Stamp.IsZero() {
			t := *d.Stamp
			// Clamp to >=0: stamps slightly before the anchor (e.g. WC order
			// created retroactively, clock skew) would render "—" via the
			// frontend's fmtHours, which negates the "Concluído em" label.
			actual := math.Max(0, t.Sub(anchor).Hours())
			row.ActualHours = ptrF(actual)
			row.DelayHours = math.Max(0, actual-float64(d.CumHours))
			row.IsOnTime = actual <= float64(d.CumHours)
			row.CompletedAt = &t
			if !row.IsOnTime {
				row.CascadeContribution = math.Max(0, row.DelayHours-cascadeDelay)
				if row.DelayHours > cascadeDelay {
					cascadeDelay = row.DelayHours
				}
			}
		} else {
			elapsed := now.Sub(anchor).Hours()
			row.IsPending = true
			h := float64(d.CumHours) - elapsed
			row.HoursToTarget = ptrF(h)
			if elapsed > float64(d.CumHours) {
				row.DelayHours = elapsed - float64(d.CumHours)
				row.IsOnTime = false
			} else {
				row.IsOnTime = true
			}
		}
		out.Stages = append(out.Stages, row)
	}

	// Diagnosis
	var firstSet bool
	var worst *StageBreakdown
	for i, s := range out.Stages {
		if s.IsPending || s.IsOnTime {
			continue
		}
		if !firstSet {
			out.Diagnosis.FirstDelayField = s.Field
			out.Diagnosis.FirstDelayLabel = s.Label
			out.Diagnosis.FirstDelayHours = s.DelayHours
			firstSet = true
		}
		if worst == nil || s.DelayHours > worst.DelayHours {
			worst = &out.Stages[i]
		}
		out.Diagnosis.TotalCascadeDelay += s.CascadeContribution
	}
	if worst != nil {
		out.Diagnosis.WorstDelayField = worst.Field
		out.Diagnosis.WorstDelayLabel = worst.Label
		out.Diagnosis.WorstDelayHours = worst.DelayHours
	}

	return out
}

func ptrF(f float64) *float64 { return &f }
