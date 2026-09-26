package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/config"
	"github.com/CookieG77/AppGDT-Server/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("Server encountered an error and was stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Loading the server config
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	addr := cfg.Address + ":" + strconv.Itoa(cfg.Port)
	srv := server.New(addr)

	// Starting the server in a goroutine to prevent a freeze of the exit signal waiter
	servErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			servErr <- err
		}
	}()

	// Waiting for server error or exit signal
	select {
	case err := <-servErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("server stopping")
	}

	// Clean stop: gives time for request to end before stopping the app
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	logger.Info("server stopped")
	return nil
}
