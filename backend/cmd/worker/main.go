// Command worker runs the periodic sync jobs that keep our local store in
// step with WooCommerce and Frenet. It owns no HTTP surface; everything is
// a ticker today and will graduate to a queue (Asynq) once we need fanout.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/config"
	"github.com/univerbeauty777/univer-tracker/backend/internal/integrations"
	"github.com/univerbeauty777/univer-tracker/backend/internal/settings"
	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
	"github.com/univerbeauty777/univer-tracker/backend/internal/sync"
	"github.com/univerbeauty777/univer-tracker/backend/pkg/logger"
)

const (
	wcInterval     = 5 * time.Minute
	frenetInterval = 10 * time.Minute
	defaultStoreID = int64(1) // single-store deploy for now
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log := logger.New(cfg.App.Env)
	log.Info("starting worker", "env", cfg.App.Env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Second)
	pool, err := store.Open(dbCtx, cfg.Database.URL)
	dbCancel()
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer pool.Close()

	ordersRepo := &store.Orders{Pool: pool}
	shipmentsRepo := &store.Shipments{Pool: pool}
	eventsRepo := &store.Events{Pool: pool}
	stateRepo := &store.SyncStates{Pool: pool}

	settingsStore := settings.New(pool)
	resolver := integrations.New(settingsStore, cfg)

	wcSync := &sync.WooCommerce{
		Store:        ordersRepo,
		Shipments:    shipmentsRepo,
		State:        stateRepo,
		Integrations: resolver,
		StoreID:      defaultStoreID,
		Log:          log,
	}
	frenetSync := &sync.Frenet{
		Shipments:    shipmentsRepo,
		Events:       eventsRepo,
		Integrations: resolver,
		BatchSize:    50,
		Log:          log,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// One-shot: replay tracking_events into stage timestamps so existing
	// shipments inherit per-etapa data without waiting for the next
	// Frenet event.
	go func() {
		bf := &sync.BackfillStages{Pool: pool, Shipments: shipmentsRepo, Log: log}
		bctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		n, err := bf.Run(bctx)
		if err != nil {
			log.Error("backfill stages failed", "err", err)
			return
		}
		log.Info("backfill stages done", "shipments_updated", n)
	}()

	// Run an initial pass right away so the first deploy has data fast.
	go runWC(ctx, log, wcSync)
	go runFrenet(ctx, log, frenetSync)

	wcTicker := time.NewTicker(wcInterval)
	defer wcTicker.Stop()
	frenetTicker := time.NewTicker(frenetInterval)
	defer frenetTicker.Stop()

	for {
		select {
		case sig := <-stop:
			log.Info("shutting down worker", "signal", sig.String())
			cancel()
			return nil
		case <-wcTicker.C:
			go runWC(ctx, log, wcSync)
		case <-frenetTicker.C:
			go runFrenet(ctx, log, frenetSync)
		}
	}
}

func runWC(ctx context.Context, log interface{ Info(msg string, args ...any); Error(msg string, args ...any) }, s *sync.WooCommerce) {
	stats, err := s.Run(ctx)
	if err != nil {
		log.Error("wc sync failed", "err", err)
		return
	}
	log.Info("wc sync done",
		"synced", stats.Synced,
		"errors", stats.Errors,
		"duration_ms", stats.Finished.Sub(stats.Started).Milliseconds())
}

func runFrenet(ctx context.Context, log interface{ Info(msg string, args ...any); Error(msg string, args ...any) }, s *sync.Frenet) {
	stats, err := s.Run(ctx)
	if err != nil {
		log.Error("frenet sync failed", "err", err)
		return
	}
	log.Info("frenet sync done",
		"synced", stats.Synced,
		"errors", stats.Errors,
		"duration_ms", stats.Finished.Sub(stats.Started).Milliseconds())
}
