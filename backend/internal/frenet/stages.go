package frenet

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// MapEventToStage classifies a Frenet event description into the
// rastreiaki stage it represents (column name in shipments). Returns
// empty string when the event doesn't map to a milestone.
//
// Order matters: more specific patterns first.
func MapEventToStage(description string) string {
	d := normalizeStage(description)

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
		strings.Contains(d, "encaminhado"),
		strings.Contains(d, "correcao de rota"),
		strings.Contains(d, "em curso"):
		return "in_transit_at"

	case strings.Contains(d, "postado"),
		strings.Contains(d, "objeto postado"),
		strings.Contains(d, "postagem efetuada"):
		return "posted_at"

	case strings.Contains(d, "pronto para coleta"),
		strings.Contains(d, "aguardando coleta"),
		strings.Contains(d, "objeto aguardando coleta"):
		return "ready_for_pickup_at"

	case strings.Contains(d, "em preparacao"),
		strings.Contains(d, "preparando objeto"),
		strings.Contains(d, "aguardando postagem"):
		return "preparing_at"

	case strings.Contains(d, "etiqueta emitida"),
		strings.Contains(d, "etiqueta gerada"):
		return "label_issued_at"
	}
	return ""
}

func normalizeStage(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, _ := transform.String(t, s)
	return strings.ToLower(out)
}
