// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package sqlstore implements gosre-auth store interfaces against Azure SQL.
package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	mssql "github.com/microsoft/go-mssqldb"

	"github.com/gosre/gosre-auth/internal/domain"
	"github.com/gosre/gosre-auth/internal/store"
)

// SQLUserStore implements store.UserStore against Azure SQL.
type SQLUserStore struct {
	db *sql.DB
}

// NewSQLUserStore returns a SQLUserStore backed by the given *sql.DB.
func NewSQLUserStore(db *sql.DB) *SQLUserStore {
	return &SQLUserStore{db: db}
}

// Create inserts a new user row. Returns store.ErrEmailTaken on unique constraint violation.
func (s *SQLUserStore) Create(ctx context.Context, u domain.User) error {
	const q = `INSERT INTO users (id, email, password_hash, role, created_at)
	           VALUES (@id, @email, @password_hash, @role, @created_at)`
	_, err := s.db.ExecContext(ctx, q,
		sql.Named("id", u.ID),
		sql.Named("email", u.Email),
		sql.Named("password_hash", u.PasswordHash),
		sql.Named("role", string(u.Role)),
		sql.Named("created_at", u.CreatedAt),
	)
	if err != nil {
		var mssqlErr mssql.Error
		if errors.As(err, &mssqlErr) && (mssqlErr.Number == 2627 || mssqlErr.Number == 2601) {
			return store.ErrEmailTaken
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetByEmail returns the user with the given email or store.ErrUserNotFound.
func (s *SQLUserStore) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	const q = `SELECT id, email, password_hash, role, created_at FROM users WHERE email = @email`
	return s.scanRow(s.db.QueryRowContext(ctx, q, sql.Named("email", email)))
}

// GetByID returns the user with the given ID or store.ErrUserNotFound.
func (s *SQLUserStore) GetByID(ctx context.Context, id string) (domain.User, error) {
	const q = `SELECT id, email, password_hash, role, created_at FROM users WHERE id = @id`
	return s.scanRow(s.db.QueryRowContext(ctx, q, sql.Named("id", id)))
}

func (s *SQLUserStore) scanRow(row *sql.Row) (domain.User, error) {
	var u domain.User
	var role string
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, store.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("scan user row: %w", err)
	}
	u.Role = domain.Role(role)
	return u, nil
}
