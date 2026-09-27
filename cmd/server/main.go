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

	"github.com/CookieG77/AppGDT-Server/internal/auth"
	"github.com/CookieG77/AppGDT-Server/internal/config"
	"github.com/CookieG77/AppGDT-Server/internal/database"
	"github.com/CookieG77/AppGDT-Server/internal/handler"
	"github.com/CookieG77/AppGDT-Server/internal/middleware"
	"github.com/CookieG77/AppGDT-Server/internal/repository"
	"github.com/CookieG77/AppGDT-Server/internal/server"
	"github.com/CookieG77/AppGDT-Server/internal/service"
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

	dbURL := database.BuildDatabaseURL(cfg.DatabaseCfg)

	// Applying migrations if needed
	if err := database.Migrate(dbURL); err != nil {
		return err
	}

	// Establishing a connection with the Database
	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Creating Repositories
	userRepository := repository.NewUserRepository(pool)
	spaceRepository := repository.NewSpaceRepository(pool)
	//noteRepository := repository.NewNoteRepository(pool)

	// Starting authentication tools
	passwordHasher := auth.NewPasswordHasher(cfg.HashingCfg.Cost)
	tokenManager :=	auth.NewTokenManager(cfg.TokenCfg.Secret, cfg.TokenCfg.TTL)

	// Starting services
	authService, err := service.NewAuthService(userRepository, passwordHasher, tokenManager)
	if err != nil {
		return err
	}

	spaceService := service.NewSpaceService(spaceRepository)

	// Creating handlers and server
	handlers := server.Handlers{
		Auth: handler.NewAuthHandler(authService),
		User: handler.NewUserHandler(authService),
		Space: handler.NewSpaceHandler(spaceService),
	}

	requireAuth := middleware.Authenticate(tokenManager)

	addr := cfg.Address + ":" + strconv.Itoa(cfg.Port)
	srv := server.New(addr, handlers, requireAuth)

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
