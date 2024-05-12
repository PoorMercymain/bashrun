package repository

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//go:generate mockgen -destination=mocks/migrator_mock.gen.go -package=mocks . Migrator
type Migrator interface {
	Up() error
	Close() (sourceErr error, databaseErr error)
}

func ApplyMigrations(m Migrator) error {
	const logPrefix = "repository.ApplyMigrations"

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	sourceErr, databaseErr := m.Close()

	if sourceErr != nil {
		return fmt.Errorf("%s: %w", logPrefix, sourceErr)
	}

	if databaseErr != nil {
		return fmt.Errorf("%s: %w", logPrefix, databaseErr)
	}

	return nil
}
