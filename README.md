# Briefed

**Your personal RSS feed aggregator.** Subscribe to the blogs and news sources you care about, and read all their articles in one place.

---

## What is Briefed?

Every day, interesting articles are published across hundreds of blogs and news sites. Instead of visiting each one individually, Briefed collects them for you automatically.

**How it works:**
1. You sign up and get a personal API key
2. You tell Briefed which RSS feeds to follow
3. Briefed checks those feeds every minute and saves new articles
4. You ask Briefed for your articles and get everything in one response

**Think of it as your own personal news desk** -- it reads the internet for you and hands you a summary whenever you ask.

---

## Quick Start

### Prerequisites

- Go 1.22 or later
- PostgreSQL 12 or later
- [goose](https://github.com/pressly/goose) for database migrations

### 1. Clone and configure

```bash
git clone https://github.com/peterintech/briefed.git
cd briefed
```

Create a `.env` file:

```
PORT=8080
DB_URL=postgresql://username:password@localhost:5432/briefed?sslmode=disable
```

### 2. Set up the database

```bash
createdb briefed
goose -dir sql/schema postgres "postgresql://username:password@localhost:5432/briefed?sslmode=disable" up
```

### 3. Run the server

```bash
go run .
```

The server starts on `http://localhost:8080`.

---

## API Usage

### Create an account

```bash
curl -X POST http://localhost:8080/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "your-name"}'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "your-name",
  "api_key": "a1b2c3d4e5f6..."
}
```

Save the `api_key` -- you'll need it for everything else.

### Add an RSS feed

```bash
curl -X POST http://localhost:8080/v1/feeds \
  -H "Content-Type: application/json" \
  -H "Authorization: ApiKey YOUR_API_KEY" \
  -d '{"name": "Hacker News", "url": "https://hnrss.org/frontpage"}'
```

### See all available feeds

```bash
curl http://localhost:8080/v1/feeds \
  -H "Authorization: ApiKey YOUR_API_KEY"
```

### Follow a feed

```bash
curl -X POST http://localhost:8080/v1/feed_follows \
  -H "Content-Type: application/json" \
  -H "Authorization: ApiKey YOUR_API_KEY" \
  -d '{"feed_id": "feed-uuid-here"}'
```

### Read your articles

```bash
curl http://localhost:8080/v1/posts \
  -H "Authorization: ApiKey YOUR_API_KEY"
```

Response:
```json
[
  {
    "id": "...",
    "title": "Show HN: A new way to do X",
    "description": "I built a tool that...",
    "url": "https://example.com/article",
    "published_at": "2025-01-15T10:30:00Z",
    "feed_id": "..."
  }
]
```

### Other endpoints

| Action | Method | Endpoint |
|---|---|---|
| Check server health | GET | `/v1/health` |
| Get your profile | GET | `/v1/users` |
| Get a specific feed | GET | `/v1/feeds/{feedId}` |
| Delete a feed | DELETE | `/v1/feeds/{feedId}` |
| List your followed feeds | GET | `/v1/feed_follows` |
| Unfollow a feed | DELETE | `/v1/feed_follows/{feedFollowID}` |

---

## How It Works (for developers)

Briefed is a Go backend with three core components:

### REST API
A `chi`-based HTTP server with API key authentication. All authenticated endpoints use a middleware that extracts the API key from the `Authorization` header and resolves the user from the database.

### Background Scraper
A goroutine that runs on a 60-second ticker. It picks the 10 feeds that haven't been fetched most recently, fetches them concurrently (up to 10 goroutines), parses the RSS XML, and stores new articles. Duplicate articles are silently skipped via a unique constraint on the URL.

### Database Layer
Uses `sqlc` to generate type-safe Go code from raw SQL queries -- no ORM. The schema is managed through `goose` migrations. All relationships cascade on delete (deleting a user removes all their data).

### Tech Stack

| Component | Technology |
|---|---|
| Language | Go |
| Router | chi |
| Database | PostgreSQL |
| Query generation | sqlc |
| Migrations | goose |
| DB driver | lib/pq |
| Auth | API key (SHA-256) |

For detailed architecture diagrams, see [architecture.md](architecture.md).

---

## Project Structure

```
briefed/
├── main.go                  # Entry point
├── handler_*.go             # HTTP request handlers
├── middleware_auth.go        # Authentication middleware
├── scrapper.go              # Background RSS scraper
├── models.go                # API response types
├── internal/
│   ├── auth/                # API key extraction
│   └── database/            # sqlc-generated DB layer
└── sql/
    ├── schema/              # Database migrations
    └── queries/             # SQL queries for sqlc
```

---

## License

MIT
