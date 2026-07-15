package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"kickexchange/trading-service/internal/assets"
	"kickexchange/trading-service/internal/config"
	"kickexchange/trading-service/internal/db"
	"kickexchange/trading-service/internal/engineclient"
	graphqlapi "kickexchange/trading-service/internal/graphql"
	"kickexchange/trading-service/internal/graphql/generated"
	"kickexchange/trading-service/internal/pricefeed"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := db.WaitAndMigrate(ctx, cfg, log); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	client := engineclient.New(cfg, log)

	feed := pricefeed.New()
	client.OnAsyncEvent(func(e engineclient.AsyncEvent) {
		if e.Executed != nil {
			feed.Publish(e.AssetID, float64(e.Executed.PriceTicks))
		}
	})

	if err := client.Connect(ctx); err != nil {
		if errors.Is(err, engineclient.ErrVersionMismatch) {
			log.Error("engine rejected protocol version, exiting", "error", err)
		} else {
			log.Error("failed to connect to matching engine", "error", err)
		}
		os.Exit(1)
	}

	assetsRepo := assets.NewRepository(pool)
	resolver := graphqlapi.NewResolver(client, assetsRepo, feed, cfg, log)
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	mux.Handle("/graphql", srv)

	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", cfg.HTTPPort), Handler: mux}
	go func() {
		log.Info("graphql server listening", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("graphql server failed", "error", err)
		}
	}()

	log.Info("trading service ready")
	<-ctx.Done()

	log.Info("shutting down")
	httpServer.Shutdown(context.Background())
	client.Close()
}
