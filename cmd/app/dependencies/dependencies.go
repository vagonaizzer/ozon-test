package dependencies

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	app_comment "github.com/vagonaizer/ozon-test-assignment/internal/application/comment"
	app_post "github.com/vagonaizer/ozon-test-assignment/internal/application/post"
	"github.com/vagonaizer/ozon-test-assignment/internal/config"
	comment_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/comment"
	post_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/post"
	"github.com/vagonaizer/ozon-test-assignment/internal/infrastructure/repository/memory"
	"github.com/vagonaizer/ozon-test-assignment/internal/infrastructure/repository/postgres"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/resolver"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/server"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/subscription"
	"github.com/vagonaizer/ozon-test-assignment/pkg/database"
)


type App struct {
	Server  *server.Server
	Cleanup func()
}

func Wire(ctx context.Context, cfg *config.Config, log *zap.Logger) (*App, error) {
	postRepo, commentRepo, cleanup, err := buildRepositories(ctx, cfg, log)
	if err != nil {
		return nil, err
	}

	subManager := subscription.New()

	postService    := app_post.NewService(postRepo)
	commentService := app_comment.NewService(commentRepo, postRepo, subManager)

	res := resolver.NewResolver(postService, commentService, subManager)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv  := server.New(addr, res, commentRepo, log)

	return &App{Server: srv, Cleanup: cleanup}, nil
}

func buildRepositories(
	ctx context.Context,
	cfg *config.Config,
	log *zap.Logger,
) (post_iface.PostRepository, comment_iface.CommentRepository, func(), error) {
	noop := func() {}

	switch cfg.Storage.Type {
	case config.StorageMemory:
		log.Info("using in-memory storage")
		return memory.NewPostRepository(), memory.NewCommentRepository(), noop, nil

	case config.StoragePostgres:
		log.Info("using PostgreSQL storage")
		pgCfg := database.Config{
			Host:     cfg.Storage.Postgres.Host,
			Port:     cfg.Storage.Postgres.Port,
			User:     cfg.Storage.Postgres.User,
			Password: cfg.Storage.Postgres.Password,
			DBName:   cfg.Storage.Postgres.DBName,
			SSLMode:  cfg.Storage.Postgres.SSLMode,
		}
		pool, err := database.NewPool(ctx, pgCfg)
		if err != nil {
			return nil, nil, noop, fmt.Errorf("connect to postgres: %w", err)
		}
		return postgres.NewPostRepository(pool),
			postgres.NewCommentRepository(pool),
			pool.Close,
			nil

	default:
		return nil, nil, noop, fmt.Errorf(
			"unknown storage type %q (allowed: memory, postgres)", cfg.Storage.Type,
		)
	}
}
