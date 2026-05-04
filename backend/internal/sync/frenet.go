package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/frenet"
	"github.com/univerbeauty777/univer-tracker/backend/internal/integrations"
	"github.com/univerbeauty777/univer-tracker/backend/internal/sla"
	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// Frenet pulls events for active shipments and persists them. The client
// is resolved fresh on every Run so dashboard credential changes take
// effect on the next tick without a worker restart.
type Frenet struct {
	Shipments    *store.Shipments
	Events       *store.Events
	Integrations *integrations.Resolver
	BatchSize    int
	Dispatcher   *TriggerDispatcher // optional: nil disables WhatsApp triggers
	Log          *slog.Logger
}

// Run picks the shipments with the oldest last_synced_at and refreshes them
// from Frenet. Per-shipment errors are logged and skipped — one slow carrier
// must not stall the whole batch.
func (s *Frenet) Run(ctx context.Context) (Stats, error) {
	stats := Stats{Started: time.Now()}

	client, err := s.Integrations.Frenet(ctx)
	if err != nil {
		s.Log.Warn("frenet sync skipped: not configured", "err", err)
		stats.Finished = time.Now()
		return stats, nil
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
		if err := s.refreshOne(ctx, client, &active[i]); err != nil {
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

func (s *Frenet) refreshOne(ctx context.Context, client *frenet.Client, ship *store.Shipment) error {
	rctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	resp, err := client.GetTrackingInfo(rctx, ship.TrackingCode, ship.ServiceCode)
	if err != nil {
		return err
	}
	if resp.ErrorMessage != "" {
		return fmt.Errorf("frenet error: %s", resp.ErrorMessage)
	}

	// Snapshot before mutation so the trigger dispatcher can detect
	// transitions (Postado, Em trânsito, Entregue, Atrasado).
	before := Snapshot(ship)

	now := time.Now().UTC()
	ship.LastSyncedAt = &now
	if resp.TrackingURL != "" && ship.TrackingURL == "" {
		ship.TrackingURL = resp.TrackingURL
	}
	if resp.ServiceDescription != "" && ship.Service == "" {
		ship.Service = resp.ServiceDescription
	}

	// Translate Frenet events → DB rows, populate per-stage timestamps,
	// and capture latest description for the badge.
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
			Location:    cleanLocation(e.EventLocation),
			Type:        strings.ToLower(strings.TrimSpace(e.EventType)),
			Raw:         raw,
		})
		if occ.After(newest) {
			newest = occ
			latestDescr = e.EventDescription
			latestStatus = frenet.MapEvent(e.EventDescription)
		}

		// Per-stage timestamp: keep the earliest occurrence. A single
		// event can hit multiple stages — see MapEventToStages docs.
		for _, stage := range frenet.MapEventToStages(e.EventDescription) {
			applyStageStamp(ship, stage, occ)
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
	}

	anchor := ship.CreatedAt
	if anchor.IsZero() {
		anchor = now
	}
	eval := sla.Evaluate(ship, anchor, now)
	ship.SLAState = string(eval.State)
	ship.SLABreachedStage = eval.BreachedStage
	if ship.EstimatedDelivery == nil {
		t := eval.EstimatedAt
		ship.EstimatedDelivery = &t
	}
	sla.Apply(ship, sla.Compute(ship, now))

	if _, err := s.Shipments.Upsert(ctx, ship); err != nil {
		return fmt.Errorf("upsert shipment: %w", err)
	}

	// Fire any configured WhatsApp triggers for milestones the
	// shipment just crossed. Errors are swallowed by the dispatcher.
	if s.Dispatcher != nil {
		s.Dispatcher.OnShipmentSynced(ctx, before, ship)
	}
	return nil
}

// applyStageStamp keeps the earliest known timestamp for a stage. Frenet
// sometimes emits the same milestone twice (initial scan + retry), so we
// always trust the first one.
func applyStageStamp(ship *store.Shipment, stage string, t time.Time) {
	earlier := func(cur *time.Time, next time.Time) *time.Time {
		if cur == nil || cur.IsZero() || next.Before(*cur) {
			tt := next
			return &tt
		}
		return cur
	}
	switch stage {
	case "label_issued_at":
		ship.LabelIssuedAt = earlier(ship.LabelIssuedAt, t)
	case "preparing_at":
		ship.PreparingAt = earlier(ship.PreparingAt, t)
	case "ready_for_pickup_at":
		ship.ReadyForPickupAt = earlier(ship.ReadyForPickupAt, t)
	case "posted_at":
		ship.PostedAt = earlier(ship.PostedAt, t)
	case "in_transit_at":
		ship.InTransitAt = earlier(ship.InTransitAt, t)
	case "at_destination_city_at":
		ship.AtDestinationCityAt = earlier(ship.AtDestinationCityAt, t)
	case "out_for_delivery_at":
		ship.OutForDeliveryAt = earlier(ship.OutForDeliveryAt, t)
	case "delivered_at":
		ship.DeliveredAt = earlier(ship.DeliveredAt, t)
	}
}

// cleanLocation strips the "-UF-BR" / "-BR" prefixes Frenet emits when
// the city/state are unknown so the dashboard shows blank instead of
// a misleading dash.
func cleanLocation(s string) string {
	t := strings.TrimSpace(s)
	t = strings.TrimPrefix(t, "-")
	t = strings.TrimSpace(t)
	if t == "BR" {
		return ""
	}
	return t
}

// parseFrenetTime accepts the formats Frenet actually returns in the
// wild. The official docs only mention "dd/MM/yyyy HH:mm:ss" but the
// production API also emits "dd/MM/yyyy HH:mm" (no seconds) for newly
// created labels — every single event on a fresh shipment lands in
// that shape, so missing it means the timeline is permanently empty.
func parseFrenetTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"2006-01-02 15:04",
		"02/01/2006",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
