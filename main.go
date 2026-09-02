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
	webhandlers "github.com/peterintech/briefed/handlers"
	"github.com/peterintech/briefed/internal/api"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/scrapper"
)

func main() {
	godotenv.Load(".env")

	port := os.Getenv("PORT")
	if port == "" {
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
	defer conn.Close()

	db := database.New(conn)
	apiConfig := api.New(db, conn)
	router := apiConfig.NewRouter()
	webhandlers.New(db, conn).RegisterRoutes(router)

	server := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%s", port),
	}
	const collectionConcurrency = 10
	const collectionInterval = time.Minute
	go scrapper.Start(db, collectionConcurrency, collectionInterval)

	log.Printf("Server is running on port %s", port)
	log.Fatal(server.ListenAndServe())
}
