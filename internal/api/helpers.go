package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
)

func newSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: token, Path: "/", Expires: expiresAt,
		MaxAge: int(time.Until(expiresAt).Seconds()), HttpOnly: true,
		Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode,
	})
}

func parsePagination(r *http.Request) (int32, int32, error) {
	limit, offset := int64(20), int64(0)
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.ParseInt(raw, 10, 32)
		if err != nil || limit < 1 || limit > 100 {
			return 0, 0, fmt.Errorf("limit must be between 1 and 100")
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.ParseInt(raw, 10, 32)
		if err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("offset must be zero or greater")
		}
	}
	return int32(limit), int32(offset), nil
}

func normalizeCategoryFilter(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	parts := strings.Split(raw, ",")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		id, err := uuid.Parse(strings.TrimSpace(part))
		if err != nil {
			return "", fmt.Errorf("category_ids must contain valid UUIDs")
		}
		normalized = append(normalized, id.String())
	}
	return strings.Join(normalized, ","), nil
}

func (ac *apiConfig) feedResponse(ctx context.Context, feed database.Feed) (Feed, error) {
	categories, err := ac.DB.GetCategoriesForFeed(ctx, feed.ID)
	if err != nil {
		return Feed{}, err
	}
	return databaseFeedToFeed(feed, categories), nil
}

func (ac *apiConfig) feedResponses(ctx context.Context, feeds []database.Feed) ([]Feed, error) {
	response := make([]Feed, 0, len(feeds))
	for _, feed := range feeds {
		item, err := ac.feedResponse(ctx, feed)
		if err != nil {
			return nil, err
		}
		response = append(response, item)
	}
	return response, nil
}
