package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/isaacjstriker/reliable-webhooks/internal/config"
	"github.com/isaacjstriker/reliable-webhooks/internal/httpserver"
	"github.com/isaacjstriker/reliable-webhooks/internal/log"
	"github.com/isaacjstriker/reliable-webhooks/internal/processing"
	"github.com/isaacjstriker/reliable-webhooks/internal/repository"
	"github.com/isaacjstriker/reliable-webhooks/internal/storage"
)

func main() {
	cfg := config.Load()

	logger := log.New(cfg.Env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := storage.NewPostgres(ctx, cfg.DBURL)
	if err != nil {
		logger.Fatal("db_connect_error", "error", err)
	}
	defer db.Close(context.Background())

	eventRepo := repository.NewEventRepository(db)
	dispatcher := processing.NewDispatcher(eventRepo, logger)

	app := httpserver.New(cfg, logger, eventRepo, dispatcher)

	go dispatcher.Run(ctx)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			logger.Error("server_listen_error", "error", err)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logger.Info("shutdown_initiated")

	shutdownCtx, scancel := context.WithTimeout(context.Background(), cfg.GracefulTimeout)
	defer scancel()

	if err := app.Shutdown(); err != nil {
		logger.Error("server_shutdown_error", "error", err)
	}

	time.Sleep(500 * time.Millisecond)

	logger.Info("shutdown_complete")
}