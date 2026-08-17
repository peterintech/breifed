package main

import (
	"net/http"

	"github.com/peterintech/rssagg/internal/auth"
	"github.com/peterintech/rssagg/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (ac *apiConfig) authMiddleware(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apikey, err := auth.GetApiKey(r.Header)
		if err != nil {
			errorResponse(w, 403, "Auth error: "+err.Error())
			return
		}

		user, err := ac.DB.GetUserByApiKey(r.Context(), apikey)
		if err != nil {
			errorResponse(w, 403, "Auth error: "+err.Error())
			return
		}

		handler(w, r, user)
	}
}
