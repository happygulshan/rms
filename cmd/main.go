package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"rms/db"
	"rms/server"

	"github.com/joho/godotenv"
)

func SeedRoles(db *sql.DB) error {
	query := `INSERT INTO roles (name, priority) VALUES
		('admin', 3),
		('subadmin', 2),
		('user', 1)
		ON CONFLICT (name) DO NOTHING;
		`
	_, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error in seeding roles table %w", err)
	}
	return nil
}

func main() {

	err := godotenv.Load("../.env")

	if err != nil {
		log.Println("error with env loading")
	}

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	defer database.Close()

	db.RunMigrations(database)
	err = SeedRoles(database)

	if err != nil {
		log.Fatalf("%w", err)
	}

	r := server.InitRoutes(database)

	log.Println("Server starting on :8080")

	http.ListenAndServe(":8080", r)

}
