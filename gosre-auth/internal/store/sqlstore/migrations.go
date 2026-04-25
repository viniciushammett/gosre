// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package sqlstore

import (
	"database/sql"
	"fmt"
)

// RunMigrations creates the users table if it does not yet exist.
// Must be called on startup before serving requests.
func RunMigrations(db *sql.DB) error {
	const ddl = `
IF NOT EXISTS (
    SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'users'
)
BEGIN
    CREATE TABLE users (
        id            VARCHAR(36)  NOT NULL,
        email         VARCHAR(255) NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        role          VARCHAR(50)  NOT NULL,
        created_at    DATETIME2    NOT NULL,
        CONSTRAINT PK_users PRIMARY KEY (id),
        CONSTRAINT UQ_users_email UNIQUE (email)
    )
END`
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
