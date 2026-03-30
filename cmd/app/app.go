package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/vagonaizer/ozon-test-assignment/cmd/app/dependencies"
	"github.com/vagonaizer/ozon-test-assignment/internal/config"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/server"
	"github.com/vagonaizer/ozon-test-assignment/pkg/logger"
)

const shutdownTimeout = 10 * time.Second

func Run() error {
	cfgPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, err := logger.New(cfg.Logger.Level, cfg.Logger.Env)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer log.Sync() 

	app, err := dependencies.Wire(context.Background(), cfg, log)
	if err != nil {
		return fmt.Errorf("wire dependencies: %w", err)
	}
	defer app.Cleanup()

	return gracefulShutdown(app.Server, log)
}


func gracefulShutdown(srv *server.Server, log *zap.Logger) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Run() }()

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		log.Sugar().Infof("received signal %s, shutting down gracefully", sig)
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}
