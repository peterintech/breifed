package feedparser

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
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
	ImageURL    string
	PublishedAt time.Time
}

type rssDocument struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title           string           `xml:"title"`
	Link            string           `xml:"link"`
	Description     string           `xml:"description"`
	Content         string           `xml:"encoded"`
	PublishedAt     string           `xml:"pubDate"`
	MediaContents   []mediaContent   `xml:"http://search.yahoo.com/mrss/ content"`
	MediaThumbnails []mediaThumbnail `xml:"http://search.yahoo.com/mrss/ thumbnail"`
	Enclosures      []enclosure      `xml:"enclosure"`
}

type atomDocument struct {
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title           string           `xml:"title"`
	Links           []atomLink       `xml:"link"`
	Summary         string           `xml:"summary"`
	Content         string           `xml:"content"`
	Published       string           `xml:"published"`
	Updated         string           `xml:"updated"`
	MediaContents   []mediaContent   `xml:"http://search.yahoo.com/mrss/ content"`
	MediaThumbnails []mediaThumbnail `xml:"http://search.yahoo.com/mrss/ thumbnail"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type mediaContent struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Medium string `xml:"medium,attr"`
}

type mediaThumbnail struct {
	URL string `xml:"url,attr"`
}

type enclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

var (
	imagePattern       = regexp.MustCompile(`(?is)<img\b[^>]*\bsrc\s*=\s*(?:"([^"]+)"|'([^']+)'|([^\s>]+))`)
	scriptStylePattern = regexp.MustCompile(`(?is)<(script|style)\b[^>]*>.*?</(script|style)>`)
	tagPattern         = regexp.MustCompile(`(?s)<[^>]+>`)
	spacePattern       = regexp.MustCompile(`\s+`)
)

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
		articleURL := strings.TrimSpace(source.Link)
		feed.Items = append(feed.Items, Item{
			Title: strings.TrimSpace(source.Title), URL: articleURL,
			Description: plainText(description),
			ImageURL:    normalizeImageURL(rssImageURL(source, description), articleURL),
			PublishedAt: parseDate(source.PublishedAt),
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
		articleURL := atomEntryURL(source.Links)
		feed.Items = append(feed.Items, Item{
			Title: strings.TrimSpace(source.Title), URL: articleURL,
			Description: plainText(description),
			ImageURL:    normalizeImageURL(atomImageURL(source, description), articleURL),
			PublishedAt: parseDate(publishedAt),
		})
	}
	return feed, nil
}

func rssImageURL(item rssItem, description string) string {
	if value := mediaImageURL(item.MediaContents, item.MediaThumbnails); value != "" {
		return value
	}
	for _, candidate := range item.Enclosures {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(candidate.Type)), "image/") {
			return strings.TrimSpace(candidate.URL)
		}
	}
	if value := firstHTMLImage(item.Content); value != "" {
		return value
	}
	return firstHTMLImage(description)
}

func atomImageURL(entry atomEntry, description string) string {
	if value := mediaImageURL(entry.MediaContents, entry.MediaThumbnails); value != "" {
		return value
	}
	for _, link := range entry.Links {
		if strings.EqualFold(strings.TrimSpace(link.Rel), "enclosure") &&
			strings.HasPrefix(strings.ToLower(strings.TrimSpace(link.Type)), "image/") {
			return strings.TrimSpace(link.Href)
		}
	}
	if value := firstHTMLImage(entry.Content); value != "" {
		return value
	}
	return firstHTMLImage(description)
}

func mediaImageURL(contents []mediaContent, thumbnails []mediaThumbnail) string {
	for _, candidate := range contents {
		mediaType := strings.ToLower(strings.TrimSpace(candidate.Type))
		medium := strings.ToLower(strings.TrimSpace(candidate.Medium))
		if candidate.URL != "" && (strings.HasPrefix(mediaType, "image/") || medium == "image" || (mediaType == "" && medium == "")) {
			return strings.TrimSpace(candidate.URL)
		}
	}
	for _, candidate := range thumbnails {
		if strings.TrimSpace(candidate.URL) != "" {
			return strings.TrimSpace(candidate.URL)
		}
	}
	return ""
}

func firstHTMLImage(value string) string {
	matches := imagePattern.FindStringSubmatch(value)
	for _, match := range matches[1:] {
		if strings.TrimSpace(match) != "" {
			return html.UnescapeString(strings.TrimSpace(match))
		}
	}
	return ""
}

func plainText(value string) string {
	value = scriptStylePattern.ReplaceAllString(value, " ")
	value = tagPattern.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	return strings.TrimSpace(spacePattern.ReplaceAllString(value, " "))
}

func normalizeImageURL(imageURL, articleURL string) string {
	imageURL = strings.TrimSpace(html.UnescapeString(imageURL))
	if imageURL == "" {
		return ""
	}
	parsedImage, err := url.Parse(imageURL)
	if err != nil || parsedImage.IsAbs() {
		return imageURL
	}
	parsedArticle, err := url.Parse(articleURL)
	if err != nil || !parsedArticle.IsAbs() {
		return imageURL
	}
	return parsedArticle.ResolveReference(parsedImage).String()
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
		"Mon, 2 Jan 2006 15:04:05 -0700", "Mon, 2 Jan 2006 15:04:05 MST",
		time.RFC822Z, time.RFC822, time.RFC850, time.ANSIC,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}
