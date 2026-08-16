package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaksymPushkash/coupons-service/internal/config"
	"github.com/MaksymPushkash/coupons-service/internal/coupon"
	"github.com/MaksymPushkash/coupons-service/internal/httpapi"
	"github.com/MaksymPushkash/coupons-service/internal/storage/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelInfo},
	))

	if err := run(logger); err != nil {
		logger.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	repository := postgres.NewCouponRepository(pool)
	service := coupon.NewService(repository)
	handler := httpapi.NewCouponHandler(service, logger)
	router := httpapi.NewRouter(handler, pool, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverError := make(chan error, 1)

	go func() {
		logger.Info("HTTP server started", "address", cfg.HTTPAddress)
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("http server: %w", err)

	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	logger.Info("HTTP server stopped")

	return nil
}
