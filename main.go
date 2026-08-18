package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/peterintech/rssagg/internal/database"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	feed, err := urlToFeed("https://wagslane.dev/index.xml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(feed)

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
	api := apiConfig{
		DB: database.New(conn),
	}

	defer conn.Close()

	go startScraping(db, 10, time.Minute)

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
	router.Mount("/v1", v1Router)

	v1Router.Get("/health", readinessHandler)
	v1Router.Get("/err", errorHandler)

	v1Router.Post("/users", api.createUserHandler)
	v1Router.Get("/users", api.authMiddleware(api.getUserByApiKey))

	v1Router.Post("/feeds", api.authMiddleware(api.createFeedHandler))
	v1Router.Get("/feeds", api.authMiddleware(api.getFeedsHandler))
	v1Router.Get("/feeds/{feedId}", api.authMiddleware(api.getFeedByIdHandler))

	v1Router.Post("/feed_follows", api.authMiddleware(api.createFeedFollowHandler))
	v1Router.Get("/feed_follows", api.authMiddleware(api.getFeedFollowsHandler))
	v1Router.Delete("/feed_follows/{feedFollowID}", api.authMiddleware(api.deleteFeedFollowHandler))

	srv := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%s", portStr),
	}

	log.Printf("Server is running on port %s", portStr)
	log.Fatal(srv.ListenAndServe())
}
