# Briefed

Briefed is a small RSS/Atom aggregator that gives each user a personalized, newest-first news feed.

Users choose interests, discover available feeds in those categories, and explicitly follow the sources they want. Every feed is shared: a feed contributed by one user can be followed by everyone.

## Stack

- Go and Chi
- PostgreSQL
- sqlc for generated database access
- goose for migrations
- Cookie sessions with bcrypt passwords
- templ and HTMX will be added after the backend

## Setup

Requirements:

- Go 1.26 or later
- PostgreSQL
- goose
- sqlc

Create a `.env` file:

```env
PORT=8080
DB_URL=postgresql://username:password@localhost:5432/briefed?sslmode=disable
```

Create a new database, run the migrations, and generate the database package:

```bash
goose -dir sql/schema postgres "$DB_URL" up
sqlc generate
go run .
```

## Main flow

1. Load categories with `GET /v1/categories`.
2. Find feeds using `GET /v1/feeds?category_ids=...`.
3. Register with the selected category and feed IDs.
4. The server creates an HTTP-only session cookie.
5. Load personalized articles with `GET /v1/posts`.
6. Follow existing feeds or contribute a new RSS/Atom URL.

The browser must send the session cookie for authenticated endpoints.

## API

### Register

```bash
curl -c cookies.txt -X POST http://localhost:8080/v1/auth/register   -H "Content-Type: application/json"   -d '{
    "name": "Peter",
    "email": "peter@example.com",
    "password": "password123",
    "category_ids": [],
    "feed_ids": []
  }'
```

### Login and logout

```bash
curl -c cookies.txt -X POST http://localhost:8080/v1/auth/login   -H "Content-Type: application/json"   -d '{"email":"peter@example.com","password":"password123"}'

curl -b cookies.txt -X POST http://localhost:8080/v1/auth/logout
```

### Discover feeds

```bash
curl "http://localhost:8080/v1/feeds?category_ids=CATEGORY_UUID&q=tech&limit=20&offset=0"
```

Multiple category IDs are comma-separated and use OR matching.

### Follow or unfollow

```bash
curl -b cookies.txt -X POST http://localhost:8080/v1/me/feeds/FEED_UUID
curl -b cookies.txt -X DELETE http://localhost:8080/v1/me/feeds/FEED_UUID
```

### Contribute a feed

```bash
curl -b cookies.txt -X POST http://localhost:8080/v1/feeds   -H "Content-Type: application/json"   -d '{
    "url": "https://example.com/feed.xml",
    "category_ids": ["CATEGORY_UUID"]
  }'
```

The URL is parsed before insertion. The feed title is extracted from the RSS/Atom document, the feed becomes globally available, and the submitter follows it automatically.

### Personalized posts

```bash
curl -b cookies.txt "http://localhost:8080/v1/posts?limit=20&offset=0"
```

Posts come only from feeds followed by the authenticated user and are ordered newest first.

## Endpoints

| Method | Path | Auth | Description |
|---|---|---:|---|
| GET | `/v1/health` | No | Health check |
| GET | `/v1/categories` | No | List seeded categories |
| GET | `/v1/feeds` | No | Search/filter the shared feed catalog |
| GET | `/v1/feeds/{feedID}` | No | Get a shared feed |
| POST | `/v1/auth/register` | No | Register and save onboarding choices |
| POST | `/v1/auth/login` | No | Log in |
| POST | `/v1/auth/logout` | No | Delete the current session |
| GET | `/v1/me` | Yes | Get profile, interests, and followed feeds |
| PUT | `/v1/me/categories` | Yes | Replace selected interests |
| GET | `/v1/me/feeds` | Yes | List followed feeds |
| POST | `/v1/me/feeds/{feedID}` | Yes | Follow a feed |
| DELETE | `/v1/me/feeds/{feedID}` | Yes | Unfollow a feed |
| POST | `/v1/feeds` | Yes | Validate and contribute a feed |
| GET | `/v1/posts` | Yes | Get personalized posts |

## Development notes

Write schema changes under `sql/schema` and queries under `sql/queries`. Never manually edit `internal/database`; run `sqlc generate` instead.

Categories and starter feeds are seeded separately. There is no category administration API in this version.
