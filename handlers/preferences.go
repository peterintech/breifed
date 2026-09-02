package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/peterintech/briefed/components"
	"github.com/peterintech/briefed/internal/preferences"
	webtypes "github.com/peterintech/briefed/types"
)

func (h *Handler) RegisterPreferenceRoutes(router *chi.Mux) {
	router.Get("/partials/preferences", h.preferencesDrawer)
	router.Get("/partials/preferences/feeds", h.preferenceFeeds)
	router.Put("/partials/preferences", h.updatePreferences)
}

func (h *Handler) preferenceData(r *http.Request, categoryIDs, feedIDs []uuid.UUID, message string) (webtypes.PreferencesData, error) {
	categories, err := h.DB.GetCategories(r.Context())
	if err != nil {
		return webtypes.PreferencesData{}, err
	}
	selectedCategories := make(map[uuid.UUID]bool, len(categoryIDs))
	for _, id := range categoryIDs {
		selectedCategories[id] = true
	}
	selectedFeeds := make(map[uuid.UUID]bool, len(feedIDs))
	for _, id := range feedIDs {
		selectedFeeds[id] = true
	}
	feeds, err := h.feedOptions(r, categoryIDs, selectedFeeds)
	if err != nil {
		return webtypes.PreferencesData{}, err
	}
	return webtypes.PreferencesData{Categories: categories, Feeds: feeds, SelectedCategoryIDs: selectedCategories, Error: message}, nil
}

func (h *Handler) preferencesDrawer(w http.ResponseWriter, r *http.Request) {
	user, _, err := h.viewer(r)
	if err != nil {
		http.Error(w, "could not load preferences", http.StatusInternalServerError)
		return
	}
	if user == nil {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	categories, err := h.DB.GetCategoriesForUser(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "could not load preferences", http.StatusInternalServerError)
		return
	}
	feeds, err := h.DB.GetFeedsForUser(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "could not load preferences", http.StatusInternalServerError)
		return
	}
	categoryIDs := make([]uuid.UUID, 0, len(categories))
	for _, category := range categories {
		categoryIDs = append(categoryIDs, category.ID)
	}
	feedIDs := make([]uuid.UUID, 0, len(feeds))
	for _, feed := range feeds {
		feedIDs = append(feedIDs, feed.ID)
	}
	data, err := h.preferenceData(r, categoryIDs, feedIDs, "")
	if err != nil {
		http.Error(w, "could not load preferences", http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, components.PreferencesDrawer(data))
}

func (h *Handler) preferenceFeeds(w http.ResponseWriter, r *http.Request) {
	user, _, err := h.viewer(r)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	categoryIDs, err := parseUUIDs(r.Form["category_ids"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	feedIDs, _ := parseUUIDs(r.Form["feed_ids"])
	selected := make(map[uuid.UUID]bool)
	for _, id := range feedIDs {
		selected[id] = true
	}
	feeds, err := h.feedOptions(r, categoryIDs, selected)
	if err != nil {
		http.Error(w, "could not load sources", http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, components.PreferenceFeeds(feeds))
}

func (h *Handler) updatePreferences(w http.ResponseWriter, r *http.Request) {
	user, _, err := h.viewer(r)
	if err != nil {
		render(w, r, http.StatusOK, components.InlineError("We could not save your changes."))
		return
	}
	if user == nil {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		render(w, r, http.StatusOK, components.InlineError("Check your selections and try again."))
		return
	}
	categoryIDs, categoryErr := parseUUIDs(r.Form["category_ids"])
	feedIDs, feedErr := parseUUIDs(r.Form["feed_ids"])
	if categoryErr != nil || feedErr != nil || len(categoryIDs) == 0 {
		render(w, r, http.StatusOK, components.InlineError("Choose at least one interest."))
		return
	}
	if err := preferences.Replace(r.Context(), h.Conn, h.DB, user.ID, categoryIDs, feedIDs); err != nil {
		render(w, r, http.StatusOK, components.InlineError("We could not save those preferences. Check your selections and try again."))
		return
	}
	w.Header().Set("HX-Trigger", "preferencesSaved")
	w.WriteHeader(http.StatusNoContent)
}
