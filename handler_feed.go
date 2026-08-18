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

func (ac *apiConfig) createFeedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		errorResponse(w, 400, fmt.Sprint("Error parsing JSON:", err))
		return
	}

	feed, err := ac.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
		Url:       params.Url,
		UserID:    user.ID,
	})

	if err != nil {
		errorResponse(w, 500, fmt.Sprint("Error creating feed:", err))
		return
	}

	jsonResponse(w, 201, databaseCreateFeedRowToFeed(feed))
}

func (ac *apiConfig) getFeedsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feeds, err := ac.DB.GetFeedsByUserId(r.Context(), user.ID)
	if err != nil {
		errorResponse(w, 500, fmt.Sprint("Error fetching feeds:", err))
		return
	}

	jsonResponse(w, 200, databaseGetFeedsByUserIdRowToFeeds(feeds))
}

func (ac *apiConfig) getFeedByIdHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feedId := chi.URLParam(r, "feedId")
	if feedId == "" {
		errorResponse(w, 400, "Missing feed ID")
		return
	}

	feedUUID, err := uuid.Parse(feedId)
	if err != nil {
		errorResponse(w, 400, fmt.Sprint("Invalid feed ID:", err))
		return
	}

	feed, err := ac.DB.GetFeedById(r.Context(), feedUUID)
	if err != nil {
		errorResponse(w, 500, fmt.Sprint("Error fetching feed:", err))
		return
	}

	if feed.UserID != user.ID {
		errorResponse(w, 403, "You do not have access to this feed")
		return
	}

	jsonResponse(w, 200, databaseGetFeedByIdRowToFeed(feed))
}
