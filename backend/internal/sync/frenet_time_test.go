package sync

import (
	"testing"
	"time"
)

func TestParseFrenetTime(t *testing.T) {
	cases := []struct {
		in   string
		want time.Time
	}{
		// Production format observed for newly emitted labels
		// (no seconds): "Etiqueta emitida ... 28/04/2026 12:38".
		{"28/04/2026 12:38", time.Date(2026, 4, 28, 12, 38, 0, 0, time.UTC)},
		// Documented BR format with seconds.
		{"15/01/2026 10:30:00", time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)},
		// ISO-ish (older shipments).
		{"2026-01-15 10:30:00", time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)},
		// Date only.
		{"15/01/2026", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
		// Garbage stays zero.
		{"", time.Time{}},
		{"not a date", time.Time{}},
	}
	for _, c := range cases {
		got := parseFrenetTime(c.in)
		if !got.Equal(c.want) {
			t.Errorf("parseFrenetTime(%q) = %v; want %v", c.in, got, c.want)
		}
	}
}
