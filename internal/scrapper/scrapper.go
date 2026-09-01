package scrapper

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/feedparser"
)

func Start(db *database.Queries, concurrency int, timeBetweenRequest time.Duration) {
	log.Printf("Collecting feeds every %s on %v goroutines...", timeBetweenRequest, concurrency)
	ticker := time.NewTicker(timeBetweenRequest)
	defer ticker.Stop()

	for {
		scrapeBatch(db, concurrency)
		<-ticker.C
	}
}

func scrapeBatch(db *database.Queries, concurrency int) {
	feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
	if err != nil {
		log.Printf("Couldn't get next feeds to fetch: %v", err)
		return
	}

	var waitGroup sync.WaitGroup
	for _, feed := range feeds {
		waitGroup.Add(1)
		go func(feed database.Feed) {
			defer waitGroup.Done()
			scrapeFeed(db, feed)
		}(feed)
	}
	waitGroup.Wait()
}

func scrapeFeed(db *database.Queries, feed database.Feed) {
	parsedFeed, fetchErr := feedparser.Fetch(context.Background(), feed.Url)
	if _, err := db.MarkFeedAsFetched(context.Background(), feed.ID); err != nil {
		log.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
	}
	if fetchErr != nil {
		log.Printf("Couldn't collect feed %s: %v", feed.Name, fetchErr)
		return
	}

	saved := 0
	for _, item := range parsedFeed.Items {
		if item.Title == "" || item.URL == "" || item.PublishedAt.IsZero() {
			continue
		}

		description := sql.NullString{}
		if strings.TrimSpace(item.Description) != "" {
			description = sql.NullString{String: item.Description, Valid: true}
		}
		now := time.Now().UTC()
		if err := db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Title: item.Title,
			Description: description, PublishedAt: item.PublishedAt, Url: item.URL, FeedID: feed.ID,
		}); err != nil {
			log.Printf("Couldn't create post %s: %v", item.Title, err)
			continue
		}
		saved++
	}
	log.Printf("Feed %s collected, %d posts processed", feed.Name, saved)
}
