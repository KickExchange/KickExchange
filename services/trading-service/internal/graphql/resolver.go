package graphql

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	"kickexchange/trading-service/internal/engineclient"
	"kickexchange/trading-service/internal/graphql/generated"
	"kickexchange/trading-service/internal/graphql/model"
)

// Resolver holds the dependencies every resolver method needs. Order IDs are
// an in-memory counter for now - durable, DB-backed IDs land with Postgres.
type Resolver struct {
	Engine      engineclient.EngineClient
	Log         *slog.Logger
	nextOrderID atomic.Uint64
}

func NewResolver(engine engineclient.EngineClient, log *slog.Logger) *Resolver {
	return &Resolver{Engine: engine, Log: log}
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

// SubmitMarketOrder is the resolver for the submitMarketOrder field.
func (r *mutationResolver) SubmitMarketOrder(ctx context.Context, assetID uint64, side model.Side, shares int) (*model.SubmitResult, error) {
	engineSide, err := toEngineSide(side)
	if err != nil {
		return nil, err
	}

	orderID := r.nextOrderID.Add(1)
	result, err := r.Engine.SubmitMarket(assetID, engineclient.NewMarketPayload{
		OrderID: orderID,
		Side:    engineSide,
		Shares:  int64(shares),
	})

	if result.Rejected != nil {
		reason := result.Rejected.Reason.String()
		return &model.SubmitResult{OrderID: orderID, Accepted: false, RejectReason: &reason}, nil
	}
	if err != nil {
		return nil, err
	}
	return &model.SubmitResult{OrderID: orderID, Accepted: true}, nil
}

// Health is the resolver for the health field.
func (r *queryResolver) Health(ctx context.Context) (bool, error) {
	return true, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)
