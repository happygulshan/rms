package db

import (
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // required
	_ "github.com/lib/pq"
)

func RunMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Migration driver error: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///home/gulshan-kumar/Documents/rms/db/migrations",
		"postgres",
		driver,
	)

	if err != nil {
		log.Fatalf("Migration init error: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration run error: %v", err)
	}

	log.Println("Database migrations applied successfully.")
}
