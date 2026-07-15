package graphql

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"

	"kickexchange/trading-service/internal/assets"
	"kickexchange/trading-service/internal/config"
	"kickexchange/trading-service/internal/engineclient"
	"kickexchange/trading-service/internal/graphql/model"
	"kickexchange/trading-service/internal/pricefeed"
	"kickexchange/trading-service/internal/transfermarkt"
)

type Resolver struct {
	Engine      engineclient.EngineClient
	Assets      *assets.Repository
	Transfer    *transfermarkt.Client
	Feed        *pricefeed.Feed
	Log         *slog.Logger
	nextOrderID atomic.Uint64
}

func NewResolver(engine engineclient.EngineClient, assetsRepo *assets.Repository, feed *pricefeed.Feed, cfg config.Config, log *slog.Logger) *Resolver {
	return &Resolver{
		Engine:   engine,
		Assets:   assetsRepo,
		Transfer: transfermarkt.New(cfg.TransfermarktAPIURL),
		Feed:     feed,
		Log:      log,
	}
}

func toEngineSide(s model.Side) (engineclient.Side, error) {
	switch s {
	case model.SideBuy:
		return engineclient.SideBuy, nil
	case model.SideSell:
		return engineclient.SideSell, nil
	default:
		return 0, fmt.Errorf("graphql: unknown side %q", s)
	}
}

// toSymbol derives a ticker-style symbol from a player's name since
// transfermarkt doesn't provide one, e.g. "Lionel Messi" -> "LIONEL_MESSI".
func toSymbol(name string) string {
	return strings.ToUpper(strings.Join(strings.Fields(name), "_"))
}

func toModelAsset(a assets.Asset) *model.Asset {
	return &model.Asset{
		AssetID:      a.AssetID,
		ExternalID:   a.ExternalID,
		Symbol:       a.Symbol,
		DisplayName:  a.DisplayName,
		InitialPrice: a.InitialPrice,
	}
}
