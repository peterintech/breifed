package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}

type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Feed struct {
	ID            uuid.UUID  `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	SubmittedBy   *uuid.UUID `json:"submitted_by"`
	LastFetchedAt *time.Time `json:"last_fetched_at"`
	Categories    []Category `json:"categories"`
}

type Profile struct {
	User       User       `json:"user"`
	Categories []Category `json:"categories"`
	Feeds      []Feed     `json:"feeds"`
}

func databaseUserToUser(dbUser database.User) User {
	return User{ID: dbUser.ID, CreatedAt: dbUser.CreatedAt, UpdatedAt: dbUser.UpdatedAt, Name: dbUser.Name, Email: dbUser.Email}
}

func databaseCategoryToCategory(dbCategory database.Category) Category {
	return Category{ID: dbCategory.ID, Name: dbCategory.Name, Slug: dbCategory.Slug, CreatedAt: dbCategory.CreatedAt}
}

func databaseCategoriesToCategories(dbCategories []database.Category) []Category {
	categories := make([]Category, 0, len(dbCategories))
	for _, category := range dbCategories {
		categories = append(categories, databaseCategoryToCategory(category))
	}
	return categories
}

func databaseFeedToFeed(dbFeed database.Feed, dbCategories []database.Category) Feed {
	var submittedBy *uuid.UUID
	if dbFeed.SubmittedBy.Valid {
		id := dbFeed.SubmittedBy.UUID
		submittedBy = &id
	}
	var lastFetchedAt *time.Time
	if dbFeed.LastFetchedAt.Valid {
		value := dbFeed.LastFetchedAt.Time
		lastFetchedAt = &value
	}
	return Feed{
		ID: dbFeed.ID, CreatedAt: dbFeed.CreatedAt, UpdatedAt: dbFeed.UpdatedAt,
		Name: dbFeed.Name, URL: dbFeed.Url, SubmittedBy: submittedBy,
		LastFetchedAt: lastFetchedAt, Categories: databaseCategoriesToCategories(dbCategories),
	}
}
