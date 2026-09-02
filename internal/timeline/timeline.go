package timeline

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
)

type Mode string

const (
	ModeGlobal         Mode = "global"
	ModePersonalized   Mode = "personalized"
	ModeGlobalFallback Mode = "global_fallback"
)

type Filters struct {
	CategoryIDs string
	Search      string
	Limit       int32
	Offset      int32
}

type Category struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

type Feed struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	URL  string    `json:"url"`
}

type Post struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	PublishedAt time.Time  `json:"published_at"`
	URL         string     `json:"url"`
	ImageURL    *string    `json:"image_url"`
	Feed        Feed       `json:"feed"`
	Categories  []Category `json:"categories"`
}

type Result struct {
	Posts   []Post `json:"posts"`
	Mode    Mode   `json:"mode"`
	Limit   int32  `json:"limit"`
	Offset  int32  `json:"offset"`
	HasMore bool   `json:"has_more"`
}

type row struct {
	ID          uuid.UUID
	Title       string
	Description sql.NullString
	PublishedAt time.Time
	URL         string
	FeedID      uuid.UUID
	ImageURL    sql.NullString
	FeedName    string
	FeedURL     string
}

func List(ctx context.Context, db *database.Queries, user *database.User, filters Filters) (Result, error) {
	mode := ModeGlobal
	queryLimit := filters.Limit + 1
	var rows []row

	if user != nil {
		hasFollows, err := db.UserHasFeedFollows(ctx, user.ID)
		if err != nil {
			return Result{}, err
		}
		if hasFollows {
			mode = ModePersonalized
			items, err := db.GetPostsForUser(ctx, database.GetPostsForUserParams{
				UserID: user.ID, CategoryIds: filters.CategoryIDs, Search: filters.Search,
				ResultLimit: queryLimit, ResultOffset: filters.Offset,
			})
			if err != nil {
				return Result{}, err
			}
			rows = make([]row, 0, len(items))
			for _, item := range items {
				rows = append(rows, row{
					ID: item.ID, Title: item.Title, Description: item.Description,
					PublishedAt: item.PublishedAt, URL: item.Url, FeedID: item.FeedID,
					ImageURL: item.ImageUrl, FeedName: item.FeedName, FeedURL: item.FeedUrl,
				})
			}
		} else {
			mode = ModeGlobalFallback
		}
	}

	if mode != ModePersonalized {
		items, err := db.GetGlobalPosts(ctx, database.GetGlobalPostsParams{
			CategoryIds: filters.CategoryIDs, Search: filters.Search,
			ResultLimit: queryLimit, ResultOffset: filters.Offset,
		})
		if err != nil {
			return Result{}, err
		}
		rows = make([]row, 0, len(items))
		for _, item := range items {
			rows = append(rows, row{
				ID: item.ID, Title: item.Title, Description: item.Description,
				PublishedAt: item.PublishedAt, URL: item.Url, FeedID: item.FeedID,
				ImageURL: item.ImageUrl, FeedName: item.FeedName, FeedURL: item.FeedUrl,
			})
		}
	}

	hasMore := len(rows) > int(filters.Limit)
	if hasMore {
		rows = rows[:filters.Limit]
	}
	posts, err := hydrateCategories(ctx, db, rows)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Posts: posts, Mode: mode, Limit: filters.Limit,
		Offset: filters.Offset, HasMore: hasMore,
	}, nil
}

func hydrateCategories(ctx context.Context, db *database.Queries, rows []row) ([]Post, error) {
	posts := make([]Post, 0, len(rows))
	if len(rows) == 0 {
		return posts, nil
	}

	feedIDs := make([]string, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, item := range rows {
		if _, ok := seen[item.FeedID]; !ok {
			seen[item.FeedID] = struct{}{}
			feedIDs = append(feedIDs, item.FeedID.String())
		}
	}
	categoryRows, err := db.GetCategoriesForFeeds(ctx, strings.Join(feedIDs, ","))
	if err != nil {
		return nil, err
	}
	categories := make(map[uuid.UUID][]Category, len(feedIDs))
	for _, item := range categoryRows {
		categories[item.FeedID] = append(categories[item.FeedID], Category{
			ID: item.ID, Name: item.Name, Slug: item.Slug,
		})
	}

	for _, item := range rows {
		var description *string
		if item.Description.Valid {
			value := item.Description.String
			description = &value
		}
		var imageURL *string
		if item.ImageURL.Valid {
			value := item.ImageURL.String
			imageURL = &value
		}
		postCategories := categories[item.FeedID]
		if postCategories == nil {
			postCategories = make([]Category, 0)
		}
		posts = append(posts, Post{
			ID: item.ID, Title: item.Title, Description: description,
			PublishedAt: item.PublishedAt, URL: item.URL, ImageURL: imageURL,
			Feed:       Feed{ID: item.FeedID, Name: item.FeedName, URL: item.FeedURL},
			Categories: postCategories,
		})
	}
	return posts, nil
}
