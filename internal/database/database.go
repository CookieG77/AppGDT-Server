package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func BuildDatabaseURL(databaseConfig *config.DatabaseConfig) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(databaseConfig.User, databaseConfig.Password),
		Host:     net.JoinHostPort(databaseConfig.Host, strconv.Itoa(databaseConfig.Port)),
		Path:     databaseConfig.Database,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool : %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database : %w", err)
	}

	return pool, nil
}
