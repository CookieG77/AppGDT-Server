// File containing the required functions to migrate the database

package database

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/CookieG77/AppGDT-Server/migrations"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func Migrate(databaseURL string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("parsing database URL: %w", err)
	}
	u.Scheme = "pgx5"

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("loading migration files: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, u.String())
	if err != nil {
		return fmt.Errorf("initializing migrations: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("applying migrations: %w", err)
	}

	return nil
}
