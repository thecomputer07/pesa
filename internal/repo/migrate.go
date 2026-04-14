package repo

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"github.com/thecomputer07/pesa/migrations"
)

// Migrate runs all pending Goose migrations against the given DB.
func Migrate(db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

// MigrateDown reverts the last migration.
func MigrateDown(db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Down(db, ".")
}

// MigrateStatus prints the migration status to stdout via goose.
func MigrateStatus(db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Status(db, ".")
}
