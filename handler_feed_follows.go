package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"

	"github.com/peterintech/rssagg/internal/database"
)

func (ac *apiConfig) createFeedFollowHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		FeedID uuid.UUID `json:"feed_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		errorResponse(w, 400, fmt.Sprint("Error parsing JSON:", err))
		return
	}

	feedFollow, err := ac.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		FeedID:    params.FeedID,
		UserID:    user.ID,
	})

	if err != nil {
		errorResponse(w, 500, fmt.Sprint("Error creating feed follow:", err))
		return
	}

	jsonResponse(w, 201, databaseFeedFollowToFeedFollow(feedFollow))
}

func (ac *apiConfig) getFeedFollowsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollows, err := ac.DB.GetFeedFollows(r.Context(), user.ID)
	if err != nil {
		errorResponse(w, 400, fmt.Sprint("Error fetching feed follows:", err))
		return
	}

	jsonResponse(w, 200, databaseFeedFollowsToFeedFollows(feedFollows))
}

func (ac *apiConfig) deleteFeedFollowHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollowIDStr := chi.URLParam(r, "feedFollowID")
	feedFollowID, err := uuid.Parse(feedFollowIDStr)
	if err != nil {
		errorResponse(w, 400, fmt.Sprintf("couldn't parse feed follow id: %v", err))
		return
	}

	err = ac.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		ID:     feedFollowID,
		UserID: user.ID,
	})
	if err != nil {
		errorResponse(w, 500, fmt.Sprintf("Error deleting feed follow: %v", err))
		return
	}
	jsonResponse(w, 200, map[string]string{"message": "Feed follow deleted successfully"})
}
