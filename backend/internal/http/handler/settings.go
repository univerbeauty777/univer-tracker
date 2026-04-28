package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/frenet"
	"github.com/univerbeauty777/univer-tracker/backend/internal/integrations"
	"github.com/univerbeauty777/univer-tracker/backend/internal/settings"
	"github.com/univerbeauty777/univer-tracker/backend/internal/woocommerce"
)

// Settings serves the integrations CRUD UI.
type Settings struct {
	Store    *settings.Store
	Resolver *integrations.Resolver
	Log      *slog.Logger
}

// maskedSecret hides any non-empty value as a fixed-length placeholder so
// the dashboard never sees raw secrets after they're saved. The PATCH
// handler treats the placeholder as "keep current value".
const maskedSecret = "••••••••"

// integrationView is the shape returned by GET — secrets masked.
type integrationView struct {
	WooCommerce wooView    `json:"woocommerce"`
	Frenet      frenetView `json:"frenet"`
	WAHA        wahaView   `json:"waha"`
}

type wooView struct {
	URL            string `json:"url"`
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	WebhookSecret  string `json:"webhook_secret"`
	Enabled        bool   `json:"enabled"`
	Configured     bool   `json:"configured"`
}

type frenetView struct {
	APIToken      string `json:"api_token"`
	PanelEmail    string `json:"panel_email"`
	PanelPassword string `json:"panel_password"`
	Enabled       bool   `json:"enabled"`
	Configured    bool   `json:"configured"`
}

type wahaView struct {
	URL        string `json:"url"`
	APIKey     string `json:"api_key"`
	Enabled    bool   `json:"enabled"`
	Configured bool   `json:"configured"`
}

// Get handles GET /api/v1/settings/integrations.
func (h *Settings) Get(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	wc, _ := h.Resolver.WooCommerceConfig(ctx)
	fr, _ := h.Resolver.FrenetConfig(ctx)
	waha, _ := h.Resolver.WAHAConfig(ctx)

	writeJSON(w, http.StatusOK, integrationView{
		WooCommerce: wooView{
			URL:            wc.URL,
			ConsumerKey:    wc.ConsumerKey,
			ConsumerSecret: maskIfSet(wc.ConsumerSecret),
			WebhookSecret:  maskIfSet(wc.WebhookSecret),
			Enabled:        wc.Enabled,
			Configured:     wc.URL != "" && wc.ConsumerKey != "" && wc.ConsumerSecret != "",
		},
		Frenet: frenetView{
			APIToken:      maskIfSet(fr.APIToken),
			PanelEmail:    fr.PanelEmail,
			PanelPassword: maskIfSet(fr.PanelPassword),
			Enabled:       fr.Enabled,
			Configured:    fr.APIToken != "",
		},
		WAHA: wahaView{
			URL:        waha.URL,
			APIKey:     maskIfSet(waha.APIKey),
			Enabled:    waha.Enabled,
			Configured: waha.URL != "" && waha.APIKey != "",
		},
	})
}

// UpdateWooCommerce handles PATCH /api/v1/settings/integrations/woocommerce.
func (h *Settings) UpdateWooCommerce(w http.ResponseWriter, r *http.Request) {
	var body wooView
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	current, _ := h.Store.GetWooCommerce(ctx)
	updated := settings.WooCommerceConfig{
		URL:            firstNonEmpty(strings.TrimSpace(body.URL), current.URL),
		ConsumerKey:    firstNonEmpty(strings.TrimSpace(body.ConsumerKey), current.ConsumerKey),
		ConsumerSecret: keepOrReplace(body.ConsumerSecret, current.ConsumerSecret),
		WebhookSecret:  keepOrReplace(body.WebhookSecret, current.WebhookSecret),
		Enabled:        body.Enabled,
	}

	if err := h.Store.SetWooCommerce(ctx, updated); err != nil {
		h.Log.Error("save wc settings", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save settings")
		return
	}
	h.Get(w, r)
}

// UpdateFrenet handles PATCH /api/v1/settings/integrations/frenet.
func (h *Settings) UpdateFrenet(w http.ResponseWriter, r *http.Request) {
	var body frenetView
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	current, _ := h.Store.GetFrenet(ctx)
	updated := settings.FrenetConfig{
		APIToken:      keepOrReplace(body.APIToken, current.APIToken),
		PanelEmail:    firstNonEmpty(strings.TrimSpace(body.PanelEmail), current.PanelEmail),
		PanelPassword: keepOrReplace(body.PanelPassword, current.PanelPassword),
		Enabled:       body.Enabled,
	}

	if err := h.Store.SetFrenet(ctx, updated); err != nil {
		h.Log.Error("save frenet settings", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save settings")
		return
	}
	h.Get(w, r)
}

// UpdateWAHA handles PATCH /api/v1/settings/integrations/waha.
func (h *Settings) UpdateWAHA(w http.ResponseWriter, r *http.Request) {
	var body wahaView
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	current, _ := h.Store.GetWAHA(ctx)
	updated := settings.WAHAConfig{
		URL:     firstNonEmpty(strings.TrimSpace(body.URL), current.URL),
		APIKey:  keepOrReplace(body.APIKey, current.APIKey),
		Enabled: body.Enabled,
	}

	if err := h.Store.SetWAHA(ctx, updated); err != nil {
		h.Log.Error("save waha settings", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save settings")
		return
	}
	h.Get(w, r)
}

// TestWooCommerce hits /system_status with the merged credentials so the
// dashboard can confirm a config works before saving permanently.
func (h *Settings) TestWooCommerce(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	client, err := h.Resolver.WooCommerce(ctx)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	if _, err := client.ListOrders(ctx, woocommerce.ListOrdersParams{PerPage: 1}); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "WooCommerce respondeu com sucesso."})
}

// TestFrenet sends a tiny tracking probe just to validate the token.
func (h *Settings) TestFrenet(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	client, err := h.Resolver.Frenet(ctx)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	// Frenet returns 200 with an ErrorMessage when the tracking is invalid,
	// which is exactly what we want to test the auth path.
	resp, err := client.GetTrackingInfo(ctx, "AA000000000BR", "")
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if isAuthError(resp) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": resp.ErrorMessage})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "Token Frenet aceito pela API."})
}

// TestWAHA pings /api/sessions which is the standard WAHA health endpoint.
func (h *Settings) TestWAHA(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	cfg, err := h.Resolver.WAHAConfig(ctx)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if cfg.URL == "" || cfg.APIKey == "" {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "WAHA URL ou API Key não configuradas."})
		return
	}

	url := strings.TrimRight(cfg.URL, "/") + "/api/sessions"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	req.Header.Set("X-Api-Key", cfg.APIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "WAHA respondeu " + resp.Status})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "WAHA respondeu com sucesso."})
}

func maskIfSet(s string) string {
	if s == "" {
		return ""
	}
	return maskedSecret
}

func keepOrReplace(input, current string) string {
	t := strings.TrimSpace(input)
	if t == "" || t == maskedSecret {
		return current
	}
	return t
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func isAuthError(resp *frenet.TrackingResponse) bool {
	if resp == nil {
		return false
	}
	msg := strings.ToLower(resp.ErrorMessage)
	return strings.Contains(msg, "token") ||
		strings.Contains(msg, "auth") ||
		strings.Contains(msg, "unauthor")
}
