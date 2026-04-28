package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/frenet"
	"github.com/univerbeauty777/univer-tracker/backend/internal/sla"
	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// Frenet pulls events for active shipments and persists them.
type Frenet struct {
	Shipments *store.Shipments
	Events    *store.Events
	Client    *frenet.Client
	BatchSize int
	Log       *slog.Logger
}

// Run picks the shipments with the oldest last_synced_at and refreshes them
// from Frenet. Per-shipment errors are logged and skipped — one slow carrier
// must not stall the whole batch.
func (s *Frenet) Run(ctx context.Context) (Stats, error) {
	stats := Stats{Started: time.Now()}

	if s.Client == nil {
		return stats, fmt.Errorf("frenet client not configured")
	}

	limit := s.BatchSize
	if limit <= 0 {
		limit = 50
	}

	active, err := s.Shipments.ListActive(ctx, limit)
	if err != nil {
		return stats, fmt.Errorf("list active: %w", err)
	}

	for i := range active {
		if err := s.refreshOne(ctx, &active[i]); err != nil {
			stats.Errors++
			s.Log.Warn("refresh shipment failed",
				"shipment_id", active[i].ID,
				"tracking", active[i].TrackingCode,
				"err", err)
			continue
		}
		stats.Synced++

		// Be polite with the Frenet API.
		select {
		case <-ctx.Done():
			return stats, ctx.Err()
		case <-time.After(150 * time.Millisecond):
		}
	}

	stats.Finished = time.Now()
	return stats, nil
}

func (s *Frenet) refreshOne(ctx context.Context, ship *store.Shipment) error {
	rctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	resp, err := s.Client.GetTrackingInfo(rctx, ship.TrackingCode, ship.ServiceCode)
	if err != nil {
		return err
	}
	if resp.ErrorMessage != "" {
		return fmt.Errorf("frenet error: %s", resp.ErrorMessage)
	}

	now := time.Now().UTC()
	ship.LastSyncedAt = &now
	if resp.TrackingURL != "" && ship.TrackingURL == "" {
		ship.TrackingURL = resp.TrackingURL
	}
	if resp.ServiceDescription != "" && ship.Service == "" {
		ship.Service = resp.ServiceDescription
	}

	// Translate Frenet events → DB rows and capture latest description.
	dbEvents := make([]store.Event, 0, len(resp.TrackingEvents))
	var newest time.Time
	var latestDescr string
	var latestStatus frenet.Status
	for _, e := range resp.TrackingEvents {
		occ := parseFrenetTime(e.EventDateTime)
		if occ.IsZero() {
			continue
		}
		raw, _ := json.Marshal(e)
		dbEvents = append(dbEvents, store.Event{
			ShipmentID:  ship.ID,
			OccurredAt:  occ,
			Description: e.EventDescription,
			Location:    e.EventLocation,
			Type:        strings.ToLower(strings.TrimSpace(e.EventType)),
			Raw:         raw,
		})
		if occ.After(newest) {
			newest = occ
			latestDescr = e.EventDescription
			latestStatus = frenet.MapEvent(e.EventDescription)
		}
	}

	if len(dbEvents) > 0 {
		if _, err := s.Events.InsertMany(ctx, dbEvents); err != nil {
			return fmt.Errorf("insert events: %w", err)
		}
	}

	if !newest.IsZero() {
		ship.LastEvent = latestDescr
		ship.LastEventAt = &newest
	}
	if latestStatus != frenet.StatusUnknown {
		ship.Status = string(latestStatus)
		if latestStatus == frenet.StatusDelivered && ship.DeliveredAt == nil {
			t := newest
			ship.DeliveredAt = &t
		}
	}

	sla.Apply(ship, sla.Compute(ship, time.Now().UTC()))

	if _, err := s.Shipments.Upsert(ctx, ship); err != nil {
		return fmt.Errorf("upsert shipment: %w", err)
	}
	return nil
}

// parseFrenetTime accepts the two formats Frenet returns:
//   - "2026-01-15 10:30:00"  (ISO-ish, UTC-naive)
//   - "15/01/2026 10:30:00"  (BR locale)
func parseFrenetTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"02/01/2006 15:04:05",
		"02/01/2006",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
