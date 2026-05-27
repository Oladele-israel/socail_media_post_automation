package database

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	*sqlx.DB
}

//  type DB struct { *sqlx.DB }
// We define our own type called DB. It embeds a pointer to an sqlx.DB.
// Embedding means our DB automatically inherits all the methods of *sqlx.DB (like Query, Exec, Get, etc.). So we can use a database.DB just like an *sqlx.DB, but we can also add our own custom methods later if we want.

func New() (*DB, error) {
	connectionStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sqlx.Connect("postgres", connectionStr)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database connected and migrations applied")
	return &DB{db}, nil
}

func runMigrations(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})

	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migrations, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)

	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := migrations.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
