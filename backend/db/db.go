// Package db provides database connection management.
package db

import (
	"fmt"
	"strings"

	"flickey/go-backend/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open creates a new GORM DB connection from the configured DSN.
// It converts the postgresql:// DSN format to the DSN expected by pgx.
func Open(cfg *config.Settings) (*gorm.DB, error) {
	dsn := convertDSN(cfg.DatabaseURL)

	logLevel := logger.Warn
	if cfg.IsDev() {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: false,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		// Disable automatic transactions for explicit control.
		SkipDefaultTransaction: false,
	})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting sql.DB: %w", err)
	}

	// Connection pool settings.
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	// Auto-migrate tables if needed
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto-migrating tables: %w", err)
	}

	return db, nil
}

// AutoMigrate ensures all tables, indexes, and constraints exist in PostgreSQL.
func AutoMigrate(database *gorm.DB) error {
	return database.AutoMigrate(
		&User{},
		&Media{},
		&ListingDraft{},
		&Listing{},
		&ListingAmenity{},
		&AuditLog{},
		&VerificationRequest{},
		&Report{},
		&UserRestriction{},
		&Notification{},
	)
}

// convertDSN converts postgresql://, postgresql+asyncpg://, or postgres:// URL to pgx DSN format.
// GORM's postgres driver accepts both URL formats natively; this is a passthrough.
func convertDSN(rawURL string) string {
	if strings.HasPrefix(rawURL, "postgresql+asyncpg://") {
		return "postgres://" + rawURL[len("postgresql+asyncpg://"):]
	}
	if strings.HasPrefix(rawURL, "postgresql://") {
		return "postgres://" + rawURL[len("postgresql://"):]
	}
	return rawURL
}
