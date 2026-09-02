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
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
	router.Mount("/v1", v1Router)

	v1Router.Get("/health", readinessHandler)
	v1Router.Get("/err", errorHandler)
	v1Router.Post("/auth/register", ac.registerHandler)
	v1Router.Post("/auth/login", ac.loginHandler)
	v1Router.Post("/auth/logout", ac.logoutHandler)

	v1Router.Get("/categories", ac.getCategoriesHandler)
	v1Router.Get("/feeds", ac.getFeedsHandler)
	v1Router.Get("/feeds/{feedID}", ac.getFeedByIDHandler)
	v1Router.Post("/feeds", ac.authMiddleware(ac.createFeedHandler))

	v1Router.Get("/me", ac.authMiddleware(ac.getMeHandler))
	v1Router.Put("/me/categories", ac.authMiddleware(ac.replaceUserCategoriesHandler))
	v1Router.Get("/me/feeds", ac.authMiddleware(ac.getFollowedFeedsHandler))
	v1Router.Post("/me/feeds/{feedID}", ac.authMiddleware(ac.followFeedHandler))
	v1Router.Delete("/me/feeds/{feedID}", ac.authMiddleware(ac.unfollowFeedHandler))
	v1Router.Get("/posts", ac.getPostsHandler)

	return router
}
