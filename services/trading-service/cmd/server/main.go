package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"kickexchange/trading-service/internal/config"
	"kickexchange/trading-service/internal/engineclient"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()

	client := engineclient.New(cfg, log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := client.Connect(ctx); err != nil {
		if errors.Is(err, engineclient.ErrVersionMismatch) {
			log.Error("engine rejected protocol version, exiting", "error", err)
		} else {
			log.Error("failed to connect to matching engine", "error", err)
		}
		os.Exit(1)
	}

	log.Info("trading service ready")
	<-ctx.Done()

	log.Info("shutting down")
	client.Close()
}
