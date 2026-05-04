// Package settings persists dynamic configuration (integrations, feature
// flags, future tunables) so the dashboard can edit them without a redeploy.
//
// Reads pass through a small in-process TTL cache so the hot path
// (sync workers, request handlers) hits Postgres at most a handful of
// times per second per process. Writes invalidate the cache key.
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Keys persisted in app_settings. Using string constants instead of an
// enum lets us treat them as opaque blobs from SQL while keeping the
// type-safe accessors below.
const (
	KeyWooCommerce = "integration.woocommerce"
	KeyFrenet      = "integration.frenet"
	KeyWAHA        = "integration.waha"
)

// WooCommerceConfig is the persisted shape of the WooCommerce integration.
type WooCommerceConfig struct {
	URL            string `json:"url"`
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	WebhookSecret  string `json:"webhook_secret"`
	Enabled        bool   `json:"enabled"`
}

// FrenetConfig is the persisted shape of the Frenet integration.
type FrenetConfig struct {
	APIToken      string `json:"api_token"`
	PanelEmail    string `json:"panel_email"`
	PanelPassword string `json:"panel_password"`
	Enabled       bool   `json:"enabled"`
}

// WAHAConfig is the persisted shape of the WhatsApp HTTP API integration.
type WAHAConfig struct {
	URL            string `json:"url"`
	APIKey         string `json:"api_key"`
	DefaultSession string `json:"default_session"`
	Enabled        bool   `json:"enabled"`
}

// Store reads and writes typed configuration to app_settings.
type Store struct {
	Pool *pgxpool.Pool
	TTL  time.Duration

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	raw       json.RawMessage
	expiresAt time.Time
}

// New returns a Store with a 30s default TTL.
func New(pool *pgxpool.Pool) *Store {
	return &Store{
		Pool:  pool,
		TTL:   30 * time.Second,
		cache: make(map[string]cacheEntry),
	}
}

// GetWooCommerce returns the persisted WooCommerce config.
func (s *Store) GetWooCommerce(ctx context.Context) (WooCommerceConfig, error) {
	var c WooCommerceConfig
	err := s.get(ctx, KeyWooCommerce, &c)
	return c, err
}

// SetWooCommerce upserts the WooCommerce config and invalidates the cache.
func (s *Store) SetWooCommerce(ctx context.Context, c WooCommerceConfig) error {
	return s.set(ctx, KeyWooCommerce, c)
}

// GetFrenet returns the persisted Frenet config.
func (s *Store) GetFrenet(ctx context.Context) (FrenetConfig, error) {
	var c FrenetConfig
	err := s.get(ctx, KeyFrenet, &c)
	return c, err
}

// SetFrenet upserts the Frenet config and invalidates the cache.
func (s *Store) SetFrenet(ctx context.Context, c FrenetConfig) error {
	return s.set(ctx, KeyFrenet, c)
}

// GetWAHA returns the persisted WAHA config.
func (s *Store) GetWAHA(ctx context.Context) (WAHAConfig, error) {
	var c WAHAConfig
	err := s.get(ctx, KeyWAHA, &c)
	return c, err
}

// SetWAHA upserts the WAHA config and invalidates the cache.
func (s *Store) SetWAHA(ctx context.Context, c WAHAConfig) error {
	return s.set(ctx, KeyWAHA, c)
}

// Invalidate forces the next Get to hit the DB.
func (s *Store) Invalidate(key string) {
	s.mu.Lock()
	delete(s.cache, key)
	s.mu.Unlock()
}

func (s *Store) get(ctx context.Context, key string, out any) error {
	s.mu.RLock()
	if entry, ok := s.cache[key]; ok && time.Now().Before(entry.expiresAt) {
		s.mu.RUnlock()
		return decodeJSON(entry.raw, out)
	}
	s.mu.RUnlock()

	var raw json.RawMessage
	err := s.Pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = $1`, key).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		raw = json.RawMessage(`{}`)
	} else if err != nil {
		return fmt.Errorf("settings get %q: %w", key, err)
	}

	s.mu.Lock()
	s.cache[key] = cacheEntry{raw: raw, expiresAt: time.Now().Add(s.TTL)}
	s.mu.Unlock()

	return decodeJSON(raw, out)
}

func (s *Store) set(ctx context.Context, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("settings marshal %q: %w", key, err)
	}
	_, err = s.Pool.Exec(ctx, `
INSERT INTO app_settings (key, value, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
		key, raw)
	if err != nil {
		return fmt.Errorf("settings upsert %q: %w", key, err)
	}
	s.Invalidate(key)
	return nil
}

func decodeJSON(raw json.RawMessage, out any) error {
	if len(raw) == 0 || string(raw) == "{}" {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("settings decode: %w", err)
	}
	return nil
}
