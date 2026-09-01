package api

import (
	"net/http"

	"github.com/peterintech/briefed/internal/database"
)

const sessionCookieName = "briefed_session"

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (ac *apiConfig) authMiddleware(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			errorResponse(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := ac.DB.GetUserBySessionToken(r.Context(), cookie.Value)
		if err != nil {
			errorResponse(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		handler(w, r, user)
	}
}
