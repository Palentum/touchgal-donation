package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"donation-site/services/api/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Connect(cfg config.Config) (*gorm.DB, error) {
	var lastErr error
	for attempt := 0; attempt < 30; attempt++ {
		gdb, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
			NowFunc: func() time.Time { return time.Now().UTC() },
		})
		if err == nil {
			sqlDB, dbErr := gdb.DB()
			if dbErr == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					return gdb, nil
				} else {
					lastErr = pingErr
				}
			} else {
				lastErr = dbErr
			}
		} else {
			lastErr = err
		}
		time.Sleep(time.Second)
	}
	return nil, lastErr
}

func Migrate(ctx context.Context, gdb *gorm.DB) error {
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	if err := ensureMigrationTable(ctx, sqlDB); err != nil {
		return err
	}
	applied, err := migrationApplied(ctx, sqlDB, "001_init")
	if err != nil {
		return err
	}
	if applied {
		return nil
	}
	body, err := migrationsFS.ReadFile("migrations/001_init.up.sql")
	if err != nil {
		return err
	}
	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, string(body)); err != nil {
		return fmt.Errorf("apply migration 001_init: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES ($1, now())`, "001_init"); err != nil {
		return err
	}
	return tx.Commit()
}

func ensureMigrationTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`)
	return err
}

func migrationApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists)
	return exists, err
}
