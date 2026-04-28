// Package integrations resolves third-party clients (WooCommerce, Frenet,
// WAHA) at request/job time using the dynamic settings store. The DB is
// the source of truth; env vars stay around as a bootstrap fallback so
// the very first deploy still works.
package integrations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/univerbeauty777/univer-tracker/backend/internal/config"
	"github.com/univerbeauty777/univer-tracker/backend/internal/frenet"
	"github.com/univerbeauty777/univer-tracker/backend/internal/settings"
	"github.com/univerbeauty777/univer-tracker/backend/internal/woocommerce"
)

// ErrNotConfigured is returned when neither the DB nor env vars define
// usable credentials for a given integration.
var ErrNotConfigured = errors.New("integration not configured")

// Resolver constructs third-party clients fresh, looking up credentials
// from settings (with env fallback) every time. Cheap because the
// settings.Store caches values in-process for 30s.
type Resolver struct {
	Settings *settings.Store
	Env      *config.Config
}

// New returns a Resolver wired to the given settings store and env config.
func New(s *settings.Store, env *config.Config) *Resolver {
	return &Resolver{Settings: s, Env: env}
}

// WooCommerceConfig returns the merged config (settings overrides env).
func (r *Resolver) WooCommerceConfig(ctx context.Context) (settings.WooCommerceConfig, error) {
	c, err := r.Settings.GetWooCommerce(ctx)
	if err != nil {
		return c, err
	}
	if strings.TrimSpace(c.URL) == "" {
		c.URL = r.Env.WooCommerce.URL
	}
	if c.ConsumerKey == "" {
		c.ConsumerKey = r.Env.WooCommerce.ConsumerKey
	}
	if c.ConsumerSecret == "" {
		c.ConsumerSecret = r.Env.WooCommerce.ConsumerSecret
	}
	if c.WebhookSecret == "" {
		c.WebhookSecret = r.Env.WooCommerce.WebhookSecret
	}
	return c, nil
}

// WooCommerce returns a configured client or ErrNotConfigured.
func (r *Resolver) WooCommerce(ctx context.Context) (*woocommerce.Client, error) {
	c, err := r.WooCommerceConfig(ctx)
	if err != nil {
		return nil, err
	}
	if c.URL == "" || c.ConsumerKey == "" || c.ConsumerSecret == "" {
		return nil, fmt.Errorf("%w: woocommerce", ErrNotConfigured)
	}
	return woocommerce.New(c.URL, c.ConsumerKey, c.ConsumerSecret), nil
}

// FrenetConfig returns the merged Frenet config.
func (r *Resolver) FrenetConfig(ctx context.Context) (settings.FrenetConfig, error) {
	c, err := r.Settings.GetFrenet(ctx)
	if err != nil {
		return c, err
	}
	if c.APIToken == "" {
		c.APIToken = r.Env.Frenet.APIToken
	}
	if c.PanelEmail == "" {
		c.PanelEmail = r.Env.Frenet.PanelEmail
	}
	if c.PanelPassword == "" {
		c.PanelPassword = r.Env.Frenet.PanelPassword
	}
	return c, nil
}

// Frenet returns a configured client or ErrNotConfigured.
func (r *Resolver) Frenet(ctx context.Context) (*frenet.Client, error) {
	c, err := r.FrenetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(c.APIToken) == "" {
		return nil, fmt.Errorf("%w: frenet", ErrNotConfigured)
	}
	return frenet.New(c.APIToken), nil
}

// WAHAConfig returns the merged WAHA config.
func (r *Resolver) WAHAConfig(ctx context.Context) (settings.WAHAConfig, error) {
	c, err := r.Settings.GetWAHA(ctx)
	if err != nil {
		return c, err
	}
	if c.URL == "" {
		c.URL = r.Env.WAHA.URL
	}
	if c.APIKey == "" {
		c.APIKey = r.Env.WAHA.APIKey
	}
	return c, nil
}
