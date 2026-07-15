package graphql

import (
	"time"

	coderws "github.com/coder/websocket"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/ast"

	"kickexchange/trading-service/internal/graphql/generated"
)

// NewServer mirrors handler.NewDefaultServer but configures the websocket
// transport to accept cross-origin connections - the frontend is served
// from a different origin, and there's no auth/session state yet for an
// origin check to protect.
func NewServer(resolver *Resolver) *handler.Server {
	es := generated.NewExecutableSchema(generated.Config{Resolvers: resolver})
	srv := handler.New(es)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Implementation: transport.CoderWebsocketImplementation{
			AcceptOptions: coderws.AcceptOptions{InsecureSkipVerify: true},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

	return srv
}
