// Command worker runs the periodic sync jobs that keep our local store in
// step with WooCommerce and Frenet. It owns no HTTP surface; everything is
// a ticker today and will graduate to a queue (Asynq) once we need fanout.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/univerbeauty777/univer-tracker/backend/internal/config"
	"github.com/univerbeauty777/univer-tracker/backend/internal/integrations"
	"github.com/univerbeauty777/univer-tracker/backend/internal/notifier"
	"github.com/univerbeauty777/univer-tracker/backend/internal/settings"
	"github.com/univerbeauty777/univer-tracker/backend/internal/store"
	"github.com/univerbeauty777/univer-tracker/backend/internal/sync"
	"github.com/univerbeauty777/univer-tracker/backend/pkg/logger"
)

// safeGo runs fn in a goroutine and recovers any panic so a single broken
// job can't take the worker process down. Panics are logged with stack.
func safeGo(log interface {
	Error(msg string, args ...any)
}, name string, fn func()) {
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic in worker goroutine",
					"job", name,
					"recover", fmt.Sprint(rec),
					"stack", string(debug.Stack()),
				)
			}
		}()
		fn()
	}()
}

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
	auditRepo := &store.Audit{Pool: pool}
	triggersRepo := &store.NotificationTriggers{Pool: pool, StoreID: defaultStoreID}

	settingsStore := settings.New(pool)
	resolver := integrations.New(settingsStore, cfg)
	wahaNotifier := notifier.New(resolver, "")
	dispatcher := &sync.TriggerDispatcher{
		Triggers: triggersRepo,
		Audit:    auditRepo,
		Orders:   ordersRepo,
		Sender:   wahaNotifier,
		Log:      log,
	}

	wcSync := &sync.WooCommerce{
		Store:        ordersRepo,
		Shipments:    shipmentsRepo,
		State:        stateRepo,
		Integrations: resolver,
		StoreID:      defaultStoreID,
		Dispatcher:   dispatcher,
		Log:          log,
	}
	frenetSync := &sync.Frenet{
		Shipments:    shipmentsRepo,
		Events:       eventsRepo,
		Integrations: resolver,
		BatchSize:    50,
		Dispatcher:   dispatcher,
		Log:          log,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// One-shot: replay tracking_events into stage timestamps so existing
	// shipments inherit per-etapa data without waiting for the next
	// Frenet event.
	safeGo(log, "backfill_stages", func() {
		bf := &sync.BackfillStages{Pool: pool, Shipments: shipmentsRepo, Log: log}
		bctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		n, err := bf.Run(bctx)
		if err != nil {
			log.Error("backfill stages failed", "err", err)
			return
		}
		log.Info("backfill stages done", "shipments_updated", n)
	})

	// One-shot: normalise legacy carrier values that hold shipping_method
	// titles ("Frete grátis", "Expresso") so the dashboard groups them
	// under the actual carrier (Correios - PAC / SEDEX, Jadlog, …).
	safeGo(log, "backfill_carriers", func() {
		cb := &sync.BackfillCarriers{Pool: pool, Log: log}
		bctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		n, err := cb.Run(bctx)
		if err != nil {
			log.Error("backfill carriers failed", "err", err)
			return
		}
		log.Info("backfill carriers done", "shipments_updated", n)
	})

	// One-shot: hide synthetic test orders the dashboard shouldn't show
	// ("E2E Buyer" from end-to-end testing, "webhook testa" from
	// webhook validation). Idempotent — already-hidden rows stay
	// hidden and are not re-counted.
	safeGo(log, "hide_synthetic_orders", func() {
		bctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		patterns := []string{"E2E Buyer", "webhook testa", "webhook test"}
		var total int64
		for _, p := range patterns {
			n, err := ordersRepo.HideByCustomerNameLike(bctx, defaultStoreID, p)
			if err != nil {
				log.Error("hide synthetic orders failed", "pattern", p, "err", err)
				continue
			}
			total += n
		}
		if total > 0 {
			log.Info("hid synthetic test orders", "rows", total)
		}
	})

	// Run an initial pass right away so the first deploy has data fast.
	safeGo(log, "wc_sync_initial", func() { runWC(ctx, log, wcSync) })
	safeGo(log, "frenet_sync_initial", func() { runFrenet(ctx, log, frenetSync) })

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
			safeGo(log, "wc_sync", func() { runWC(ctx, log, wcSync) })
		case <-frenetTicker.C:
			safeGo(log, "frenet_sync", func() { runFrenet(ctx, log, frenetSync) })
		}
	}
}

func runWC(ctx context.Context, log interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}, s *sync.WooCommerce) {
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

func runFrenet(ctx context.Context, log interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}, s *sync.Frenet) {
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
