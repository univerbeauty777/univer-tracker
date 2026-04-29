package frenet

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// MapEventToStages classifies a Frenet event description into one or
// more rastreiaki stages it represents. Returns an empty slice when
// the event doesn't map to any milestone.
//
// Multiple stages can apply because Frenet's first canonical event is
// "Etiqueta emitida - Aguardando postagem pelo remetente": that single
// event simultaneously means label_issued_at AND preparing_at — the
// label was generated AND the seller is in the preparing window.
// Returning both keeps the funnel honest (149 etiquetas emitidas → 149
// em preparação) instead of forcing a single classification.
func MapEventToStages(description string) []string {
	d := normalizeStage(description)

	// Combined first event: label issued + currently preparing.
	if strings.Contains(d, "etiqueta emitida") && strings.Contains(d, "aguardando postagem") {
		return []string{"label_issued_at", "preparing_at"}
	}
	if s := mapSingleStage(d); s != "" {
		return []string{s}
	}
	return nil
}

// MapEventToStage is the back-compat single-result variant; first stage
// wins. New code should prefer MapEventToStages so multi-stage events
// don't drop the secondary classification.
func MapEventToStage(description string) string {
	stages := MapEventToStages(description)
	if len(stages) == 0 {
		return ""
	}
	return stages[0]
}

func mapSingleStage(d string) string {
	switch {
	case strings.Contains(d, "entregue"),
		strings.Contains(d, "entrega efetuada"),
		strings.Contains(d, "entrega realizada"):
		return "delivered_at"

	case strings.Contains(d, "saiu para entrega"),
		strings.Contains(d, "encaminhado para entrega"),
		strings.Contains(d, "em rota de entrega"),
		strings.Contains(d, "com o entregador"):
		return "out_for_delivery_at"

	case strings.Contains(d, "chegou na unidade de tratamento"),
		strings.Contains(d, "chegou na regiao"),
		strings.Contains(d, "na unidade de distribuicao"):
		return "at_destination_city_at"

	case strings.Contains(d, "em transito"),
		strings.Contains(d, "em transferencia"),
		strings.Contains(d, "correcao de rota"),
		strings.Contains(d, "em curso"):
		return "in_transit_at"

	case strings.Contains(d, "postado"),
		strings.Contains(d, "postagem efetuada"):
		return "posted_at"

	case strings.Contains(d, "etiqueta emitida"),
		strings.Contains(d, "etiqueta gerada"):
		return "label_issued_at"

	case strings.Contains(d, "pronto para coleta"),
		strings.Contains(d, "aguardando coleta"),
		strings.Contains(d, "objeto aguardando coleta"):
		return "ready_for_pickup_at"

	case strings.Contains(d, "em preparacao"),
		strings.Contains(d, "preparando objeto"),
		strings.Contains(d, "aguardando postagem"):
		return "preparing_at"

	case strings.Contains(d, "encaminhado"):
		// Generic "encaminhado" comes last so the more specific
		// "encaminhado para entrega" wins above.
		return "in_transit_at"
	}
	return ""
}

func normalizeStage(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, _ := transform.String(t, s)
	return strings.ToLower(out)
}
