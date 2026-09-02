package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/peterintech/briefed/components"
	"github.com/peterintech/briefed/internal/feedcatalog"
	webtypes "github.com/peterintech/briefed/types"
)

func (h *Handler) RegisterFeedContributionRoutes(router *chi.Mux) {
	router.Get("/partials/feeds/new", h.feedContributionDrawer)
	router.Post("/partials/feeds", h.createFeedContribution)
}

func (h *Handler) contributionData(r *http.Request, feedURL, message string, selectedIDs []uuid.UUID) (webtypes.FeedContributionData, error) {
	categories, err := h.DB.GetCategories(r.Context())
	if err != nil {
		return webtypes.FeedContributionData{}, err
	}
	selected := make(map[uuid.UUID]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		selected[id] = true
	}
	return webtypes.FeedContributionData{URL: feedURL, Categories: categories, Selected: selected, Error: message}, nil
}

func (h *Handler) feedContributionDrawer(w http.ResponseWriter, r *http.Request) {
	user, _, err := h.viewer(r)
	if err != nil {
		http.Error(w, "could not load feed form", http.StatusInternalServerError)
		return
	}
	if user == nil {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	data, err := h.contributionData(r, "", "", nil)
	if err != nil {
		http.Error(w, "could not load feed form", http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, components.FeedContributionDrawer(data))
}

func (h *Handler) createFeedContribution(w http.ResponseWriter, r *http.Request) {
	user, _, err := h.viewer(r)
	if err != nil {
		http.Error(w, "could not add feed", http.StatusInternalServerError)
		return
	}
	if user == nil {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.renderContributionError(w, r, "", nil, "Check the form and try again.")
		return
	}
	categoryIDs, parseErr := parseUUIDs(r.Form["category_ids"])
	feedURL := r.FormValue("url")
	if parseErr != nil || len(categoryIDs) == 0 {
		h.renderContributionError(w, r, feedURL, categoryIDs, "Choose at least one valid category.")
		return
	}
	result, err := feedcatalog.Contribute(r.Context(), h.Conn, h.DB, user.ID, feedURL, categoryIDs)
	switch {
	case errors.Is(err, feedcatalog.ErrInvalidInput):
		h.renderContributionError(w, r, feedURL, categoryIDs, "Add a direct feed URL and choose at least one category.")
		return
	case errors.Is(err, feedcatalog.ErrInvalidFeed):
		h.renderContributionError(w, r, feedURL, categoryIDs, "We could not find a valid RSS or Atom feed at that URL.")
		return
	case errors.Is(err, feedcatalog.ErrInvalidCategories):
		h.renderContributionError(w, r, feedURL, categoryIDs, "One of those categories is no longer available. Refresh and try again.")
		return
	case err != nil:
		h.renderContributionError(w, r, feedURL, categoryIDs, "We could not add that feed right now. Try again.")
		return
	}
	w.Header().Set("HX-Trigger", "feedCreated")
	render(w, r, http.StatusOK, components.FeedContributionSuccess(webtypes.FeedContributionSuccess{Name: result.Feed.Name, URL: result.Feed.Url, Created: result.Created}))
}

func (h *Handler) renderContributionError(w http.ResponseWriter, r *http.Request, feedURL string, categoryIDs []uuid.UUID, message string) {
	data, err := h.contributionData(r, feedURL, message, categoryIDs)
	if err != nil {
		http.Error(w, "could not load feed form", http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, components.FeedContributionForm(data))
}
