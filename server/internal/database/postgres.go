package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/lnardon/arete/internal/config"
)

type DB struct{ *sql.DB }

func New(cfg config.DatabaseConfig) (*DB, error) {
	db_url := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
	db, err := sql.Open("postgres", db_url)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := runMigrations(db, cfg.DBName); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}
	return &DB{db}, nil
}

func runMigrations(db *sql.DB, dbName string) error {
	migrationsPath := resolveMigrationsPath()
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, dbName, driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func resolveMigrationsPath() string {
	// Try relative to CWD first
	if cwd, err := os.Getwd(); err == nil {
		p := filepath.Join(cwd, "migrations")
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}

	// Try relative to executable
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "migrations")
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return "migrations"
}
