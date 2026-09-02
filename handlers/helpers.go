package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/sessionauth"
	"github.com/peterintech/briefed/internal/timeline"
	webtypes "github.com/peterintech/briefed/types"
)

func (h *Handler) viewer(r *http.Request) (*database.User, *webtypes.Viewer, error) {
	user, err := sessionauth.CurrentUser(r.Context(), h.DB, r)
	if err != nil || user == nil {
		return user, nil, err
	}
	return user, &webtypes.Viewer{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}

func parseFilters(r *http.Request) timeline.Filters {
	limit := int32(20)
	if value, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && value > 0 && value <= 100 {
		limit = int32(value)
	}
	offset := int32(0)
	if value, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && value >= 0 {
		offset = int32(value)
	}
	return timeline.Filters{CategoryIDs: strings.TrimSpace(r.URL.Query().Get("category_ids")), Search: strings.TrimSpace(r.URL.Query().Get("q")), Limit: limit, Offset: offset}
}

func parseUUIDs(values []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := uuid.Parse(part)
			if err != nil {
				return nil, err
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result, nil
}

func sourceOptions(posts []timeline.Post, followed map[uuid.UUID]bool) []webtypes.SourceOption {
	result := make([]webtypes.SourceOption, 0, 6)
	seen := make(map[uuid.UUID]struct{})
	for _, post := range posts {
		if _, ok := seen[post.Feed.ID]; ok {
			continue
		}
		seen[post.Feed.ID] = struct{}{}
		result = append(result, webtypes.SourceOption{ID: post.Feed.ID, Name: post.Feed.Name, Followed: followed[post.Feed.ID]})
		if len(result) == 6 {
			break
		}
	}
	return result
}
