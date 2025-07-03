package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"strings"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() (*sql.DB, error) {
	typeOFDB := os.Getenv("DB")
	host := os.Getenv("HOST")
	user := os.Getenv("DB_USER")
	name := os.Getenv("DB_NAME")
	pass := os.Getenv("DB_PASS")
	port := os.Getenv("DB_PORT")

	var connStr strings.Builder

	fmt.Fprintf(&connStr, "%s://%s:%s@%s:%s/%s?sslmode=disable", typeOFDB, user, pass, host, port, name)
	db, err := sql.Open("postgres", connStr.String())

	if err != nil {
		return nil, fmt.Errorf("error in opening db %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error in pinging db %w", err)
	}

	log.Println("successfully connected to db")
	return db, nil
}
