package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"

	"kickexchange/trading-service/internal/config"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const migrateRetryAttempts = 10

// RunMigrations applies all pending migrations. Safe to call on every
// startup - a no-op once the schema is current.
func RunMigrations(cfg config.Config) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("db: load migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, cfg.MigrateDSN())
	if err != nil {
		return fmt.Errorf("db: init migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("db: apply migrations: %w", err)
	}
	return nil
}

// WaitAndMigrate retries RunMigrations while Postgres finishes its own
// startup (the official image restarts itself once during init, so
// depends_on alone doesn't guarantee it's accepting connections yet).
func WaitAndMigrate(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	var err error
	for attempt := 1; attempt <= migrateRetryAttempts; attempt++ {
		if err = RunMigrations(cfg); err == nil {
			return nil
		}
		log.Warn("postgres not ready yet, retrying migrations", "attempt", attempt, "error", err)
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}
