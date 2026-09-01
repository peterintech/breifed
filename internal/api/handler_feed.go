package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/feedparser"
)

func (ac *apiConfig) createFeedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	var params struct {
		URL         string      `json:"url"`
		CategoryIDs []uuid.UUID `json:"category_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.URL = strings.TrimSpace(params.URL)
	if params.URL == "" || len(params.CategoryIDs) == 0 {
		errorResponse(w, http.StatusBadRequest, "url and at least one category_id are required")
		return
	}

	existing, err := ac.DB.GetFeedByURL(r.Context(), params.URL)
	if err == nil {
		ac.followAndRespond(w, r, user, existing, http.StatusOK)
		return
	}
	if err != sql.ErrNoRows {
		errorResponse(w, http.StatusInternalServerError, "could not check feed")
		return
	}

	parsedFeed, err := feedparser.Fetch(r.Context(), params.URL)
	if err != nil || strings.TrimSpace(parsedFeed.Title) == "" {
		errorResponse(w, http.StatusBadRequest, "url is not a valid RSS or Atom feed")
		return
	}

	tx, err := ac.Conn.BeginTx(r.Context(), nil)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not create feed")
		return
	}
	defer tx.Rollback()
	queries := ac.DB.WithTx(tx)
	now := time.Now().UTC()

	feed, err := queries.CreateFeed(r.Context(), database.CreateFeedParams{
		ID: uuid.New(), CreatedAt: now, UpdatedAt: now,
		Name: strings.TrimSpace(parsedFeed.Title), Url: params.URL,
		SubmittedBy: uuid.NullUUID{UUID: user.ID, Valid: true},
	})
	if err != nil {
		_ = tx.Rollback()
		if existing, getErr := ac.DB.GetFeedByURL(r.Context(), params.URL); getErr == nil {
			ac.followAndRespond(w, r, user, existing, http.StatusOK)
			return
		}
		errorResponse(w, http.StatusInternalServerError, "could not create feed")
		return
	}

	for _, categoryID := range params.CategoryIDs {
		if err := queries.CreateFeedCategory(r.Context(), database.CreateFeedCategoryParams{
			FeedID: feed.ID, CategoryID: categoryID,
		}); err != nil {
			errorResponse(w, http.StatusBadRequest, "one or more category IDs are invalid")
			return
		}
	}
	if err := queries.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		UserID: user.ID, FeedID: feed.ID,
	}); err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not follow feed")
		return
	}
	if err := tx.Commit(); err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not create feed")
		return
	}

	response, err := ac.feedResponse(r.Context(), feed)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch feed categories")
		return
	}
	jsonResponse(w, http.StatusCreated, response)
}

func (ac *apiConfig) followAndRespond(w http.ResponseWriter, r *http.Request, user database.User, feed database.Feed, status int) {
	if err := ac.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		UserID: user.ID, FeedID: feed.ID,
	}); err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not follow feed")
		return
	}
	response, err := ac.feedResponse(r.Context(), feed)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch feed")
		return
	}
	jsonResponse(w, status, response)
}

func (ac *apiConfig) getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	categoryIDs, err := normalizeCategoryFilter(r.URL.Query().Get("category_ids"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	feeds, err := ac.DB.GetFeeds(r.Context(), database.GetFeedsParams{
		CategoryIds: categoryIDs, Search: strings.TrimSpace(r.URL.Query().Get("q")),
		ResultLimit: limit, ResultOffset: offset,
	})
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch feeds")
		return
	}
	response, err := ac.feedResponses(r.Context(), feeds)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch feed categories")
		return
	}
	jsonResponse(w, http.StatusOK, response)
}

func (ac *apiConfig) getFeedByIDHandler(w http.ResponseWriter, r *http.Request) {
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
	response, err := ac.feedResponse(r.Context(), feed)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch feed categories")
		return
	}
	jsonResponse(w, http.StatusOK, response)
}
