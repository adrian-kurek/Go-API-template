package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type DB struct {
	DBConnection *sql.DB
}

var migrationDirs = []string{
	"users",
	"refresh_tokens",
}

func NewDB(databaseLink, dbDriver string) (*DB, error) {
	dbConnection, err := sql.Open(dbDriver, databaseLink)
	if err != nil {
		return nil, err
	}

	dbConnection.SetConnMaxIdleTime(30 * time.Second)
	dbConnection.SetMaxOpenConns(20)
	dbConnection.SetMaxIdleConns(20)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = dbConnection.PingContext(ctx); err != nil {
		return nil, err
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	cmd := os.Args[len(os.Args)-1]

	for _, dirName := range migrationDirs {
		migrationsPath := filepath.Join(wd, "migrations", dirName)

		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			log.Printf("migrations directory not found, skipping: %s", migrationsPath)
			continue
		}

		migrateDriver, err := postgres.WithInstance(dbConnection, &postgres.Config{
			MigrationsTable: fmt.Sprintf("%s_migrations", dirName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create driver for %s: %w", dirName, err)
		}

		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", migrationsPath),
			"postgres",
			migrateDriver,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create migrate instance for %s: %w", dirName, err)
		}

		switch cmd {
		case "up":
			if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return nil, fmt.Errorf("migration up failed for %s: %w", dirName, err)
			}
			fmt.Printf("Migration up completed for %s\n", dirName)
		case "down":
			if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return nil, fmt.Errorf("migration down failed for %s: %w", dirName, err)
			}
			fmt.Printf("Migration down completed for %s\n", dirName)
		}
	}

	return &DB{
		DBConnection: dbConnection,
	}, nil
}

func (db *DB) Close() error {
	return db.DBConnection.Close()
}
