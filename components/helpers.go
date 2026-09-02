package components

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/timeline"
	webtypes "github.com/peterintech/briefed/types"
)

func stepClass(active bool) string {
	if active {
		return "h-1 rounded-full bg-forest-800"
	}
	return "h-1 rounded-full bg-rule"
}
func backToFeedsURL(categoryIDs []uuid.UUID) string {
	values := url.Values{}
	for _, id := range categoryIDs {
		values.Add("category_ids", id.String())
	}
	return "/partials/onboarding/feeds?" + values.Encode()
}
func selectedFeedCount(feeds []webtypes.FeedOption) int {
	count := 0
	for _, feed := range feeds {
		if feed.Selected {
			count++
		}
	}
	return count
}

func categoryClass(selected bool) string {
	base := "inline-flex shrink-0 items-center border-b-2 px-1 py-3 text-sm font-medium transition-colors duration-200 "
	if selected {
		return base + "border-forest-800 text-forest-900"
	}
	return base + "border-transparent text-muted hover:border-rule hover:text-ink"
}

func choiceClass(selected bool) string {
	base := "group flex cursor-pointer items-center justify-between gap-3 rounded-control border px-4 py-3 text-sm font-medium transition-colors duration-200 "
	if selected {
		return base + "border-forest-800 bg-forest-50 text-forest-900"
	}
	return base + "border-rule bg-paper-raised text-ink hover:border-forest-700"
}

func partialPostsURL(categoryID, search string, offset int32) string {
	values := url.Values{}
	if categoryID != "" {
		values.Set("category_ids", categoryID)
	}
	if search != "" {
		values.Set("q", search)
	}
	if offset > 0 {
		values.Set("offset", fmt.Sprintf("%d", offset))
	}
	if query := values.Encode(); query != "" {
		return "/partials/posts?" + query
	}
	return "/partials/posts"
}

func homeURL(categoryID, search string) string {
	values := url.Values{}
	if categoryID != "" {
		values.Set("category_ids", categoryID)
	}
	if search != "" {
		values.Set("q", search)
	}
	if query := values.Encode(); query != "" {
		return "/?" + query
	}
	return "/"
}

func primaryCategory(post timeline.Post) *timeline.Category {
	if len(post.Categories) == 0 {
		return nil
	}
	return &post.Categories[0]
}

func excerpt(value *string, max int) string {
	if value == nil {
		return ""
	}
	text := strings.TrimSpace(*value)
	if utf8.RuneCountInString(text) <= max {
		return text
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:max])) + "…"
}

func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "B"
	}
	if len(parts) == 1 {
		runes := []rune(parts[0])
		if len(runes) > 1 {
			return strings.ToUpper(string(runes[:2]))
		}
		return strings.ToUpper(string(runes))
	}
	return strings.ToUpper(string([]rune(parts[0])[:1]) + string([]rune(parts[1])[:1]))
}

func relativeTime(published time.Time) string {
	elapsed := time.Since(published)
	if elapsed < time.Minute {
		return "just now"
	}
	if elapsed < time.Hour {
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	}
	if elapsed < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	}
	if elapsed < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(elapsed.Hours()/24))
	}
	return published.Format("2 Jan 2006")
}

func selected(set map[uuid.UUID]bool, id uuid.UUID) bool {
	return set[id]
}

func sourceOptions(posts []timeline.Post, followed map[uuid.UUID]bool) []webtypes.SourceOption {
	result := make([]webtypes.SourceOption, 0, 6)
	seen := make(map[uuid.UUID]struct{})
	for _, post := range posts {
		if _, ok := seen[post.Feed.ID]; ok {
			continue
		}
		seen[post.Feed.ID] = struct{}{}
		result = append(result, webtypes.SourceOption{
			ID: post.Feed.ID, Name: post.Feed.Name, Followed: followed[post.Feed.ID],
		})
		if len(result) == 6 {
			break
		}
	}
	return result
}

func modeLabel(mode timeline.Mode) string {
	if mode == timeline.ModePersonalized {
		return "For You"
	}
	return "Latest"
}
