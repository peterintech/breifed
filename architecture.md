# Briefed - Architecture

## System Overview

Briefed is a backend RSS aggregator that collects articles from RSS feeds and serves them through a REST API. It has two main execution paths: an HTTP server handling client requests, and a background scraper that periodically fetches new articles.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Briefed Server                                  │
│                                                                         │
│  ┌──────────────────────────────┐   ┌────────────────────────────────┐  │
│  │        HTTP Server           │   │      Background Scraper        │  │
│  │                              │   │                                │  │
│  │  chi Router                  │   │  time.Ticker (every 60s)       │  │
│  │    ├── /v1/health            │   │    ├── GetNextFeedsToFetch(10) │  │
│  │    ├── /v1/users             │   │    ├── MarkFeedAsFetched()     │  │
│  │    ├── /v1/feeds             │   │    ├── fetchFeed(url)          │  │
│  │    ├── /v1/feed_follows      │   │    └── CreatePost() per item   │  │
│  │    └── /v1/posts             │   │                                │  │
│  │                              │   │  (10 concurrent goroutines)    │  │
│  └──────────────┬───────────────┘   └──────────────┬─────────────────┘  │
│                 │                                  │                    │
│                 └──────────────┬───────────────────┘                    │
│                                │                                        │
│                 ┌──────────────▼───────────────┐                        │
│                 │     database.Queries          │                        │
│                 │     (sqlc-generated layer)    │                        │
│                 └──────────────┬───────────────┘                        │
│                                │                                        │
└────────────────────────────────┼────────────────────────────────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │   PostgreSQL (port 5432) │
                    │   Database: briefed       │
                    └─────────────────────────┘
```

## Project Structure

```
briefed/
├── main.go                    # Entry point: DB connection, routes, server startup
├── handler_user.go            # User creation and retrieval handlers
├── handler_feed.go            # Feed CRUD handlers
├── handler_feed_follows.go    # Feed follow/unfollow handlers
├── handler_readiness.go       # Health check endpoint
├── middleware_auth.go          # API key authentication middleware
├── scrapper.go                # Background RSS feed scraper
├── models.go                  # API response structs + DB-to-API converters
├── go.mod                     # Module definition and dependencies
├── .env                       # PORT and DB_URL environment variables
├── sqlc.yaml                  # sqlc code generation config
├── internal/
│   ├── auth/
│   │   └── auth.go            # API key extraction from headers
│   └── database/
│       ├── db.go              # DBTX interface, Queries struct, New(), WithTx()
│       ├── models.go          # Generated Go structs for each table
│       ├── users.sql.go       # Generated: CreateUser, GetUserByApiKey
│       ├── feeds.sql.go       # Generated: CreateFeed, GetFeeds, GetNextFeedsToFetch, etc.
│       ├── feed_follows.sql.go# Generated: CreateFeedFollow, GetFeedFollows, DeleteFeedFollow
│       └── posts.sql.go       # Generated: CreatePost, GetPostsForUser
└── sql/
    ├── schema/                # Goose migration files
    │   ├── 001_users.sql
    │   ├── 002_users_apikey.sql
    │   ├── 003_feeds.sql
    │   ├── 004_feed_follows.sql
    │   ├── 005_feeds_lastfetchedat.sql
    │   └── 006_posts.sql
    └── queries/               # SQL input files for sqlc
        ├── users.sql
        ├── feeds.sql
        ├── feed_follows.sql
        └── posts.sql
```

## Request Lifecycle

```
Client Request
     │
     ▼
┌─────────────────────────────┐
│  chi Router                  │
│  └── CORS Middleware         │
└──────────────┬──────────────┘
               │
       ┌───────┴───────┐
       │               │
  Public Route    Protected Route
  (POST /users)   (all others)
       │               │
       │        ┌──────▼──────────────┐
       │        │  authMiddleware     │
       │        │                     │
       │        │ 1. Extract API key  │
       │        │    from "ApiKey"    │
       │        │    header           │
       │        │ 2. DB lookup:       │
       │        │    GetUserByApiKey  │
       │        │ 3. Pass user to     │
       │        │    inner handler    │
       │        └──────┬──────────────┘
       │               │
       └───────┬───────┘
               │
       ┌───────▼───────┐
       │  Route Handler │
       │  (method on    │
       │  *apiConfig)   │
       │                │
       │  ac.DB.Method()│
       └───────┬───────┘
               │
       ┌───────▼───────────────┐
       │  database.Queries     │
       │  (sqlc-generated)     │
       │                       │
       │  Calls:               │
       │  - QueryRowContext()  │
       │  - QueryContext()     │
       │  - ExecContext()      │
       └───────┬───────────────┘
               │
       ┌───────▼───────────────┐
       │  *sql.DB Connection   │
       │  Pool (lib/pq driver) │
       └───────┬───────────────┘
               │
               ▼
         PostgreSQL
```

## Dependency Injection Pattern

The project uses a manual dependency injection pattern via struct method receivers. There is no DI container or context-based injection.

```
main.go
  │
  ├── apiConfig struct
  │     └── DB *database.Queries
  │
  ├── All handlers are methods on *apiConfig:
  │     (ac *apiConfig) createUserHandler(...)
  │     (ac *apiConfig) createFeedHandler(...)
  │     (ac *apiConfig) getFeedsHandler(...)
  │     (ac *apiConfig) createFeedFollowHandler(...)
  │     ...etc
  │
  └── Background scraper receives *database.Queries as a function argument
        startScraping(db *database.Queries, ...)
```

## Code Generation Pipeline (sqlc)

The project uses sqlc instead of an ORM. SQL is written by hand and Go code is generated from it.

```
┌──────────────────────┐     ┌──────────────────────┐
│  sql/schema/*.sql    │     │  sql/queries/*.sql    │
│  (table definitions) │     │  (query definitions)  │
└──────────┬───────────┘     └──────────┬───────────┘
           │                            │
           └──────────┬─────────────────┘
                      │
               ┌──────▼──────┐
               │  sqlc generate  │
               │  (sqlc.yaml)    │
               └──────┬──────┘
                      │
           ┌──────────▼───────────┐
           │  internal/database/  │
           │                      │
           │  ├── db.go           │  ← Interfaces + constructor
           │  ├── models.go       │  ← Go structs matching tables
           │  ├── users.sql.go    │
           │  ├── feeds.sql.go    │  ← Type-safe query methods
           │  ├── feed_follows.sql.go
           │  └── posts.sql.go    │
           └──────────────────────┘
```

## Scraper Architecture

The background scraper runs in a separate goroutine and fetches RSS feeds on a configurable interval.

```
┌────────────────────────────────────────────────────────────────┐
│                     startScraping()                             │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  for {                                                   │   │
│  │    <-ticker.C  (every 60 seconds)                        │   │
│  │                                                          │   │
│  │    1. GetNextFeedsToFetch(10)                            │   │
│  │       → Returns feeds ordered by last_fetched_at ASC     │   │
│  │       → Limited to 10 (concurrency limit)                │   │
│  │                                                          │   │
│  │    2. For each feed, launch goroutine:                   │   │
│  │       ┌──────────────────────────────────────────────┐   │   │
│  │       │  scrapeFeed(db, wg, feed)                     │   │   │
│  │       │                                               │   │   │
│  │       │  a. MarkFeedAsFetched(feed.ID)                │   │   │
│  │       │  b. fetchFeed(feed.Url)                       │   │   │
│  │       │     → HTTP GET (10s timeout)                  │   │   │
│  │       │     → XML unmarshal                           │   │   │
│  │       │  c. For each RSS item:                        │   │   │
│  │       │     → Parse pubDate (RFC1123Z or RFC1123)     │   │   │
│  │       │     → CreatePost(title, url, desc, feed_id)  │   │   │
│  │       │     → Skip duplicates (unique constraint)     │   │   │
│  │       └──────────────────────────────────────────────┘   │   │
│  │                                                          │   │
│  │    3. wg.Wait() — block until all goroutines finish      │   │
│  │  }                                                       │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────┘
```

## Database Schema

### Entity Relationship Diagram

```
┌──────────────────────┐
│        users          │
├──────────────────────┤
│ PK │ id         UUID │
│    │ created_at  TS  │
│    │ updated_at  TS  │
│    │ name        TEXT│
│ UQ │ api_key  VARCHAR│
└──────┬───────┬───────┘
       │       │
       │       │  1:N (user_id FK)
       │       │
       │       │  N:1 (user_id FK)
       │    ┌──▼──────────────────┐
       │    │   feed_follows       │
       │    ├─────────────────────┤
       │    │ PK │ id        UUID │
       │    │    │ created_at  TS │
       │    │    │ updated_at  TS │
       │    │ FK │ user_id   UUID │──┐
       │    │ FK │ feed_id   UUID │──┼──┐
       │    │ UQ │ (feed_id,      │  │  │
       │    │     │  user_id)     │  │  │
       │    └────┬────────────────┘  │  │
       │         │                   │  │
       │  1:N    │                   │  │  N:1
       │  (user_id FK)               │  │  (feed_id FK)
       │    ┌────▼───────────────┐   │  │
       │    │      feeds          │   │  │
       │    ├────────────────────┤   │  │
       │    │PK│ id         UUID │◄──┘  │
       │    │  │ created_at   TS │      │
       │    │  │ updated_at   TS │      │
       │    │  │ name       TEXT │      │
       │    │UQ│ url        TEXT │      │
       │    │FK│ user_id   UUID │◄─────┘
       │    │  │ description TXT │
       │    │  │ last_fetched_at │
       │    └────┬───────────────┘
       │         │
       │  1:N    │  (feed_id FK)
       │    ┌────▼──────────────┐
       │    │      posts         │
       │    ├───────────────────┤
       │    │PK│ id        UUID │
       │    │  │ created_at  TS │
       │    │  │ updated_at  TS │
       │    │  │ title      TXT │
       │    │  │ description TXT │
       │    │UQ│ published_at TS │
       │    │UQ│ url        TXT │
       │    │FK│ feed_id   UUID │
       │    └───────────────────┘
```

### Relationship Summary

| Relationship         | Type         | FK                                | On Delete |
| -------------------- | ------------ | --------------------------------- | --------- |
| users → feeds        | One-to-Many  | `feeds.user_id → users.id`        | CASCADE   |
| users → feed_follows | One-to-Many  | `feed_follows.user_id → users.id` | CASCADE   |
| feeds → feed_follows | One-to-Many  | `feed_follows.feed_id → feeds.id` | CASCADE   |
| feeds → posts        | One-to-Many  | `posts.feed_id → feeds.id`        | CASCADE   |
| users ↔ feeds        | Many-to-Many | via `feed_follows` junction table | —         |

All foreign keys use `ON DELETE CASCADE`. Deleting a user removes all their feeds, follows, and (transitively) posts. Deleting a feed removes all its follows and posts.

## Authentication Flow

```
Client                              Server
  │                                   │
  │  POST /users { "name": "..." }   │  ← Unauthenticated
  │ ────────────────────────────────▶ │
  │                                   │  CreateUser()
  │  ◀──── { "api_key": "abc123" }   │
  │                                   │
  │  (client stores api_key)          │
  │                                   │
  │  GET /users                       │  ← Authenticated
  │  Header: ApiKey abc123            │
  │ ────────────────────────────────▶ │
  │                                   │  authMiddleware:
  │                                   │    GetApiKey(header)
  │                                   │    GetUserByApiKey(key)
  │  ◀──── { "id":..., "name":... }  │
```

The API key is a SHA-256 hash auto-generated at user creation time. It is transmitted in the `Authorization` header using the format `ApiKey <key>`.

## Middleware Stack

```
Request
  │
  ▼
┌─────────────────────┐
│  CORS Middleware     │  ← chi/cors (allows all origins)
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  authMiddleware      │  ← Custom (for protected routes only)
│                      │
│  1. Extract API key  │
│  2. DB user lookup   │
│  3. Inject user      │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Route Handler       │
│  (receives user)     │
└─────────────────────┘
```

## Model Layer

The project maintains two sets of structs: sqlc-generated models (in `internal/database/models.go`) and API-facing models (in `models.go`). Converter functions bridge the two.

```
┌──────────────────────┐         ┌──────────────────────┐
│  database.User       │  ──▶   │  User                 │
│  (sqlc-generated)    │         │  (API response)       │
└──────────────────────┘         └──────────────────────┘

┌──────────────────────┐         ┌──────────────────────┐
│  database.CreateFeed │  ──▶   │  Feed                 │
│  Row                 │         │  (API response)       │
└──────────────────────┘         └──────────────────────┘

┌──────────────────────┐         ┌──────────────────────┐
│  database.FeedFollow │  ──▶   │  FeedFollow           │
│                      │         │  (API response)       │
└──────────────────────┘         └──────────────────────┘

┌──────────────────────┐         ┌──────────────────────┐
│  database.Post       │  ──▶   │  Post                 │
│                      │         │  (API response)       │
└──────────────────────┘         └──────────────────────┘
```

This separation keeps sqlc-generated code untouched while giving full control over JSON serialization tags and nullable field handling.

## API Endpoints

| Method | Path                              | Auth | Handler                   | Description                    |
| ------ | --------------------------------- | ---- | ------------------------- | ------------------------------ |
| GET    | `/v1/health`                      | No   | `readinessHandler`        | Liveness probe                 |
| GET    | `/v1/err`                         | No   | `errorHandler`            | Test error handling            |
| POST   | `/v1/users`                       | No   | `createUserHandler`       | Register new user              |
| GET    | `/v1/users`                       | Yes  | `getUserByApiKey`         | Get current user profile       |
| POST   | `/v1/feeds`                       | Yes  | `createFeedHandler`       | Create a new feed subscription |
| GET    | `/v1/feeds`                       | Yes  | `getFeedsHandler`         | List user's feeds              |
| GET    | `/v1/feeds/{feedId}`              | Yes  | `getFeedByIdHandler`      | Get a specific feed            |
| DELETE | `/v1/feeds/{feedId}`              | Yes  | `deleteFeedHandler`       | Delete a feed                  |
| POST   | `/v1/feed_follows`                | Yes  | `createFeedFollowHandler` | Follow a feed                  |
| GET    | `/v1/feed_follows`                | Yes  | `getFeedFollowsHandler`   | List followed feeds            |
| DELETE | `/v1/feed_follows/{feedFollowID}` | Yes  | `deleteFeedFollowHandler` | Unfollow a feed                |
| GET    | `/v1/posts`                       | Yes  | `getPostsForUserHandler`  | Get posts from followed feeds  |

## Dependencies

| Package                    | Purpose                   |
| -------------------------- | ------------------------- |
| `github.com/go-chi/chi`    | HTTP router               |
| `github.com/go-chi/cors`   | CORS middleware           |
| `github.com/lib/pq`        | PostgreSQL driver         |
| `github.com/google/uuid`   | UUID generation           |
| `github.com/joho/godotenv` | .env file loading         |
| `github.com/pressly/goose` | Database migrations       |
| `github.com/sqlc-dev/sqlc` | SQL-to-Go code generation |
