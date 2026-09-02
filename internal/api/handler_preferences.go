package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/preferences"
)

func (ac *apiConfig) replacePreferencesHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	var params struct {
		CategoryIDs []uuid.UUID `json:"category_ids"`
		FeedIDs     []uuid.UUID `json:"feed_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := preferences.Replace(r.Context(), ac.Conn, ac.DB, user.ID, params.CategoryIDs, params.FeedIDs); err != nil {
		errorResponse(w, http.StatusBadRequest, "one or more category or feed IDs are invalid")
		return
	}
	profile, err := ac.profileResponse(r.Context(), user)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "preferences were saved but the profile could not be loaded")
		return
	}
	jsonResponse(w, http.StatusOK, profile)
}
