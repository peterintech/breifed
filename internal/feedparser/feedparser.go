package feedparser

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Feed struct {
	Title string
	Items []Item
}

type Item struct {
	Title       string
	URL         string
	Description string
	PublishedAt time.Time
}

type rssDocument struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Content     string `xml:"encoded"`
	PublishedAt string `xml:"pubDate"`
}

type atomDocument struct {
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

func Fetch(ctx context.Context, feedURL string) (*Feed, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed responded with status %d", response.StatusCode)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

func Parse(data []byte) (*Feed, error) {
	var root struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	switch strings.ToLower(root.XMLName.Local) {
	case "rss":
		return parseRSS(data)
	case "feed":
		return parseAtom(data)
	default:
		return nil, fmt.Errorf("unsupported feed format")
	}
}

func parseRSS(data []byte) (*Feed, error) {
	var document rssDocument
	if err := xml.Unmarshal(data, &document); err != nil {
		return nil, err
	}

	feed := &Feed{Title: strings.TrimSpace(document.Channel.Title), Items: make([]Item, 0, len(document.Channel.Items))}
	for _, source := range document.Channel.Items {
		description := source.Description
		if strings.TrimSpace(description) == "" {
			description = source.Content
		}
		feed.Items = append(feed.Items, Item{
			Title: strings.TrimSpace(source.Title), URL: strings.TrimSpace(source.Link),
			Description: strings.TrimSpace(description), PublishedAt: parseDate(source.PublishedAt),
		})
	}
	return feed, nil
}

func parseAtom(data []byte) (*Feed, error) {
	var document atomDocument
	if err := xml.Unmarshal(data, &document); err != nil {
		return nil, err
	}

	feed := &Feed{Title: strings.TrimSpace(document.Title), Items: make([]Item, 0, len(document.Entries))}
	for _, source := range document.Entries {
		description := source.Summary
		if strings.TrimSpace(description) == "" {
			description = source.Content
		}
		publishedAt := source.Published
		if publishedAt == "" {
			publishedAt = source.Updated
		}
		feed.Items = append(feed.Items, Item{
			Title: strings.TrimSpace(source.Title), URL: atomEntryURL(source.Links),
			Description: strings.TrimSpace(description), PublishedAt: parseDate(publishedAt),
		})
	}
	return feed, nil
}

func atomEntryURL(links []atomLink) string {
	for _, link := range links {
		if strings.TrimSpace(link.Href) != "" && (link.Rel == "" || link.Rel == "alternate") {
			return strings.TrimSpace(link.Href)
		}
	}
	for _, link := range links {
		if strings.TrimSpace(link.Href) != "" {
			return strings.TrimSpace(link.Href)
		}
	}
	return ""
}

func parseDate(value string) time.Time {
	value = strings.TrimSpace(value)
	layouts := []string{
		time.RFC3339, time.RFC3339Nano, time.RFC1123Z, time.RFC1123,
		time.RFC822Z, time.RFC822, time.RFC850, time.ANSIC,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}
