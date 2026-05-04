package sync

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
)

// TriggerSender is what the dispatcher needs from a notification channel.
// Kept narrow so a future SMS / Email channel can plug in without
// pulling the whole notifier package.
type TriggerSender interface {
	SendTextWith(ctx context.Context, session, phone, message string) error
}

// TriggerDispatcher fires the configured WhatsApp messages when a
// shipment crosses a milestone (postado, in_transit, delivered,
// breached). De-duplication is delegated to the notifications audit
// table — the dispatcher refuses to fire the same (order, event)
// twice within a window, regardless of how many syncs report the
// transition.
type TriggerDispatcher struct {
	Triggers *store.NotificationTriggers
	Audit    *store.Audit
	Orders   *store.Orders
	Sender   TriggerSender
	Log      *slog.Logger
}

// ShipmentSnapshot captures the fields that matter for transition
// detection. Pass `before` (snapshot the moment the row is loaded) and
// `after` (the mutated shipment about to be upserted) to OnShipmentSynced.
type ShipmentSnapshot struct {
	PostedAt    *time.Time
	InTransitAt *time.Time
	DeliveredAt *time.Time
	SLAState    string
}

// Snapshot extracts the relevant fields from a shipment so the caller
// can hold a pre-update copy for diffing.
func Snapshot(s *store.Shipment) ShipmentSnapshot {
	return ShipmentSnapshot{
		PostedAt:    cloneTime(s.PostedAt),
		InTransitAt: cloneTime(s.InTransitAt),
		DeliveredAt: cloneTime(s.DeliveredAt),
		SLAState:    s.SLAState,
	}
}

// OnShipmentSynced inspects the diff between `before` and the current
// shipment and fires every applicable trigger. Errors are logged and
// swallowed — a misbehaving WAHA must not break the Frenet sync loop.
func (d *TriggerDispatcher) OnShipmentSynced(ctx context.Context, before ShipmentSnapshot, ship *store.Shipment) {
	if d == nil || d.Triggers == nil {
		return
	}

	triggers, err := d.Triggers.EnabledByKey(ctx)
	if err != nil {
		d.Log.Warn("trigger dispatcher: load failed", "err", err)
		return
	}
	if len(triggers) == 0 {
		return
	}

	type firing struct {
		event string
	}
	var pending []firing
	if isNewlySet(before.PostedAt, ship.PostedAt) {
		pending = append(pending, firing{store.EventPostado})
	}
	if isNewlySet(before.InTransitAt, ship.InTransitAt) {
		pending = append(pending, firing{store.EventInTransit})
	}
	if isNewlySet(before.DeliveredAt, ship.DeliveredAt) {
		pending = append(pending, firing{store.EventDelivered})
	}
	// SLA breach: fired the first time we observe the transition into
	// BREACHED (or COMPLETED_LATE) for a shipment that wasn't already
	// in one of those states.
	if isNewBreach(before.SLAState, ship.SLAState) {
		pending = append(pending, firing{store.EventBreached})
	}

	if len(pending) == 0 {
		return
	}

	// Need the order to render the template + know the customer phone.
	dbOrder, err := d.Orders.GetByID(ctx, ship.OrderID)
	if err != nil {
		d.Log.Warn("trigger dispatcher: order lookup failed", "order_id", ship.OrderID, "err", err)
		return
	}
	if strings.TrimSpace(dbOrder.CustomerPhone) == "" {
		// Silent skip — the dashboard surfaces "no phone" already.
		return
	}

	for _, f := range pending {
		t, ok := triggers[f.event]
		if !ok {
			continue
		}
		// De-dup: skip if we already fired this event for this order
		// within the cooldown window (or ever, if cooldown=0).
		fired, err := d.Audit.HasNotificationSince(ctx, dbOrder.ID, "trigger:"+f.event,
			time.Now().Add(-time.Duration(t.CooldownMinutes)*time.Minute))
		if err != nil {
			d.Log.Warn("trigger dispatcher: dedup check failed", "err", err)
			continue
		}
		if fired {
			continue
		}

		message := renderTriggerTemplate(t.Template, dbOrder, ship)
		err = d.Sender.SendTextWith(ctx, t.Session, dbOrder.CustomerPhone, message)
		rec := store.Notification{
			OrderID:  dbOrder.ID,
			Channel:  "waha",
			Template: "trigger:" + f.event,
			Status:   "sent",
		}
		if err != nil {
			rec.Status = "failed"
			rec.Error = err.Error()
			d.Log.Warn("trigger dispatcher: send failed",
				"order_id", dbOrder.ID, "event", f.event, "err", err)
		} else {
			d.Log.Info("trigger fired",
				"order_id", dbOrder.ID, "event", f.event)
		}
		_ = d.Audit.RecordNotification(ctx, rec)
	}
}

// renderTriggerTemplate replaces the documented placeholders in the
// trigger template. Unknown placeholders are left untouched so a typo
// is visible to ops instead of silently producing empty text.
func renderTriggerTemplate(tpl string, o *store.Order, s *store.Shipment) string {
	first := strings.TrimSpace(strings.SplitN(o.CustomerName, " ", 2)[0])
	if first == "" {
		first = "cliente"
	}
	last := s.LastEvent
	if last == "" {
		last = "—"
	}
	eta := "em breve"
	if s.EstimatedDelivery != nil && !s.EstimatedDelivery.IsZero() {
		eta = s.EstimatedDelivery.In(brTZ()).Format("02/01/2006")
	}
	tracking := s.TrackingCode
	if tracking == "" {
		tracking = "—"
	}
	url := s.TrackingURL
	if url == "" {
		url = "(em breve)"
	}
	repls := []string{
		"{first_name}", first,
		"{customer_name}", o.CustomerName,
		"{order_id}", fmt.Sprintf("%d", o.WCOrderID),
		"{tracking}", tracking,
		"{track_url}", url,
		"{last_event}", last,
		"{eta}", eta,
		"{carrier}", s.Carrier,
	}
	r := strings.NewReplacer(repls...)
	return r.Replace(tpl)
}

func brTZ() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.UTC
	}
	return loc
}

func isNewlySet(before, after *time.Time) bool {
	had := before != nil && !before.IsZero()
	has := after != nil && !after.IsZero()
	return !had && has
}

func isNewBreach(before, after string) bool {
	wasBreach := before == "BREACHED" || before == "COMPLETED_LATE"
	isBreach := after == "BREACHED" || after == "COMPLETED_LATE"
	return !wasBreach && isBreach
}

func cloneTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	c := *t
	return &c
}
