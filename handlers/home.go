package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/components"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/timeline"
	webtypes "github.com/peterintech/briefed/types"
	"github.com/peterintech/briefed/views"
)

func (h *Handler) timelineData(r *http.Request) (*webtypes.Viewer, webtypes.TimelineData, error) {
	user, viewer, err := h.viewer(r)
	if err != nil {
		return nil, webtypes.TimelineData{}, err
	}
	filters := parseFilters(r)
	result, err := timeline.List(r.Context(), h.DB, user, filters)
	if err != nil {
		return nil, webtypes.TimelineData{}, err
	}
	categories, err := h.DB.GetCategories(r.Context())
	if err != nil {
		return nil, webtypes.TimelineData{}, err
	}
	followed := make(map[uuid.UUID]bool)
	if user != nil {
		feeds, err := h.DB.GetFeedsForUser(r.Context(), user.ID)
		if err != nil {
			return nil, webtypes.TimelineData{}, err
		}
		for _, feed := range feeds {
			followed[feed.ID] = true
		}
	}
	return viewer, webtypes.TimelineData{Result: result, Categories: categories, Sources: sourceOptions(result.Posts, followed), SelectedCategory: filters.CategoryIDs, Search: filters.Search}, nil
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	viewer, data, err := h.timelineData(r)
	if err != nil {
		http.Error(w, "could not load Briefed", http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, views.Home(webtypes.HomeData{Viewer: viewer, Categories: data.Categories, Timeline: data}))
}

func (h *Handler) posts(w http.ResponseWriter, r *http.Request) {
	_, data, err := h.timelineData(r)
	if err != nil {
		http.Error(w, "could not load stories", http.StatusInternalServerError)
		return
	}
	if parseFilters(r).Offset > 0 {
		render(w, r, http.StatusOK, components.PostRows(data))
		return
	}
	query := r.URL.Query()
	query.Del("offset")
	path := "/"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	w.Header().Set("HX-Push-Url", path)
	render(w, r, http.StatusOK, components.Timeline(data))
}

func (h *Handler) feedOptions(r *http.Request, categoryIDs []uuid.UUID, selected map[uuid.UUID]bool) ([]webtypes.FeedOption, error) {
	parts := make([]string, 0, len(categoryIDs))
	for _, id := range categoryIDs {
		parts = append(parts, id.String())
	}
	feeds, err := h.DB.GetFeeds(r.Context(), database.GetFeedsParams{CategoryIds: strings.Join(parts, ","), Search: "", ResultLimit: 200, ResultOffset: 0})
	if err != nil {
		return nil, err
	}
	result := make([]webtypes.FeedOption, 0, len(feeds))
	for _, feed := range feeds {
		categories, err := h.DB.GetCategoriesForFeed(r.Context(), feed.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, webtypes.FeedOption{ID: feed.ID, Name: feed.Name, URL: feed.Url, Categories: categories, Selected: selected[feed.ID]})
	}
	return result, nil
}
