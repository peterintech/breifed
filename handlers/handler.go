package handlers

import (
	"database/sql"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi"
	"github.com/peterintech/briefed/internal/database"
)

type Handler struct {
	DB   *database.Queries
	Conn *sql.DB
}

func New(db *database.Queries, conn *sql.DB) *Handler { return &Handler{DB: db, Conn: conn} }

func (h *Handler) RegisterRoutes(router *chi.Mux) {
	router.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	router.Get("/", h.home)
	router.Get("/partials/posts", h.posts)
}

func render(w http.ResponseWriter, r *http.Request, status int, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "could not render page", http.StatusInternalServerError)
	}
}
