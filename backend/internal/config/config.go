// Package config loads and validates application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	App         App
	Database    Database
	Redis       Redis
	Auth        Auth
	WooCommerce WooCommerce
	Frenet      Frenet
	ZapGrup     ZapGrup
}

type App struct {
	Env       string
	Port      int
	URL       string
	PublicURL string
}

type Database struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type Redis struct {
	URL string
}

type Auth struct {
	JWTSecret    string
	JWTExpiresIn time.Duration
}

type WooCommerce struct {
	URL            string
	ConsumerKey    string
	ConsumerSecret string
	WebhookSecret  string
}

type Frenet struct {
	APIToken      string
	PanelEmail    string
	PanelPassword string
}

type ZapGrup struct {
	URL      string
	APIToken string
}

// Load reads configuration from environment and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		App: App{
			Env:       getEnv("APP_ENV", "development"),
			Port:      getEnvInt("APP_PORT", 8080),
			URL:       getEnv("APP_URL", "http://localhost:8080"),
			PublicURL: getEnv("PUBLIC_URL", "http://localhost:3000"),
		},
		Database: Database{
			URL:             getEnv("DATABASE_URL", ""),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: Redis{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		Auth: Auth{
			JWTSecret:    getEnv("JWT_SECRET", ""),
			JWTExpiresIn: getEnvDuration("JWT_EXPIRES_IN", 7*24*time.Hour),
		},
		WooCommerce: WooCommerce{
			URL:            getEnv("WC_URL", ""),
			ConsumerKey:    getEnv("WC_CONSUMER_KEY", ""),
			ConsumerSecret: getEnv("WC_CONSUMER_SECRET", ""),
			WebhookSecret:  getEnv("WC_WEBHOOK_SECRET", ""),
		},
		Frenet: Frenet{
			APIToken:      getEnv("FRENET_API_TOKEN", ""),
			PanelEmail:    getEnv("FRENET_PANEL_EMAIL", ""),
			PanelPassword: getEnv("FRENET_PANEL_PASSWORD", ""),
		},
		ZapGrup: ZapGrup{
			URL:      getEnv("ZAPGRUP_URL", ""),
			APIToken: getEnv("ZAPGRUP_API_TOKEN", ""),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.IsProduction() && c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required in production")
	}
	if c.IsProduction() && len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
	}
	return nil
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.App.Env) == "production"
}

// IsDevelopment returns true if running in development environment.
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.App.Env) == "development"
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
