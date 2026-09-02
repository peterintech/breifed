package feedparser

import (
	"testing"
	"time"
)

func TestParseRSS(t *testing.T) {
	data := []byte(`<?xml version="1.0"?>
<rss version="2.0" xmlns:media="http://search.yahoo.com/mrss/"><channel><title>Example RSS</title>
<item><title>First</title><link>https://example.com/first</link>
<description><![CDATA[<p>Hello <strong>world</strong></p>]]></description>
<media:content url="https://cdn.example.com/first.jpg" type="image/jpeg"/>
<pubDate>Mon, 01 Sep 2025 12:00:00 +0000</pubDate></item>
</channel></rss>`)

	feed, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Title != "Example RSS" || len(feed.Items) != 1 {
		t.Fatalf("unexpected feed: %#v", feed)
	}
	if feed.Items[0].URL != "https://example.com/first" || feed.Items[0].PublishedAt.IsZero() {
		t.Fatalf("unexpected item: %#v", feed.Items[0])
	}
	if feed.Items[0].Description != "Hello world" || feed.Items[0].ImageURL != "https://cdn.example.com/first.jpg" {
		t.Fatalf("unexpected item content: %#v", feed.Items[0])
	}
}

func TestParseAtom(t *testing.T) {
	data := []byte(`<?xml version="1.0"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/"><title>Example Atom</title>
<entry><title>First</title><link rel="alternate" href="https://example.com/first"/>
<summary type="html">&lt;p&gt;Hello Atom&lt;/p&gt;</summary>
<media:thumbnail url="https://cdn.example.com/atom.jpg"/>
<updated>2025-09-01T12:00:00Z</updated></entry>
</feed>`)

	feed, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Title != "Example Atom" || len(feed.Items) != 1 {
		t.Fatalf("unexpected feed: %#v", feed)
	}
	if feed.Items[0].URL != "https://example.com/first" || feed.Items[0].PublishedAt.IsZero() {
		t.Fatalf("unexpected item: %#v", feed.Items[0])
	}
	if feed.Items[0].Description != "Hello Atom" || feed.Items[0].ImageURL != "https://cdn.example.com/atom.jpg" {
		t.Fatalf("unexpected item content: %#v", feed.Items[0])
	}
}

func TestParseRSSImageFallbacks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "image enclosure",
			content:  `<enclosure url="https://cdn.example.com/enclosure.jpg" type="image/jpeg"/>`,
			expected: "https://cdn.example.com/enclosure.jpg",
		},
		{
			name:     "relative image in html content",
			content:  `<content:encoded><![CDATA[<p>Story</p><img src="/images/story.jpg">]]></content:encoded>`,
			expected: "https://example.com/images/story.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(`<?xml version="1.0"?>
<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
<channel><title>Example RSS</title><item>
<title>First</title><link>https://example.com/news/first</link>
<description>Story</description>` + tt.content + `
<pubDate>Mon, 01 Sep 2025 12:00:00 +0000</pubDate>
</item></channel></rss>`)

			feed, err := Parse(data)
			if err != nil {
				t.Fatal(err)
			}
			if feed.Items[0].ImageURL != tt.expected {
				t.Fatalf("expected %q, got %#v", tt.expected, feed.Items[0])
			}
		})
	}
}

func TestParseAtomImageEnclosure(t *testing.T) {
	data := []byte(`<?xml version="1.0"?>
<feed xmlns="http://www.w3.org/2005/Atom"><title>Example Atom</title>
<entry><title>First</title>
<link rel="alternate" href="https://example.com/first"/>
<link rel="enclosure" type="image/webp" href="https://cdn.example.com/atom.webp"/>
<summary>Story</summary><updated>2025-09-01T12:00:00Z</updated>
</entry></feed>`)

	feed, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Items[0].ImageURL != "https://cdn.example.com/atom.webp" {
		t.Fatalf("unexpected item: %#v", feed.Items[0])
	}
}

func TestParseRSSDateWithSingleDigitDay(t *testing.T) {
	tests := []struct {
		name string
		date string
	}{
		{name: "numeric timezone", date: "Tue, 1 Sep 2026 05:24:02 +0000"},
		{name: "named timezone", date: "Tue, 1 Sep 2026 00:38:25 EST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parseDate(tt.date)
			if parsed.IsZero() {
				t.Fatalf("expected %q to parse", tt.date)
			}
			if parsed.Day() != 1 || parsed.Month() != time.September || parsed.Year() != 2026 {
				t.Fatalf("unexpected parsed date: %s", parsed)
			}
		})
	}
}

func TestParseRejectsUnknownXML(t *testing.T) {
	if _, err := Parse([]byte(`<html><body>not a feed</body></html>`)); err == nil {
		t.Fatal("expected unsupported feed error")
	}
}
