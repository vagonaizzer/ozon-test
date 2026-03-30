package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"go.uber.org/zap"

	"github.com/vagonaizer/ozon-test-assignment/api/graphql/generated"
	comment_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/comment"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/dataloader"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/resolver"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/server/middleware"
)

type Server struct {
	httpServer *http.Server
	log        *zap.Logger
}

func New(
	addr string,
	res *resolver.Resolver,
	commentRepo comment_iface.CommentRepository,
	log *zap.Logger,
) *Server {
	schema := generated.NewExecutableSchema(generated.Config{Resolvers: res})

	srv := handler.New(schema)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Расширения
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	loaders := dataloader.NewLoaders(commentRepo)

	mux := http.NewServeMux()

	withDataLoader := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := dataloader.Attach(r.Context(), loaders)
		srv.ServeHTTP(w, r.WithContext(ctx))
	})

	mux.Handle("/query", withDataLoader)
	mux.Handle("/playground", playground.Handler("OzonPosts GraphQL", "/query"))

	mux.Handle("/", http.FileServer(http.Dir("web")))

	var h http.Handler = mux
	h = middleware.Recovery(log)(h)
	h = middleware.Logger(log)(h)
	h = middleware.CORS(h)
	h = middleware.WithLogger(log)(h)

	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           h,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      0,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		},
		log: log,
	}
}

func (s *Server) Run() error {
	s.log.Info("GraphQL server starting", zap.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down server")
	return s.httpServer.Shutdown(ctx)
}
