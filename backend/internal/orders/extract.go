// Package orders builds the unified Order projection that the API exposes:
// WooCommerce data, Frenet tracking events and a normalized status, all
// merged in a single response.
package orders

import (
	"strings"

	"github.com/univerbeauty777/univer-tracker/backend/internal/frenet"
	"github.com/univerbeauty777/univer-tracker/backend/internal/woocommerce"
)

// trackingMetaKeys are the meta keys WooCommerce stores tracking numbers under,
// in priority order. Several plugins (Frenet, Melhor Envio, AfterShip, plain
// Correios) use different keys, so we check them all.
var trackingMetaKeys = []string{
	"_utrk_tracking_number",
	"_tracking_number",
	"_tracking_code",
	"_frenet_tracking_number",
	"_frenet_tracking_code",
	"_correios_tracking_code",
	"_aftership_tracking_number",
}

var carrierMetaKeys = []string{
	"_utrk_carrier_name",
	"_tracking_provider",
	"_frenet_carrier",
}

var serviceCodeMetaKeys = []string{
	"_utrk_shipping_service_code",
	"_frenet_service_code",
}

// TrackingInfo summarizes what we know about a shipment tied to an order.
type TrackingInfo struct {
	Number      string                 `json:"number"`
	Carrier     string                 `json:"carrier"`
	ServiceCode string                 `json:"service_code,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Status      frenet.Status          `json:"status"`
	StatusLabel string                 `json:"status_label"`
	Events      []frenet.TrackingEvent `json:"events,omitempty"`
}

// FromOrder picks the first tracking number/carrier/service code present in
// the order's meta_data, regardless of which plugin wrote it.
func FromOrder(o *woocommerce.Order) TrackingInfo {
	if o == nil {
		return TrackingInfo{}
	}
	return TrackingInfo{
		Number:      lookup(o.MetaData, trackingMetaKeys),
		Carrier:     lookup(o.MetaData, carrierMetaKeys),
		ServiceCode: lookup(o.MetaData, serviceCodeMetaKeys),
		Status:      frenet.StatusUnknown,
		StatusLabel: StatusLabel(frenet.StatusUnknown),
	}
}

// HasTracking returns true when we have at least a tracking number.
func (t TrackingInfo) HasTracking() bool {
	return strings.TrimSpace(t.Number) != ""
}

// StatusLabel translates a normalized status into a Portuguese display label.
func StatusLabel(s frenet.Status) string {
	switch s {
	case frenet.StatusLabelCreated:
		return "Etiqueta emitida"
	case frenet.StatusShipped:
		return "Postado"
	case frenet.StatusInTransit:
		return "Em trânsito"
	case frenet.StatusOutForDelivery:
		return "Saiu para entrega"
	case frenet.StatusDelivered:
		return "Entregue"
	case frenet.StatusDeliveryFailed:
		return "Tentativa de entrega"
	case frenet.StatusReturned:
		return "Devolvido"
	default:
		return "Aguardando rastreio"
	}
}

func lookup(meta []woocommerce.Meta, keys []string) string {
	for _, key := range keys {
		for _, m := range meta {
			if m.Key != key {
				continue
			}
			s, ok := m.Value.(string)
			if ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
