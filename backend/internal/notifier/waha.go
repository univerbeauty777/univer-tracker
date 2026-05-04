// Package notifier wraps the outbound channels (WhatsApp via WAHA today,
// email/SMS later) into one Send API the application code consumes.
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/integrations"
)

// ErrNotConfigured surfaces when the WAHA settings are missing.
var ErrNotConfigured = errors.New("waha not configured")

// WAHA is a thin client over the WhatsApp HTTP API gateway.
type WAHA struct {
	Resolver *integrations.Resolver
	Session  string // WAHA session name; defaults to "default"
}

// New returns a WAHA client. Pass an empty session to use "default".
func New(r *integrations.Resolver, session string) *WAHA {
	if session == "" {
		session = "default"
	}
	return &WAHA{Resolver: r, Session: session}
}

// SendText posts a plain-text message using the configured default
// session.
func (w *WAHA) SendText(ctx context.Context, phone, message string) error {
	return w.SendTextWith(ctx, "", phone, message)
}

// SendTextWith posts a plain-text message and lets the caller override
// the WAHA session. Pass an empty `session` to fall back to the
// configured default (settings.waha.default_session) — and finally to
// the literal "default" session if neither is set.
func (w *WAHA) SendTextWith(ctx context.Context, session, phone, message string) error {
	cfg, err := w.Resolver.WAHAConfig(ctx)
	if err != nil {
		return fmt.Errorf("resolve waha: %w", err)
	}
	if cfg.URL == "" || cfg.APIKey == "" {
		return ErrNotConfigured
	}

	chatID, err := normalizePhoneToChat(phone)
	if err != nil {
		return err
	}

	useSession := strings.TrimSpace(session)
	if useSession == "" {
		useSession = strings.TrimSpace(cfg.DefaultSession)
	}
	if useSession == "" {
		useSession = w.Session
	}
	if useSession == "" {
		useSession = "default"
	}

	body, _ := json.Marshal(map[string]any{
		"session": useSession,
		"chatId":  chatID,
		"text":    message,
	})

	url := strings.TrimRight(cfg.URL, "/") + "/api/sendText"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", cfg.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("waha send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("waha http %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

// SessionInfo is the slim shape of a WAHA session the dashboard cares
// about — name + working state, nothing else.
type SessionInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ListSessions calls WAHA's /api/sessions and returns the names + status
// so the UI can populate a session picker.
func (w *WAHA) ListSessions(ctx context.Context) ([]SessionInfo, error) {
	cfg, err := w.Resolver.WAHAConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve waha: %w", err)
	}
	if cfg.URL == "" || cfg.APIKey == "" {
		return nil, ErrNotConfigured
	}

	url := strings.TrimRight(cfg.URL, "/") + "/api/sessions"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Api-Key", cfg.APIKey)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("waha list sessions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("waha http %d: %s", resp.StatusCode, string(raw))
	}

	var raw []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode sessions: %w", err)
	}
	out := make([]SessionInfo, 0, len(raw))
	for _, s := range raw {
		if s.Name == "" {
			continue
		}
		out = append(out, SessionInfo{Name: s.Name, Status: s.Status})
	}
	return out, nil
}

var digitsOnly = regexp.MustCompile(`\D+`)

// normalizePhoneToChat turns "(82) 99608-7578" / "82996087578" / "+55 82 99608-7578"
// into "5582996087578@c.us", which is what WAHA expects on chatId.
func normalizePhoneToChat(raw string) (string, error) {
	digits := digitsOnly.ReplaceAllString(raw, "")
	if digits == "" {
		return "", fmt.Errorf("empty phone")
	}
	// Add Brazilian country code if missing.
	if len(digits) <= 11 {
		digits = "55" + digits
	}
	if len(digits) < 12 || len(digits) > 13 {
		return "", fmt.Errorf("invalid phone length: %s", raw)
	}
	return digits + "@c.us", nil
}
