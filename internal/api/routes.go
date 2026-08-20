package api

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func (ac *apiConfig) NewRouter() *chi.Mux {
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

	v1Router.Post("/users", ac.createUserHandler)
	v1Router.Get("/users", ac.authMiddleware(ac.getUserByApiKey))

	v1Router.Post("/feeds", ac.authMiddleware(ac.createFeedHandler))
	v1Router.Get("/feeds", ac.authMiddleware(ac.getFeedsHandler))
	v1Router.Get("/feeds/{feedId}", ac.authMiddleware(ac.getFeedByIdHandler))
	v1Router.Delete("/feeds/{feedId}", ac.authMiddleware(ac.deleteFeedHandler))

	v1Router.Post("/feed_follows", ac.authMiddleware(ac.createFeedFollowHandler))
	v1Router.Get("/feed_follows", ac.authMiddleware(ac.getFeedFollowsHandler))
	v1Router.Delete("/feed_follows/{feedFollowID}", ac.authMiddleware(ac.deleteFeedFollowHandler))

	v1Router.Get("/posts", ac.authMiddleware(ac.getPostsForUserHandler))

	return router
}
