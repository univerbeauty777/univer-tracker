package sla

import (
	"testing"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

func TestDays(t *testing.T) {
	cases := []struct {
		carrier, service string
		want             int
	}{
		{"Correios", "SEDEX - Contrato", 4},
		{"Correios", "PAC", 8},
		{"correios", "", 6},
		{"Jadlog", "Package", 5},
		{"Loggi", "Express", 3},
		{"unknown carrier", "", 7},
	}
	for _, c := range cases {
		got := Days(c.carrier, c.service)
		if got != c.want {
			t.Errorf("Days(%q, %q) = %d; want %d", c.carrier, c.service, got, c.want)
		}
	}
}

func TestCompute(t *testing.T) {
	now := time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		ship store.Shipment
		want Health
	}{
		{
			name: "delivered short-circuits",
			ship: store.Shipment{
				DeliveredAt: ptr(now.AddDate(0, 0, -2)),
			},
			want: HealthDelivered,
		},
		{
			name: "fresh PAC, no events yet, on track",
			ship: store.Shipment{
				Carrier:   "Correios",
				Service:   "PAC",
				CreatedAt: now.AddDate(0, 0, -2),
			},
			want: HealthOnTrack,
		},
		{
			name: "PAC near ETA, at risk",
			ship: store.Shipment{
				Carrier:     "Correios",
				Service:     "PAC",
				CreatedAt:   now.AddDate(0, 0, -8),
				LastEventAt: ptr(now.AddDate(0, 0, -1)),
			},
			want: HealthAtRisk,
		},
		{
			name: "Jadlog 3 days past ETA, breached",
			ship: store.Shipment{
				Carrier:     "Jadlog",
				Service:     "",
				CreatedAt:   now.AddDate(0, 0, -8),
				LastEventAt: ptr(now.AddDate(0, 0, -2)),
			},
			want: HealthBreached,
		},
		{
			name: "fresh shipment idle 5 days, at risk",
			ship: store.Shipment{
				Carrier:     "Correios",
				Service:     "PAC",
				CreatedAt:   now.AddDate(0, 0, -5),
				LastEventAt: ptr(now.AddDate(0, 0, -5)),
			},
			want: HealthAtRisk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute(&tt.ship, now)
			if got.Health != tt.want {
				t.Errorf("Compute() health = %v; want %v (result=%+v)", got.Health, tt.want, got)
			}
		})
	}
}

func ptr(t time.Time) *time.Time { return &t }
