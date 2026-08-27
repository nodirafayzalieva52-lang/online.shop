package migrations

import (
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationFiles embed.FS

func Run(dsn string) error {
	sourceDriver, err := iofs.New(migrationFiles, ".")
	if err != nil {
		return fmt.Errorf("migrator: failed to create iofs source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dsn)
	if err != nil {
		return fmt.Errorf("migrator: failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Database auto-migration: no new changes")
			return nil
		}
		return fmt.Errorf("migrator: failed to apply migrations: %w", err)
	}

	log.Println("Database auto-migration: applied successfully")
	return nil
}
