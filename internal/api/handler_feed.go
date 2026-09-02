package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/feedcatalog"
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
	result, err := feedcatalog.Contribute(r.Context(), ac.Conn, ac.DB, user.ID, params.URL, params.CategoryIDs)
	if errors.Is(err, feedcatalog.ErrInvalidInput) || errors.Is(err, feedcatalog.ErrInvalidFeed) || errors.Is(err, feedcatalog.ErrInvalidCategories) {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not create feed")
		return
	}
	response, err := ac.feedResponse(r.Context(), result.Feed)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch feed categories")
		return
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
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
