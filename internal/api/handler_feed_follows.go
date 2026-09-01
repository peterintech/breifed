package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
)

func (ac *apiConfig) getFollowedFeedsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feeds, err := ac.DB.GetFeedsForUser(r.Context(), user.ID)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch followed feeds")
		return
	}
	response, err := ac.feedResponses(r.Context(), feeds)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch feed categories")
		return
	}
	jsonResponse(w, http.StatusOK, response)
}

func (ac *apiConfig) followFeedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feedID, err := uuid.Parse(chi.URLParam(r, "feedID"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid feed ID")
		return
	}
	feed, err := ac.DB.GetFeedByID(r.Context(), feedID)
	if err == sql.ErrNoRows {
		errorResponse(w, http.StatusNotFound, "feed not found")
		return
	}
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch feed")
		return
	}
	ac.followAndRespond(w, r, user, feed, http.StatusOK)
}

func (ac *apiConfig) unfollowFeedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feedID, err := uuid.Parse(chi.URLParam(r, "feedID"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid feed ID")
		return
	}
	if err := ac.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		UserID: user.ID, FeedID: feedID,
	}); err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not unfollow feed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
