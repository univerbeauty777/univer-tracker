// Package handler holds the HTTP handlers for the public REST API.
package handler

import (
	"context"
	"encoding/csv"
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
	StoreID      int64
	Orders       *store.Orders
	Shipments    *store.Shipments
	Events       *store.Events
	Facets       *store.Facets
	Audit        *store.Audit
	Notifier     OrderNotifier
	Integrations *integrations.Resolver
	Log          *slog.Logger
}

// OrderNotifier is what we need from a notification channel.
type OrderNotifier interface {
	SendText(ctx context.Context, phone, message string) error
}

// Facets handles GET /api/v1/orders/facets.
func (h *Orders) FacetsList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	out, err := h.Facets.All(ctx, h.StoreID)
	if err != nil {
		h.Log.Error("facets failed", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load facets")
		return
	}
	writeJSON(w, http.StatusOK, out)
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
	Number              string          `json:"number"`
	Carrier             string          `json:"carrier"`
	Service             string          `json:"service,omitempty"`
	ServiceCode         string          `json:"service_code,omitempty"`
	URL                 string          `json:"url,omitempty"`
	Status              frenet.Status   `json:"status"`
	StatusLabel         string          `json:"status_label"`
	Health              string          `json:"health"`
	HealthLabel         string          `json:"health_label"`
	SLAState            string          `json:"sla_state,omitempty"`
	SLABreachedStage    string          `json:"sla_breached_stage,omitempty"`
	LastEvent           string          `json:"last_event,omitempty"`
	LastEventAt         *time.Time      `json:"last_event_at,omitempty"`
	EstDelivery         *time.Time      `json:"estimated_delivery,omitempty"`
	DeliveredAt         *time.Time      `json:"delivered_at,omitempty"`
	LabelIssuedAt       *time.Time      `json:"label_issued_at,omitempty"`
	PreparingAt         *time.Time      `json:"preparing_at,omitempty"`
	ReadyForPickupAt    *time.Time      `json:"ready_for_pickup_at,omitempty"`
	PostedAt            *time.Time      `json:"posted_at,omitempty"`
	OutForDeliveryAt    *time.Time      `json:"out_for_delivery_at,omitempty"`
	IdleSince           *time.Time      `json:"idle_since,omitempty"`
	RiskScore           int16           `json:"risk_score"`
	Events              []timelineEvent `json:"events,omitempty"`
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
		Limit:   parseInt(q.Get("per_page"), 50, 200),
		Offset:  parseInt(q.Get("offset"), 0, 0),
		Search:  strings.TrimSpace(q.Get("q")),
		Sort:    strings.TrimSpace(q.Get("sort")),
		SortDir: strings.TrimSpace(q.Get("dir")),
	}
	if s := q.Get("status"); s != "" {
		filters.Statuses = strings.Split(s, ",")
	}
	if hf := q.Get("health"); hf != "" {
		filters.Health = strings.Split(hf, ",")
	}
	if c := q.Get("carrier"); c != "" {
		filters.Carriers = strings.Split(c, ",")
	}
	if u := q.Get("uf"); u != "" {
		filters.UFs = strings.Split(strings.ToUpper(u), ",")
	}
	if since := q.Get("since"); since != "" {
		if t, err := time.Parse("2006-01-02", since); err == nil {
			filters.Since = &t
		}
	}
	if until := q.Get("until"); until != "" {
		if t, err := time.Parse("2006-01-02", until); err == nil {
			next := t.AddDate(0, 0, 1) // inclusive end-of-day
			filters.Until = &next
		}
	}

	res, err := h.Orders.List(ctx, filters)
	if err != nil {
		h.Log.Error("list orders failed", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load orders")
		return
	}

	out := make([]orderListItem, 0, len(res.Rows))
	for i := range res.Rows {
		out = append(out, projectListItem(&res.Rows[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"orders": out,
		"total":  res.Total,
		"count":  len(out),
		"limit":  filters.Limit,
		"offset": filters.Offset,
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

	// Capture previous status before mutating.
	var prev string
	dbBefore, _ := h.Orders.GetByWCID(ctx, h.StoreID, wcID)
	if dbBefore != nil {
		prev = dbBefore.Status
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

	if dbBefore != nil {
		_ = h.Orders.UpdateStatus(ctx, dbBefore.ID, body.Status)
		if h.Audit != nil && prev != body.Status {
			_ = h.Audit.RecordStatusChange(ctx, store.StatusChange{
				OrderID:    dbBefore.ID,
				FromStatus: prev,
				ToStatus:   body.Status,
				Source:     "manual",
				Note:       body.Note,
				Actor:      "dashboard",
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"id": wcID, "status": body.Status})
}

// Breakdown handles GET /api/v1/orders/{id}/breakdown — per-stage SLA
// analysis with diagnosis (first delay, worst delay, cascade contribution).
func (h *Orders) Breakdown(w http.ResponseWriter, r *http.Request) {
	wcID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dbOrder, err := h.Orders.GetByWCID(ctx, h.StoreID, wcID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load order")
		return
	}
	ships, _ := h.Shipments.ListByOrder(ctx, dbOrder.ID)
	if len(ships) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"stages": []any{}, "diagnosis": map[string]any{}})
		return
	}
	res := orders.ComputeBreakdown(&ships[0], time.Now().UTC())
	writeJSON(w, http.StatusOK, res)
}

// History handles GET /api/v1/orders/{id}/history (status audit log).
func (h *Orders) History(w http.ResponseWriter, r *http.Request) {
	wcID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dbOrder, err := h.Orders.GetByWCID(ctx, h.StoreID, wcID)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusOK, map[string]any{"changes": []any{}, "notifications": []any{}})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load order")
		return
	}

	changes, _ := h.Audit.ListStatusChanges(ctx, dbOrder.ID)
	notes, _ := h.Audit.ListNotifications(ctx, dbOrder.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"changes":       changes,
		"notifications": notes,
	})
}

// Notify handles POST /api/v1/orders/{id}/notify — sends a WhatsApp text
// to the customer's phone via the configured WAHA gateway and records it.
func (h *Orders) Notify(w http.ResponseWriter, r *http.Request) {
	wcID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var body struct {
		Message  string `json:"message"`
		Template string `json:"template,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	body.Message = strings.TrimSpace(body.Message)
	if body.Message == "" {
		writeError(w, http.StatusUnprocessableEntity, "message is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	dbOrder, err := h.Orders.GetByWCID(ctx, h.StoreID, wcID)
	if err != nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}
	if strings.TrimSpace(dbOrder.CustomerPhone) == "" {
		writeError(w, http.StatusUnprocessableEntity, "this order has no phone number")
		return
	}
	if h.Notifier == nil {
		writeError(w, http.StatusServiceUnavailable, "notifier not configured")
		return
	}

	sendErr := h.Notifier.SendText(ctx, dbOrder.CustomerPhone, body.Message)
	rec := store.Notification{
		OrderID:  dbOrder.ID,
		Channel:  "waha",
		Template: body.Template,
		Status:   "sent",
	}
	if sendErr != nil {
		rec.Status = "failed"
		rec.Error = sendErr.Error()
	}
	rec.Payload, _ = json.Marshal(map[string]any{"message": body.Message, "phone": dbOrder.CustomerPhone})
	_ = h.Audit.RecordNotification(ctx, rec)

	if sendErr != nil {
		h.Log.Warn("waha send failed", "wc_order_id", wcID, "err", sendErr)
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": sendErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "Mensagem enviada via WhatsApp."})
}

// ExportCSV handles GET /api/v1/orders/export.csv with the same filters as
// List, streaming a UTF-8 CSV the dashboard can let ops download.
func (h *Orders) ExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	q := r.URL.Query()
	filters := store.ListFilters{
		StoreID: h.StoreID,
		Limit:   1000,
		Offset:  0,
		Search:  strings.TrimSpace(q.Get("q")),
	}
	if s := q.Get("status"); s != "" {
		filters.Statuses = strings.Split(s, ",")
	}
	if hf := q.Get("health"); hf != "" {
		filters.Health = strings.Split(hf, ",")
	}
	if c := q.Get("carrier"); c != "" {
		filters.Carriers = strings.Split(c, ",")
	}
	if u := q.Get("uf"); u != "" {
		filters.UFs = strings.Split(strings.ToUpper(u), ",")
	}
	if since := q.Get("since"); since != "" {
		if t, err := time.Parse("2006-01-02", since); err == nil {
			filters.Since = &t
		}
	}
	if until := q.Get("until"); until != "" {
		if t, err := time.Parse("2006-01-02", until); err == nil {
			next := t.AddDate(0, 0, 1)
			filters.Until = &next
		}
	}

	res, err := h.Orders.List(ctx, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load orders")
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="pedidos.csv"`)
	w.Write([]byte("\xEF\xBB\xBF")) // UTF-8 BOM for Excel compatibility
	cw := csv.NewWriter(w)
	defer cw.Flush()

	_ = cw.Write([]string{
		"pedido_wc", "criado_em", "pago_em", "status_wc", "cliente",
		"cidade", "uf", "metodo_envio", "total_brl",
		"rastreio", "transportadora", "servico",
		"saude", "ultimo_evento", "ultimo_evento_em", "eta",
	})
	for _, row := range res.Rows {
		o := row.Order
		var s store.Shipment
		if row.Shipment != nil {
			s = *row.Shipment
		}
		_ = cw.Write([]string{
			strconv.FormatInt(o.WCOrderID, 10),
			o.CreatedAt.Format(time.RFC3339),
			optionalTime(o.PaidAt),
			o.Status,
			o.CustomerName,
			o.CustomerCity,
			o.CustomerUF,
			o.ShippingMethod,
			strconv.FormatFloat(o.TotalBRL, 'f', 2, 64),
			s.TrackingCode,
			s.Carrier,
			s.Service,
			s.Health,
			s.LastEvent,
			optionalTime(s.LastEventAt),
			optionalTime(s.EstimatedDelivery),
		})
	}
}

func optionalTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
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
		li.Tracking.SLAState = s.SLAState
		li.Tracking.SLABreachedStage = s.SLABreachedStage
		li.Tracking.LastEvent = s.LastEvent
		li.Tracking.LastEventAt = s.LastEventAt
		li.Tracking.EstDelivery = s.EstimatedDelivery
		li.Tracking.DeliveredAt = s.DeliveredAt
		li.Tracking.LabelIssuedAt = s.LabelIssuedAt
		li.Tracking.PreparingAt = s.PreparingAt
		li.Tracking.ReadyForPickupAt = s.ReadyForPickupAt
		li.Tracking.PostedAt = s.PostedAt
		li.Tracking.OutForDeliveryAt = s.OutForDeliveryAt
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
			Location:    prettifyLocation(e.Location),
			Type:        e.Type,
		})
	}
	return out
}

// prettifyLocation turns Frenet's "BELO HORIZONTE-MG" / "RECIFE-PE-BR" into
// the title-cased "Belo Horizonte · MG" form the dashboard expects.
func prettifyLocation(loc string) string {
	loc = strings.TrimSpace(loc)
	if loc == "" || loc == "-BR" || loc == "BR" {
		return ""
	}
	loc = strings.TrimSuffix(loc, "-BR")
	parts := strings.Split(loc, "-")
	city := strings.TrimSpace(parts[0])
	city = titleCasePT(city)
	if len(parts) >= 2 {
		uf := strings.ToUpper(strings.TrimSpace(parts[1]))
		if city == "" {
			return uf
		}
		return city + " · " + uf
	}
	return city
}

// titleCasePT lowercases acronyms-free input and uppercases each word's
// first letter — Frenet sends "BELO HORIZONTE", we want "Belo Horizonte".
// Conjunctions like "do/da/de" stay lowercase except at the start.
func titleCasePT(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	parts := strings.Fields(s)
	keepLower := map[string]bool{"de": true, "do": true, "da": true, "dos": true, "das": true, "e": true}
	for i, p := range parts {
		if i > 0 && keepLower[p] {
			continue
		}
		runes := []rune(p)
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
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
