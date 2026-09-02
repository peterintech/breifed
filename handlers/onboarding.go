package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/components"
	webtypes "github.com/peterintech/briefed/types"
)

func (h *Handler) onboardingInterests(w http.ResponseWriter, r *http.Request) {
	categories, err := h.DB.GetCategories(r.Context())
	if err != nil {
		http.Error(w, "could not load interests", http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, components.InterestStep(webtypes.InterestStepData{Categories: categories, Selected: map[uuid.UUID]bool{}}))
}

func (h *Handler) onboardingFeeds(w http.ResponseWriter, r *http.Request) {
	categoryIDs, err := parseUUIDs(r.URL.Query()["category_ids"])
	if err != nil || len(categoryIDs) == 0 {
		categories, dbErr := h.DB.GetCategories(r.Context())
		if dbErr != nil {
			http.Error(w, "could not load interests", http.StatusInternalServerError)
			return
		}
		render(w, r, http.StatusOK, components.InterestStep(webtypes.InterestStepData{Categories: categories, Selected: map[uuid.UUID]bool{}, Error: "Choose at least one interest to continue."}))
		return
	}
	feeds, err := h.feedOptions(r, categoryIDs, map[uuid.UUID]bool{})
	if err != nil {
		http.Error(w, "could not load sources", http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, components.FeedStep(webtypes.FeedStepData{CategoryIDs: categoryIDs, Feeds: feeds}))
}

func (h *Handler) onboardingAccount(w http.ResponseWriter, r *http.Request) {
	categoryIDs, categoryErr := parseUUIDs(r.URL.Query()["category_ids"])
	feedIDs, feedErr := parseUUIDs(r.URL.Query()["feed_ids"])
	if categoryErr != nil || feedErr != nil || len(categoryIDs) == 0 || len(feedIDs) == 0 {
		selected := make(map[uuid.UUID]bool)
		for _, id := range feedIDs {
			selected[id] = true
		}
		feeds, err := h.feedOptions(r, categoryIDs, selected)
		if err != nil {
			http.Error(w, "could not load sources", http.StatusInternalServerError)
			return
		}
		render(w, r, http.StatusOK, components.FeedStep(webtypes.FeedStepData{CategoryIDs: categoryIDs, Feeds: feeds, Error: "Choose at least one source to continue."}))
		return
	}
	render(w, r, http.StatusOK, components.AccountStep(webtypes.AccountStepData{CategoryIDs: categoryIDs, FeedIDs: feedIDs}))
}
