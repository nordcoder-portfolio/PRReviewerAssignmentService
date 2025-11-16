package main

import (
	"avito_test/config"
	"avito_test/internal/app"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	level := parseLogLevel(cfg.Log.Level)
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})).With(
		slog.String("service", "pr-reviewer-service"),
		slog.String("env", cfg.Env),
	)

	logger.Info("config loaded",
		slog.String("http_addr", cfg.HTTP.Addr),
		slog.String("log_level", cfg.Log.Level),
	)

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New(rootCtx, cfg, logger)
	if err != nil {
		logger.Error("failed to init app", slog.String("error", err.Error()))
		return err
	}

	go func() {
		if runErr := application.Run(); runErr != nil {
			logger.Error("server run error", slog.String("error", runErr.Error()))
			stop()
		}
	}()

	logger.Info("application started",
		slog.String("addr", cfg.HTTP.Addr),
	)

	<-rootCtx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		return err
	}

	logger.Info("application stopped cleanly")
	return nil
}

func parseLogLevel(lvl string) slog.Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info":
		fallthrough
	default:
		return slog.LevelInfo
	}
}
