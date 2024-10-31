package server

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/vkupriya/gophkeeper/internal/server/config"
	grpcserver "github.com/vkupriya/gophkeeper/internal/server/grpc"
	"github.com/vkupriya/gophkeeper/internal/server/storage"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const TimeoutShutdown time.Duration = 10 * time.Second
const TimeoutServerShutdown time.Duration = 5 * time.Second

func Start(logger *zap.Logger) error {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(zap.Error(err))
	}
	cfg.Logger = logger

	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancelCtx()

	g, ctx := errgroup.WithContext(rootCtx)

	_ = context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), TimeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		logger.Sugar().Error("failed to gracefully shutdown the service")
	})

	s, err := storage.NewPostgresDB(cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgresDB: %w", err)
	}

	g.Go(func() error {
		defer logger.Sugar().Info("closed GRPC server")

		if err := grpcserver.Run(ctx, s, cfg); err != nil {
			return fmt.Errorf("failed to run grpc server: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("go routines stopped with error: %w", err)
	}
	return nil
}
