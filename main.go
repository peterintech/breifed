package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/peterintech/briefed/internal/api"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/scrapper"
)

func main() {
	godotenv.Load(".env")

	portStr := os.Getenv("PORT")
	if portStr == "" {
		log.Fatal("PORT environment variable is not set")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("can't connect to the database:", err)
	}

	db := database.New(conn)
	defer conn.Close()

	api := api.New(db)
	router := api.NewRouter()

	srv := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%s", portStr),
	}
	const collectionConcurrency = 10
	const collectionInterval = time.Minute
	go scrapper.Start(db, collectionConcurrency, collectionInterval)

	log.Printf("Server is running on port %s", portStr)
	log.Fatal(srv.ListenAndServe())
}
