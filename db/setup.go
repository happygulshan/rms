package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() (*sql.DB, error) {

	connStr := "postgres://postgres:gulshan@localhost:5432/rmsappdb?sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, fmt.Errorf("error in opening db %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error in pinging db %w", err)
	}

	log.Println("successfully connected to db")
	return db, nil
}
