package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/peterintech/briefed/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func GetApiKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header is missing")
	}

	// Split the header into two parts
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid Authorization header format")
	}
	if parts[0] != "ApiKey" {
		return "", fmt.Errorf("Authorization header must start with 'ApiKey'")
	}

	return parts[1], nil
}

func (ac *apiConfig) authMiddleware(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apikey, err := GetApiKey(r.Header)
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
