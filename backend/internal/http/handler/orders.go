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
	"sync"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/frenet"
	"github.com/univerbeauty777/univer-tracker/backend/internal/orders"
	"github.com/univerbeauty777/univer-tracker/backend/internal/woocommerce"
)

// Orders wires the WooCommerce and Frenet clients to HTTP routes.
type Orders struct {
	WC     *woocommerce.Client
	Frenet *frenet.Client
	Log    *slog.Logger
}

// orderListItem is a slim projection used by the table view.
type orderListItem struct {
	ID            int64               `json:"id"`
	Status        string              `json:"status"`
	StatusLabel   string              `json:"status_label"`
	CustomerName  string              `json:"customer_name"`
	CustomerCity  string              `json:"customer_city"`
	CustomerState string              `json:"customer_state"`
	Total         string              `json:"total"`
	CreatedAt     time.Time           `json:"created_at"`
	PaidAt        *time.Time          `json:"paid_at,omitempty"`
	Tracking      orders.TrackingInfo `json:"tracking"`
}

// orderDetail is the rich projection returned by GET /orders/:id — adds line items,
// addresses and (when available) the live Frenet event timeline.
type orderDetail struct {
	orderListItem
	Email      string                  `json:"email"`
	Phone      string                  `json:"phone"`
	LineItems  []woocommerce.LineItem  `json:"line_items"`
	Shipping   woocommerce.Address     `json:"shipping"`
	Billing    woocommerce.Address     `json:"billing"`
	ShippingMethod string              `json:"shipping_method,omitempty"`
}

// List handles GET /api/v1/orders — proxies WooCommerce with optional status filter.
func (h *Orders) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	q := r.URL.Query()
	params := woocommerce.ListOrdersParams{
		PerPage: parseInt(q.Get("per_page"), 50, 200),
		Page:    parseInt(q.Get("page"), 1, 0),
	}
	if s := q.Get("status"); s != "" {
		params.Status = strings.Split(s, ",")
	}
	if d := q.Get("after"); d != "" {
		if t, err := time.Parse(time.RFC3339, d); err == nil {
			params.After = t
		}
	}

	wcOrders, err := h.WC.ListOrders(ctx, params)
	if err != nil {
		h.Log.Error("list orders failed", "err", err)
		writeError(w, http.StatusBadGateway, "could not reach woocommerce")
		return
	}

	out := make([]orderListItem, 0, len(wcOrders))
	for i := range wcOrders {
		out = append(out, projectListItem(&wcOrders[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"orders": out,
		"page":   params.Page,
		"count":  len(out),
	})
}

// Get handles GET /api/v1/orders/{id} — returns a single order enriched with
// Frenet tracking events when a tracking number is known.
func (h *Orders) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	wcOrder, err := h.WC.GetOrder(ctx, id)
	if err != nil {
		h.Log.Error("get order failed", "err", err, "order_id", id)
		writeError(w, http.StatusBadGateway, "could not reach woocommerce")
		return
	}

	detail := projectDetail(wcOrder)

	// Enrich with live Frenet events in parallel — tracking is best-effort,
	// don't fail the whole request if Frenet is down.
	if detail.Tracking.HasTracking() && h.Frenet != nil {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			tctx, tcancel := context.WithTimeout(ctx, 10*time.Second)
			defer tcancel()
			info, err := h.Frenet.GetTrackingInfo(tctx, detail.Tracking.Number, detail.Tracking.ServiceCode)
			if err != nil {
				h.Log.Warn("frenet enrichment failed", "err", err, "order_id", id, "tracking", detail.Tracking.Number)
				return
			}
			detail.Tracking.Events = info.TrackingEvents
			if info.TrackingURL != "" {
				detail.Tracking.URL = info.TrackingURL
			}
			if status := classify(info.TrackingEvents); status != frenet.StatusUnknown {
				detail.Tracking.Status = status
				detail.Tracking.StatusLabel = orders.StatusLabel(status)
			}
		}()
		wg.Wait()
	}

	writeJSON(w, http.StatusOK, detail)
}

// UpdateStatus handles PATCH /api/v1/orders/{id}/status — pushes a new status
// to WooCommerce and returns the refreshed order.
func (h *Orders) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
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
	if body.Status == "" {
		writeError(w, http.StatusUnprocessableEntity, "status is required")
		return
	}
	body.Status = strings.TrimPrefix(body.Status, "wc-")

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	if err := h.WC.UpdateOrderStatus(ctx, id, body.Status); err != nil {
		h.Log.Error("update status failed", "err", err, "order_id", id, "status", body.Status)
		writeError(w, http.StatusBadGateway, "could not update woocommerce order")
		return
	}

	if body.Note != "" {
		if err := h.WC.AddOrderNote(ctx, id, body.Note, false); err != nil {
			h.Log.Warn("add note failed", "err", err, "order_id", id)
		}
	}

	wcOrder, err := h.WC.GetOrder(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": body.Status})
		return
	}
	writeJSON(w, http.StatusOK, projectDetail(wcOrder))
}

// classify takes the freshest event from Frenet and returns the normalized
// status. We trust the most recent event in the list (Frenet returns DESC).
func classify(events []frenet.TrackingEvent) frenet.Status {
	for _, e := range events {
		if s := frenet.MapEvent(e.EventDescription); s != frenet.StatusUnknown {
			return s
		}
	}
	return frenet.StatusUnknown
}

func projectListItem(o *woocommerce.Order) orderListItem {
	li := orderListItem{
		ID:            o.ID,
		Status:        o.Status,
		StatusLabel:   wcStatusLabel(o.Status),
		CustomerName:  strings.TrimSpace(o.Shipping.FirstName + " " + o.Shipping.LastName),
		CustomerCity:  o.Shipping.City,
		CustomerState: o.Shipping.State,
		Total:         o.Total,
		CreatedAt:     o.DateCreatedGMT.Time,
		Tracking:      orders.FromOrder(o),
	}
	if li.CustomerName == "" {
		li.CustomerName = strings.TrimSpace(o.Billing.FirstName + " " + o.Billing.LastName)
	}
	if li.CustomerCity == "" {
		li.CustomerCity = o.Billing.City
		li.CustomerState = o.Billing.State
	}
	if !o.DatePaidGMT.Time.IsZero() {
		t := o.DatePaidGMT.Time
		li.PaidAt = &t
	}
	return li
}

func projectDetail(o *woocommerce.Order) orderDetail {
	d := orderDetail{
		orderListItem: projectListItem(o),
		Email:         o.Billing.Email,
		Phone:         o.Billing.Phone,
		LineItems:     o.LineItems,
		Shipping:      o.Shipping,
		Billing:       o.Billing,
	}
	if len(o.ShippingLines) > 0 {
		d.ShippingMethod = o.ShippingLines[0].MethodTitle
	}
	return d
}

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
	case "shipped":
		return "Enviado"
	case "in-transit":
		return "Em trânsito"
	case "out-for-delivery":
		return "Saiu para entrega"
	default:
		return s
	}
}

func parseInt(s string, def, max int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return def
	}
	if max > 0 && n > max {
		return max
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

// ErrNotConfigured is returned when a required client (e.g. WC) is unset.
var ErrNotConfigured = errors.New("client not configured")
