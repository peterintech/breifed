package api

import (
	"net/http"

	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/sessionauth"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (ac *apiConfig) authMiddleware(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := sessionauth.CurrentUser(r.Context(), ac.DB, r)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "could not verify session")
			return
		}
		if user == nil {
			errorResponse(w, http.StatusUnauthorized, "authentication required")
			return
		}
		handler(w, r, *user)
	}
}
