package types

import (
	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/timeline"
)

type Viewer struct {
	ID    uuid.UUID
	Name  string
	Email string
}

type FeedOption struct {
	ID         uuid.UUID
	Name       string
	URL        string
	Categories []database.Category
	Selected   bool
}

type SourceOption struct {
	ID       uuid.UUID
	Name     string
	Followed bool
}

type TimelineData struct {
	Result           timeline.Result
	Categories       []database.Category
	Sources          []SourceOption
	SelectedCategory string
	Search           string
	CanContribute    bool
}

type HomeData struct {
	Viewer     *Viewer
	Categories []database.Category
	Timeline   TimelineData
}

type LoginData struct {
	Email      string
	Error      string
	Categories []database.Category
}

type InterestStepData struct {
	Categories []database.Category
	Selected   map[uuid.UUID]bool
	Error      string
}

type FeedStepData struct {
	CategoryIDs []uuid.UUID
	Feeds       []FeedOption
	Error       string
}

type AccountStepData struct {
	CategoryIDs []uuid.UUID
	FeedIDs     []uuid.UUID
	Name        string
	Email       string
	Error       string
}

type PreferencesData struct {
	Categories          []database.Category
	Feeds               []FeedOption
	SelectedCategoryIDs map[uuid.UUID]bool
	Error               string
}

type FeedContributionData struct {
	URL        string
	Categories []database.Category
	Selected   map[uuid.UUID]bool
	Error      string
}

type FeedContributionSuccess struct {
	Name    string
	URL     string
	Created bool
}
