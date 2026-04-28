package frenet

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Status is a normalized shipment status used internally.
type Status string

const (
	StatusUnknown        Status = "unknown"
	StatusLabelCreated   Status = "label-created"
	StatusShipped        Status = "shipped"
	StatusInTransit      Status = "in-transit"
	StatusOutForDelivery Status = "out-for-delivery"
	StatusDelivered      Status = "delivered"
	StatusDeliveryFailed Status = "delivery-failed"
	StatusReturned       Status = "returned"
)

// MapEvent classifies a Frenet event description into a normalized Status.
// Order matters: negative patterns checked before positive ones.
func MapEvent(description string) Status {
	d := normalize(description)

	// 1. Delivery failed (must check before "entregue").
	for _, p := range []string{
		"nao entregue", "nao foi possivel", "ausente", "endereco incorreto",
		"endereco insuficiente", "recusado", "tentativa de entrega",
		"nao atendeu", "local nao encontrado",
	} {
		if strings.Contains(d, p) {
			return StatusDeliveryFailed
		}
	}

	// 2. Returned.
	for _, p := range []string{"devolvido", "devolucao", "retorno"} {
		if strings.Contains(d, p) {
			return StatusReturned
		}
	}

	// 3. Delivered.
	for _, p := range []string{"entregue", "entrega efetuada", "entrega realizada"} {
		if strings.Contains(d, p) {
			return StatusDelivered
		}
	}

	// 4. Out for delivery.
	for _, p := range []string{
		"saiu para entrega", "encaminhado para entrega",
		"em rota de entrega", "veiculo saiu para entrega", "com o entregador",
	} {
		if strings.Contains(d, p) {
			return StatusOutForDelivery
		}
	}

	// 5. In transit.
	for _, p := range []string{
		"em transito", "em transferencia", "encaminhado",
		"em curso", "recebido na unidade", "saiu da unidade",
		"correcao de rota",
	} {
		if strings.Contains(d, p) {
			return StatusInTransit
		}
	}

	// 6. Shipped.
	for _, p := range []string{
		"postado", "coletado", "objeto postado", "objeto coletado",
		"recebido na transportadora", "postagem efetuada",
	} {
		if strings.Contains(d, p) {
			return StatusShipped
		}
	}

	// 7. Label created (the carrier accepted but didn't pick up yet).
	for _, p := range []string{
		"etiqueta emitida", "aguardando postagem", "aguardando coleta",
		"objeto aguardando",
	} {
		if strings.Contains(d, p) {
			return StatusLabelCreated
		}
	}

	return StatusUnknown
}

// normalize lowercases and strips diacritics for substring matching.
func normalize(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, _ := transform.String(t, s)
	return strings.ToLower(out)
}
