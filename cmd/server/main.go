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
	"github.com/CookieG77/AppGDT-Server/internal/logging"
	"github.com/CookieG77/AppGDT-Server/internal/middleware"
	"github.com/CookieG77/AppGDT-Server/internal/ratelimit"
	"github.com/CookieG77/AppGDT-Server/internal/repository"
	"github.com/CookieG77/AppGDT-Server/internal/server"
	"github.com/CookieG77/AppGDT-Server/internal/service"
)

func main() {
	// Every log written with a request context also gets the request ID and
	// the authenticated user ID (see the logging package)
	logger := slog.New(logging.NewContextHandler(slog.NewJSONHandler(os.Stdout, nil)))
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
	noteRepository := repository.NewNoteRepository(pool)

	// Starting authentication tools
	passwordHasher := auth.NewPasswordHasher(cfg.HashingCfg.Cost)
	tokenManager := auth.NewTokenManager(cfg.TokenCfg.Secret, cfg.TokenCfg.TTL)

	// Limits on failed logins, kept in memory
	loginLimiters := service.LoginLimiters{
		ByEmail: ratelimit.New(cfg.LoginCfg.MaxPerEmail, cfg.LoginCfg.Window),
		ByIP:    ratelimit.New(cfg.LoginCfg.MaxPerIP, cfg.LoginCfg.Window),
	}

	// Starting services
	authService, err := service.NewAuthService(userRepository, passwordHasher, tokenManager, loginLimiters)
	if err != nil {
		return err
	}

	accountService := service.NewAccountService(userRepository, spaceRepository, noteRepository, passwordHasher, loginLimiters)
	spaceService := service.NewSpaceService(spaceRepository)
	noteService := service.NewNoteService(noteRepository, spaceRepository)

	// Creating handlers and server
	handlers := server.Handlers{
		Health: handler.NewHealthHandler(pool),
		Auth:   handler.NewAuthHandler(authService),
		User:   handler.NewUserHandler(authService, accountService),
		Space:  handler.NewSpaceHandler(spaceService),
		Note:   handler.NewNoteHandler(noteService),
	}

	requireAuth := middleware.Authenticate(tokenManager, userRepository)

	addr := cfg.Address + ":" + strconv.Itoa(cfg.Port)
	srv := server.New(addr, handlers, requireAuth)

	// HTTPS between the clients (the web client) and the API, if a certificate is configured
	scheme := "http"
	if cfg.TLSCfg.Enabled() {
		if srv.TLSConfig, err = server.LoadTLSConfig(cfg.TLSCfg.CertFile, cfg.TLSCfg.KeyFile); err != nil {
			return err
		}
		scheme = "https"
	}

	// Starting the server in a goroutine to prevent a freeze of the exit signal waiter
	servErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", addr, "url", scheme+"://"+addr)
		if err := listen(srv); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

// listen serves HTTPS when the server has a TLS config, HTTP otherwise.
func listen(srv *http.Server) error {
	if srv.TLSConfig != nil {
		// The certificate is already loaded in TLSConfig
		return srv.ListenAndServeTLS("", "")
	}
	return srv.ListenAndServe()
}
