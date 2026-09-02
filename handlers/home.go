package handlers

import (
	"net/http"
	"net/url"
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
	categories, err := h.DB.GetCategories(r.Context())
	if err != nil {
		return nil, webtypes.TimelineData{}, err
	}
	selectedCategory := ""
	requestedCategory := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("category")))
	for _, category := range categories {
		if category.Slug == requestedCategory {
			filters.CategoryIDs = category.ID.String()
			selectedCategory = category.Slug
			break
		}
	}
	result, err := timeline.List(r.Context(), h.DB, user, filters)
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
	return viewer, webtypes.TimelineData{Result: result, Categories: categories, Sources: sourceOptions(result.Posts, followed), SelectedCategory: selectedCategory, Search: filters.Search}, nil
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	viewer, data, err := h.timelineData(r)
	if err != nil {
		http.Error(w, "could not load Briefed", http.StatusInternalServerError)
		return
	}
	canonicalURL := timelinePageURL(data.SelectedCategory, data.Search)
	if r.URL.RequestURI() != canonicalURL {
		http.Redirect(w, r, canonicalURL, http.StatusSeeOther)
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
	w.Header().Set("HX-Push-Url", timelinePageURL(data.SelectedCategory, data.Search))
	render(w, r, http.StatusOK, components.Timeline(data))
}

func timelinePageURL(categorySlug, search string) string {
	query := url.Values{}
	if categorySlug != "" {
		query.Set("category", categorySlug)
	}
	if search != "" {
		query.Set("q", search)
	}
	if encoded := query.Encode(); encoded != "" {
		return "/?" + encoded
	}
	return "/"
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
