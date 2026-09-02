package feedcatalog

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/feedparser"
)

var (
	ErrInvalidInput      = errors.New("a feed URL and at least one category are required")
	ErrInvalidFeed       = errors.New("the URL is not a valid RSS or Atom feed")
	ErrInvalidCategories = errors.New("one or more categories are invalid")
)

type Result struct {
	Feed    database.Feed
	Created bool
}

func Contribute(ctx context.Context, conn *sql.DB, db *database.Queries, userID uuid.UUID, feedURL string, categoryIDs []uuid.UUID) (Result, error) {
	feedURL = strings.TrimSpace(feedURL)
	if feedURL == "" || len(categoryIDs) == 0 {
		return Result{}, ErrInvalidInput
	}

	existing, err := db.GetFeedByURL(ctx, feedURL)
	if err == nil {
		return followExisting(ctx, db, userID, existing)
	}
	if err != sql.ErrNoRows {
		return Result{}, err
	}

	parsedFeed, err := feedparser.Fetch(ctx, feedURL)
	if err != nil || strings.TrimSpace(parsedFeed.Title) == "" {
		return Result{}, ErrInvalidFeed
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()
	queries := db.WithTx(tx)
	now := time.Now().UTC()
	feed, err := queries.CreateFeed(ctx, database.CreateFeedParams{
		ID: uuid.New(), CreatedAt: now, UpdatedAt: now,
		Name: strings.TrimSpace(parsedFeed.Title), Url: feedURL,
		SubmittedBy: uuid.NullUUID{UUID: userID, Valid: true},
	})
	if err != nil {
		if existing, getErr := db.GetFeedByURL(ctx, feedURL); getErr == nil {
			return followExisting(ctx, db, userID, existing)
		}
		return Result{}, err
	}

	for _, categoryID := range categoryIDs {
		if err := queries.CreateFeedCategory(ctx, database.CreateFeedCategoryParams{FeedID: feed.ID, CategoryID: categoryID}); err != nil {
			return Result{}, ErrInvalidCategories
		}
	}
	if err := queries.CreateFeedFollow(ctx, database.CreateFeedFollowParams{UserID: userID, FeedID: feed.ID}); err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return Result{Feed: feed, Created: true}, nil
}

func followExisting(ctx context.Context, db *database.Queries, userID uuid.UUID, feed database.Feed) (Result, error) {
	if err := db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{UserID: userID, FeedID: feed.ID}); err != nil {
		return Result{}, err
	}
	return Result{Feed: feed, Created: false}, nil
}
