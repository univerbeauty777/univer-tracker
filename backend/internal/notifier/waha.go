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

// SendText posts a plain-text message to the given E.164-ish phone number.
// We accept Brazilian numbers in any common shape and normalize them.
func (w *WAHA) SendText(ctx context.Context, phone, message string) error {
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

	body, _ := json.Marshal(map[string]any{
		"session": w.Session,
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
