// Command worker runs background jobs for Univer Tracker
// (Frenet polling, status sync, WhatsApp notifications via WAHA).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/univerbeauty777/univer-tracker/backend/internal/config"
	"github.com/univerbeauty777/univer-tracker/backend/pkg/logger"
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

	// TODO: initialize job processor (Asynq), register handlers, start polling.
	_ = ctx

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sig := <-stop
	log.Info("shutting down worker", "signal", sig.String())

	cancel()
	log.Info("worker stopped")
	return nil
}
