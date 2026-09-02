package api

import (
	"net/http"
	"strings"

	"github.com/peterintech/briefed/internal/sessionauth"
	"github.com/peterintech/briefed/internal/timeline"
)

func (ac *apiConfig) getPostsHandler(w http.ResponseWriter, r *http.Request) {
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
	user, err := sessionauth.CurrentUser(r.Context(), ac.DB, r)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not verify session")
		return
	}
	result, err := timeline.List(r.Context(), ac.DB, user, timeline.Filters{
		CategoryIDs: categoryIDs, Search: strings.TrimSpace(r.URL.Query().Get("q")),
		Limit: limit, Offset: offset,
	})
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch posts")
		return
	}
	jsonResponse(w, http.StatusOK, result)
}
