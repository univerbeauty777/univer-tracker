// Package handler holds the HTTP handlers for the public REST API.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/frenet"
	"github.com/univerbeauty777/univer-tracker/backend/internal/integrations"
	"github.com/univerbeauty777/univer-tracker/backend/internal/orders"
	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
	"github.com/univerbeauty777/univer-tracker/backend/internal/woocommerce"
)

// Orders serves the public Orders REST API.
type Orders struct {
	StoreID     int64
	Orders      *store.Orders
	Shipments   *store.Shipments
	Events      *store.Events
	Integrations *integrations.Resolver
	Log         *slog.Logger
}

// orderListItem is the slim shape used by the table view.
type orderListItem struct {
	ID            int64       `json:"id"`
	WCOrderID     int64       `json:"wc_order_id"`
	Status        string      `json:"status"`
	StatusLabel   string      `json:"status_label"`
	CustomerName  string      `json:"customer_name"`
	CustomerCity  string      `json:"customer_city"`
	CustomerState string      `json:"customer_state"`
	Total         float64     `json:"total"`
	CreatedAt    time.Time    `json:"created_at"`
	PaidAt       *time.Time   `json:"paid_at,omitempty"`
	Tracking     trackingView `json:"tracking"`
}

type trackingView struct {
	Number       string             `json:"number"`
	Carrier      string             `json:"carrier"`
	Service      string             `json:"service,omitempty"`
	ServiceCode  string             `json:"service_code,omitempty"`
	URL          string             `json:"url,omitempty"`
	Status       frenet.Status      `json:"status"`
	StatusLabel  string             `json:"status_label"`
	Health       string             `json:"health"`
	HealthLabel  string             `json:"health_label"`
	LastEvent    string             `json:"last_event,omitempty"`
	LastEventAt  *time.Time         `json:"last_event_at,omitempty"`
	EstDelivery  *time.Time         `json:"estimated_delivery,omitempty"`
	DeliveredAt  *time.Time         `json:"delivered_at,omitempty"`
	IdleSince    *time.Time         `json:"idle_since,omitempty"`
	RiskScore    int16              `json:"risk_score"`
	Events       []timelineEvent    `json:"events,omitempty"`
}

type timelineEvent struct {
	OccurredAt  time.Time `json:"occurred_at"`
	Description string    `json:"description"`
	Location    string    `json:"location,omitempty"`
	Type        string    `json:"type,omitempty"`
}

// orderDetail extends the list item with the full WooCommerce projection
// so the detail page can show line items and addresses without a 2nd hit.
type orderDetail struct {
	orderListItem
	Email          string                 `json:"email"`
	Phone          string                 `json:"phone"`
	ShippingMethod string                 `json:"shipping_method,omitempty"`
	LineItems      []woocommerce.LineItem `json:"line_items"`
	Shipping       woocommerce.Address    `json:"shipping"`
	Billing        woocommerce.Address    `json:"billing"`
}

// List handles GET /api/v1/orders.
func (h *Orders) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	q := r.URL.Query()
	filters := store.ListFilters{
		StoreID: h.StoreID,
		Limit:   parseInt(q.Get("per_page"), 100, 200),
		Offset:  parseInt(q.Get("offset"), 0, 0),
		Search:  strings.TrimSpace(q.Get("q")),
	}
	if s := q.Get("status"); s != "" {
		filters.Statuses = strings.Split(s, ",")
	}
	if hf := q.Get("health"); hf != "" {
		filters.Health = strings.Split(hf, ",")
	}

	rows, err := h.Orders.List(ctx, filters)
	if err != nil {
		h.Log.Error("list orders failed", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load orders")
		return
	}

	out := make([]orderListItem, 0, len(rows))
	for i := range rows {
		out = append(out, projectListItem(&rows[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"orders": out,
		"count":  len(out),
	})
}

// Get handles GET /api/v1/orders/{id} where id is the WooCommerce order id.
// We reach back to WC for line items + addresses (which we don't store) and
// load the persisted shipment + events from Postgres.
func (h *Orders) Get(w http.ResponseWriter, r *http.Request) {
	wcID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	dbOrder, err := h.Orders.GetByWCID(ctx, h.StoreID, wcID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		h.Log.Error("get order failed", "wc_order_id", wcID, "err", err)
		writeError(w, http.StatusInternalServerError, "could not load order")
		return
	}

	var wcOrder *woocommerce.Order
	if wc, err := h.Integrations.WooCommerce(ctx); err == nil {
		o, err := wc.GetOrder(ctx, wcID)
		if err != nil {
			h.Log.Warn("wc get order failed", "wc_order_id", wcID, "err", err)
		}
		wcOrder = o
	} else {
		h.Log.Warn("wc resolver", "err", err)
	}

	ships, err := h.Shipments.ListByOrder(ctx, dbOrder.ID)
	if err != nil {
		h.Log.Warn("list shipments failed", "order_id", dbOrder.ID, "err", err)
	}
	var primary *store.Shipment
	if len(ships) > 0 {
		primary = &ships[0]
	}

	row := &store.OrderRow{Order: *dbOrder, Shipment: primary}
	detail := orderDetail{orderListItem: projectListItem(row)}

	if primary != nil {
		evts, err := h.Events.ListByShipment(ctx, primary.ID)
		if err == nil {
			detail.Tracking.Events = projectEvents(evts)
		}
	}

	if wcOrder != nil {
		detail.Email = wcOrder.Billing.Email
		detail.Phone = wcOrder.Billing.Phone
		detail.LineItems = wcOrder.LineItems
		detail.Shipping = wcOrder.Shipping
		detail.Billing = wcOrder.Billing
		if len(wcOrder.ShippingLines) > 0 {
			detail.ShippingMethod = wcOrder.ShippingLines[0].MethodTitle
		}
	}

	writeJSON(w, http.StatusOK, detail)
}

// UpdateStatus handles PATCH /api/v1/orders/{id}/status.
func (h *Orders) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	wcID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var body struct {
		Status string `json:"status"`
		Note   string `json:"note,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	body.Status = strings.TrimSpace(body.Status)
	body.Status = strings.TrimPrefix(body.Status, "wc-")
	if body.Status == "" {
		writeError(w, http.StatusUnprocessableEntity, "status is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	wc, err := h.Integrations.WooCommerce(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "woocommerce não configurada — abra Configurações")
		return
	}

	if err := wc.UpdateOrderStatus(ctx, wcID, body.Status); err != nil {
		h.Log.Error("wc update status failed", "wc_order_id", wcID, "err", err)
		writeError(w, http.StatusBadGateway, "could not update woocommerce")
		return
	}
	if body.Note != "" {
		if err := wc.AddOrderNote(ctx, wcID, body.Note, false); err != nil {
			h.Log.Warn("wc add note failed", "wc_order_id", wcID, "err", err)
		}
	}

	if dbOrder, err := h.Orders.GetByWCID(ctx, h.StoreID, wcID); err == nil {
		_ = h.Orders.UpdateStatus(ctx, dbOrder.ID, body.Status)
	}

	writeJSON(w, http.StatusOK, map[string]any{"id": wcID, "status": body.Status})
}

func projectListItem(row *store.OrderRow) orderListItem {
	o := row.Order
	li := orderListItem{
		ID:            o.WCOrderID,
		WCOrderID:     o.WCOrderID,
		Status:        o.Status,
		StatusLabel:   wcStatusLabel(o.Status),
		CustomerName:  o.CustomerName,
		CustomerCity:  o.CustomerCity,
		CustomerState: o.CustomerUF,
		Total:         o.TotalBRL,
		CreatedAt:     o.CreatedAt,
		PaidAt:        o.PaidAt,
		Tracking: trackingView{
			Status:      frenet.StatusUnknown,
			StatusLabel: orders.StatusLabel(frenet.StatusUnknown),
			Health:      "unknown",
			HealthLabel: healthLabel("unknown"),
		},
	}
	if row.Shipment != nil {
		s := row.Shipment
		li.Tracking.Number = s.TrackingCode
		li.Tracking.Carrier = s.Carrier
		li.Tracking.Service = s.Service
		li.Tracking.ServiceCode = s.ServiceCode
		li.Tracking.URL = s.TrackingURL
		st := frenet.Status(s.Status)
		li.Tracking.Status = st
		li.Tracking.StatusLabel = orders.StatusLabel(st)
		li.Tracking.Health = s.Health
		li.Tracking.HealthLabel = healthLabel(s.Health)
		li.Tracking.LastEvent = s.LastEvent
		li.Tracking.LastEventAt = s.LastEventAt
		li.Tracking.EstDelivery = s.EstimatedDelivery
		li.Tracking.DeliveredAt = s.DeliveredAt
		li.Tracking.IdleSince = s.IdleSince
		li.Tracking.RiskScore = s.RiskScore
	}
	return li
}

func projectEvents(evts []store.Event) []timelineEvent {
	out := make([]timelineEvent, 0, len(evts))
	for _, e := range evts {
		out = append(out, timelineEvent{
			OccurredAt:  e.OccurredAt,
			Description: e.Description,
			Location:    e.Location,
			Type:        e.Type,
		})
	}
	return out
}

// wcStatusLabel translates a WooCommerce status slug into a Portuguese label.
// Covers native statuses, the official UniverTracking custom ones (shipped /
// in-transit / out-for-delivery) and the custom slugs Lizzon already uses
// in production (separacao, enviado, em-rota, entregue, retornado).
func wcStatusLabel(s string) string {
	switch s {
	case "pending":
		return "Aguardando pagamento"
	case "processing":
		return "Processando"
	case "on-hold":
		return "Em espera"
	case "completed":
		return "Concluído"
	case "cancelled":
		return "Cancelado"
	case "refunded":
		return "Estornado"
	case "failed":
		return "Falhou"
	case "separacao":
		return "Em separação"
	case "aguardando":
		return "Aguardando"
	case "enviado", "shipped":
		return "Enviado"
	case "in-transit", "em-transito":
		return "Em trânsito"
	case "out-for-delivery", "em-rota":
		return "Saiu para entrega"
	case "entregue":
		return "Entregue"
	case "retornado":
		return "Retornado"
	default:
		// Fall back to a Title Case version of the slug.
		if s == "" {
			return ""
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}
}

func healthLabel(h string) string {
	switch h {
	case "on_track":
		return "No prazo"
	case "at_risk":
		return "Em risco"
	case "breached":
		return "SLA quebrado"
	default:
		return "Sem dados"
	}
}

func parseInt(s string, def, max int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	if max > 0 && n > max {
		return max
	}
	if n < 0 {
		return def
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}
