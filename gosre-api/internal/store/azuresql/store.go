// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package azuresql

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	sqlservermigrate "github.com/golang-migrate/migrate/v4/database/sqlserver"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/microsoft/go-mssqldb"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store holds the Azure SQL database connection.
type Store struct {
	db *sql.DB
}

// scanner abstracts *sql.Row and *sql.Rows to share scan helpers.
type scanner interface {
	Scan(dest ...any) error
}

// New opens an Azure SQL database at dsn, verifies connectivity, and runs all pending migrations.
func New(dsn string) (*Store, error) {
	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return nil, fmt.Errorf("azuresql: open: %w", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("azuresql: ping: %w", err)
	}

	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close releases the database connection.
func (s *Store) Close() error { return s.db.Close() }

func runMigrations(db *sql.DB) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("azuresql: create migration source: %w", err)
	}

	driver, err := sqlservermigrate.WithInstance(db, &sqlservermigrate.Config{})
	if err != nil {
		return fmt.Errorf("azuresql: create migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlserver", driver)
	if err != nil {
		return fmt.Errorf("azuresql: init migrations: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("azuresql: run migrations: %w", err)
	}

	return nil
}
