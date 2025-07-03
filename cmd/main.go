package main

import (
	"log"
	"net/http"
	"rms/db"
	"rms/server"

	"github.com/joho/godotenv"
)

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

	if err != nil {
		log.Fatalf("%v", err)
	}

	r := server.InitRoutes(database)

	log.Println("Server starting on :8080")

	http.ListenAndServe(":8080", r)
}