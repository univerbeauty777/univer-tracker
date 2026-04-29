package store

import (
	"encoding/json"
	"time"
)

// Order is the persisted projection of a WooCommerce order. We mirror only
// the fields we need for tracking, list views and analytics — the original
// payload is not kept (we always have WooCommerce as the system of record).
type Order struct {
	ID              int64           `json:"id"`
	StoreID         int64           `json:"store_id"`
	WCOrderID       int64           `json:"wc_order_id"`
	Status          string          `json:"status"`
	CustomerName    string          `json:"customer_name"`
	CustomerEmail   string          `json:"customer_email"`
	CustomerPhone   string          `json:"customer_phone"`
	CustomerCity    string          `json:"customer_city"`
	CustomerUF      string          `json:"customer_uf"`
	ShippingMethod  string          `json:"shipping_method"`
	TotalBRL        float64         `json:"total_brl"`
	DeclaredValue   float64         `json:"declared_value"`
	Tags            []string        `json:"tags"`
	PaidAt          *time.Time      `json:"paid_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// Shipment is the persisted shipment tied to one Order. A single order
// usually has one shipment, but the schema allows more (split shipments).
type Shipment struct {
	ID                int64      `json:"id"`
	OrderID           int64      `json:"order_id"`
	TrackingCode      string     `json:"tracking_code"`
	Carrier           string     `json:"carrier"`
	Service           string     `json:"service"`
	ServiceCode       string     `json:"service_code"`
	TrackingURL       string     `json:"tracking_url"`
	Status            string     `json:"status"`
	LastEvent         string     `json:"last_event"`
	LastEventAt       *time.Time `json:"last_event_at,omitempty"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	LabelIssuedAt        *time.Time `json:"label_issued_at,omitempty"`
	PreparingAt          *time.Time `json:"preparing_at,omitempty"`
	ReadyForPickupAt     *time.Time `json:"ready_for_pickup_at,omitempty"`
	PostedAt             *time.Time `json:"posted_at,omitempty"`
	InTransitAt          *time.Time `json:"in_transit_at,omitempty"`
	AtDestinationCityAt  *time.Time `json:"at_destination_city_at,omitempty"`
	OutForDeliveryAt     *time.Time `json:"out_for_delivery_at,omitempty"`
	DeliveredAt          *time.Time `json:"delivered_at,omitempty"`
	LastSyncedAt         *time.Time `json:"last_synced_at,omitempty"`
	Health               string     `json:"health"`
	SLAState             string     `json:"sla_state"`
	SLABreachedStage     string     `json:"sla_breached_stage"`
	IdleSince            *time.Time `json:"idle_since,omitempty"`
	RiskScore            int16      `json:"risk_score"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// Event is one carrier event (Postado / Em trânsito / Entregue …) on a Shipment.
type Event struct {
	ID          int64           `json:"id"`
	ShipmentID  int64           `json:"shipment_id"`
	OccurredAt  time.Time       `json:"occurred_at"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Location    string          `json:"location"`
	Raw         json.RawMessage `json:"raw,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// SyncState is the worker's bookkeeping for incremental syncs.
type SyncState struct {
	Entity       string          `json:"entity"`
	LastSyncedAt *time.Time      `json:"last_synced_at,omitempty"`
	Extra        json.RawMessage `json:"extra"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
