// Package assets is the Postgres-backed mapping between our asset_id (the
// same id used on the matching engine's wire protocol) and each player's
// transfermarkt external_id.
package assets

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAlreadyExists = errors.New("assets: external_id already exists")
	ErrNotFound      = errors.New("assets: not found")
)

const uniqueViolation = "23505"

type Asset struct {
	AssetID      uint64
	ExternalID   string
	Symbol       string
	DisplayName  string
	InitialPrice float64
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create relies on the external_id unique constraint rather than a
// check-then-insert, so concurrent creates for the same player can't race.
func (r *Repository) Create(ctx context.Context, externalID, symbol, displayName string, initialPrice float64) (Asset, error) {
	var a Asset
	err := r.pool.QueryRow(ctx,
		`INSERT INTO assets (external_id, symbol, display_name, initial_price)
		 VALUES ($1, $2, $3, $4)
		 RETURNING asset_id, external_id, symbol, display_name, initial_price`,
		externalID, symbol, displayName, initialPrice,
	).Scan(&a.AssetID, &a.ExternalID, &a.Symbol, &a.DisplayName, &a.InitialPrice)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return Asset{}, ErrAlreadyExists
		}
		return Asset{}, fmt.Errorf("assets: create: %w", err)
	}
	return a, nil
}

// Search matches display_name or symbol against query, most recently
// added first. Only ever returns assets that exist in our own table - this
// is what makes home page search only show tradable players.
func (r *Repository) Search(ctx context.Context, query string) ([]Asset, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT asset_id, external_id, symbol, display_name, initial_price FROM assets
		 WHERE display_name ILIKE '%' || $1 || '%' OR symbol ILIKE '%' || $1 || '%'
		 ORDER BY asset_id DESC LIMIT 20`,
		query,
	)
	if err != nil {
		return nil, fmt.Errorf("assets: search: %w", err)
	}
	defer rows.Close()

	var out []Asset
	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.AssetID, &a.ExternalID, &a.Symbol, &a.DisplayName, &a.InitialPrice); err != nil {
			return nil, fmt.Errorf("assets: search scan: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("assets: search rows: %w", err)
	}
	return out, nil
}

func (r *Repository) GetByAssetID(ctx context.Context, assetID uint64) (Asset, error) {
	var a Asset
	err := r.pool.QueryRow(ctx,
		`SELECT asset_id, external_id, symbol, display_name, initial_price FROM assets WHERE asset_id = $1`,
		assetID,
	).Scan(&a.AssetID, &a.ExternalID, &a.Symbol, &a.DisplayName, &a.InitialPrice)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Asset{}, ErrNotFound
		}
		return Asset{}, fmt.Errorf("assets: get: %w", err)
	}
	return a, nil
}
