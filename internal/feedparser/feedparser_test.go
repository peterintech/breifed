package feedparser

import "testing"

func TestParseRSS(t *testing.T) {
	data := []byte(`<?xml version="1.0"?>
<rss version="2.0"><channel><title>Example RSS</title>
<item><title>First</title><link>https://example.com/first</link>
<description>Hello</description><pubDate>Mon, 01 Sep 2025 12:00:00 +0000</pubDate></item>
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
}

func TestParseAtom(t *testing.T) {
	data := []byte(`<?xml version="1.0"?>
<feed xmlns="http://www.w3.org/2005/Atom"><title>Example Atom</title>
<entry><title>First</title><link rel="alternate" href="https://example.com/first"/>
<summary>Hello</summary><updated>2025-09-01T12:00:00Z</updated></entry>
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
}

func TestParseRejectsUnknownXML(t *testing.T) {
	if _, err := Parse([]byte(`<html><body>not a feed</body></html>`)); err == nil {
		t.Fatal("expected unsupported feed error")
	}
}
