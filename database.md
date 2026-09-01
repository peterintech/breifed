# Briefed — Database Documentation

This document describes the PostgreSQL database that backs the Briefed RSS aggregator. It covers the schema, every table, its columns, data types, constraints, relationships, and how the database is accessed from the application (sqlc + goose).

## Overview

- **Engine:** PostgreSQL
- **Database name:** `briefed`
- **Connection string:** provided via the `DB_URL` environment variable in `.env`
- **Driver:** `github.com/lib/pq` (Go)
- **Migrations:** `github.com/pressly/goose` — SQL files under `sql/schema/`
- **Code generation:** `github.com/sqlc-dev/sqlc` — generates Go code into `internal/database/` from the schemas and queries

The schema is defined as a set of numbered goose migration files applied in order:

```
sql/schema/
├── 001_users.sql
├── 002_users_apikey.sql
├── 003_feeds.sql
├── 004_feed_follows.sql
├── 005_feeds_lastfetchedat.sql
├── 006_posts.sql
└── 007_feeds_category.sql
```

## Entity Relationship Diagram

```
┌────────────────────────────┐
│           users            │
├────────────────────────────┤
│ PK │ id         UUID       │
│    │ created_at TIMESTAMP  │
│    │ updated_at TIMESTAMP  │
│    │ name       TEXT       │
│ UQ │ api_key    VARCHAR(64)│
└──────┬──────────────┬──────┘
       │              │
       │ 1:N          │ 1:N
       │ (user_id FK) │ (user_id FK)
       │              │
       ▼              ▼
┌──────────────┐   ┌──────────────────┐
│    feeds     │   │   feed_follows   │
├──────────────┤   ├──────────────────┤
│PK│ id     UUID│   │PK│ id        UUID│
│  │ created_at │   │  │ created_at TS │
│  │ updated_at │   │  │ updated_at TS │
│  │ name   TEXT│   │FK│ feed_id   UUID│──┐
│UQ│ url    TEXT│   │FK│ user_id   UUID│  │
│FK│ user_id UUID│◄─┼──┘              │  │
│  │ last_fetched │   │UQ│ (feed_id,   │  │
│  │   at         │   │   │  user_id)  │  │
│  │ category     │   └────────────────┘  │
│  │  VARCHAR[]   │                       │
└──────┬──────────┘                       │
       │ 1:N (feed_id FK)                 │ N:1 (feed_id FK)
       ▼                                  │
┌──────────────┐                          │
│    posts     │                          │
├──────────────┤                          │
│PK│ id     UUID│                         │
│  │ created_at │                         │
│  │ updated_at │                         │
│  │ title  TEXT│                         │
│  │ desc   TEXT│                         │
│UQ│pub_at  TS  │                         │
│UQ│ url    TEXT│                         │
│FK│ feed_id UUID│────────────────────────┘
└──────────────┘
```

## Tables

### `users`

Represents an API client / registered user.

| Column      | Type                    | Constraints               | Description                                  |
| ----------- | ----------------------- | ------------------------- | -------------------------------------------- |
| `id`        | `UUID`                  | `PRIMARY KEY`             | Unique identifier of the user                |
| `created_at`| `TIMESTAMP`             | `NOT NULL DEFAULT NOW()`  | Row creation timestamp                       |
| `updated_at`| `TIMESTAMP`             | `NOT NULL DEFAULT NOW()`  | Row last-modified timestamp                  |
| `name`      | `TEXT`                  | `NOT NULL`                | Display name of the user                     |
| `api_key`   | `VARCHAR(64)`           | `UNIQUE NOT NULL`         | SHA-256 hex (64 chars) key used for auth; auto-generated |

Notes:
- `api_key` has no explicit database-level `DEFAULT` in the schema file, but it is auto-generated at insert time by the application (`CreateUser` query generates `encode(sha256(random()::text::bytea), 'hex')`).
- The migration `002_users_apikey.sql` adds `api_key` to the original `users` table.

Go model: `database.User` (`internal/database/models.go`).

---

### `feeds`

Represents an RSS feed that has been added/subscribed.

Created by migration `003_feeds.sql`; extended by `005_feeds_lastfetchedat.sql` (`last_fetched_at`) and `007_feeds_category.sql` (`category`).

| Column           | Type                    | Constraints               | Description                                         |
| ---------------- | ----------------------- | ------------------------- | --------------------------------------------------- |
| `id`             | `UUID`                  | `PRIMARY KEY`             | Unique identifier of the feed                       |
| `created_at`     | `TIMESTAMP`             | `NOT NULL DEFAULT NOW()`  | Row creation timestamp                              |
| `updated_at`     | `TIMESTAMP`             | `NOT NULL DEFAULT NOW()`  | Row last-modified timestamp (bumped on fetch)       |
| `name`           | `TEXT`                  | `NOT NULL`                | Display name of the feed                            |
| `url`            | `TEXT`                  | `UNIQUE NOT NULL`         | The RSS feed URL (globally unique)                  |
| `user_id`        | `UUID`                  | `NOT NULL REFERENCES users(id) ON DELETE CASCADE` | Owning user            |
| `last_fetched_at`| `TIMESTAMP WITH TIME ZONE` | (nullable)             | When the feed was last scraped (used by scheduler)  |
| `category`       | `VARCHAR(100)[]`        | (nullable)                | Array of category tags for the feed                 |

Notes:
- `last_fetched_at` is used by `GetNextFeedsToFetch` to prioritize feeds that have not been fetched recently (`ORDER BY last_fetched_at ASC NULLS FIRST`).
- `category` is a PostgreSQL array column (`VARCHAR(100)[]`); it is not currently reflected in the sqlc-generated `database.Feed` model.

Go model: `database.Feed` (`internal/database/models.go`).

---

### `feed_follows`

A junction / join table implementing a many-to-many relationship between `users` and `feeds`. A user can follow many feeds, and a feed can be followed by many users.

Created by migration `004_feed_follows.sql`.

| Column     | Type                    | Constraints               | Description                              |
| ---------- | ----------------------- | ------------------------- | ---------------------------------------- |
| `id`       | `UUID`                  | `PRIMARY KEY`             | Unique identifier of the follow record   |
| `created_at`| `TIMESTAMP`            | `NOT NULL DEFAULT NOW()`  | Row creation timestamp                   |
| `updated_at`| `TIMESTAMP`            | `NOT NULL DEFAULT NOW()`  | Row last-modified timestamp              |
| `feed_id`  | `UUID`                  | `NOT NULL REFERENCES feeds(id) ON DELETE CASCADE` | The followed feed        |
| `user_id`  | `UUID`                  | `NOT NULL REFERENCES users(id) ON DELETE CASCADE` | The following user        |

Constraints:
- `UNIQUE (feed_id, user_id)` — a user can follow a given feed only once.

Go model: `database.FeedFollow` (`internal/database/models.go`).

---

### `posts`

Represents an individual article/item scraped from an RSS feed.

Created by migration `006_posts.sql`.

| Column         | Type                         | Constraints                | Description                          |
| -------------- | ---------------------------- | -------------------------- | ------------------------------------ |
| `id`           | `UUID`                       | `PRIMARY KEY`              | Unique identifier of the post        |
| `created_at`   | `TIMESTAMP WITH TIME ZONE`   | `NOT NULL DEFAULT NOW()`   | Row creation timestamp               |
| `updated_at`   | `TIMESTAMP WITH TIME ZONE`   | `NOT NULL DEFAULT NOW()`   | Row last-modified timestamp          |
| `title`        | `TEXT`                       | `NOT NULL`                 | Article title                        |
| `description`  | `TEXT`                       | (nullable)                 | Article summary / description        |
| `published_at` | `TIMESTAMP WITH TIME ZONE`   | `NOT NULL UNIQUE`          | Article publication date (unique)    |
| `url`          | `TEXT`                       | `NOT NULL UNIQUE`          | Link to the full article (unique)    |
| `feed_id`      | `UUID`                       | `NOT NULL REFERENCES feeds(id) ON DELETE CASCADE` | Source feed          |

Notes:
- `published_at` and `url` are each `UNIQUE`; the unique constraints naturally prevent duplicate posts from being inserted when the same feed is scraped repeatedly.
- Unlike the other tables, timestamps here are `TIMESTAMP WITH TIME ZONE` (with timezone) rather than plain `TIMESTAMP`.

Go model: `database.Post` (`internal/database/models.go`).

---

## Relationships Summary

| Relationship                  | Type        | Foreign Key                              | On Delete |
| ----------------------------- | ----------- | ---------------------------------------- | --------- |
| users → feeds                 | One-to-Many | `feeds.user_id → users.id`               | CASCADE   |
| users → feed_follows          | One-to-Many | `feed_follows.user_id → users.id`        | CASCADE   |
| feeds → feed_follows          | One-to-Many | `feed_follows.feed_id → feeds.id`        | CASCADE   |
| feeds → posts                 | One-to-Many | `posts.feed_id → feeds.id`               | CASCADE   |
| users ↔ feeds (via follows)   | Many-to-Many| `feed_follows` junction table            | —         |

Key points:

- **All foreign keys use `ON DELETE CASCADE`:**
  - Deleting a `user` removes their `feeds`, their `feed_follows`, and (transitively via feeds) their `posts`.
  - Deleting a `feed` removes its `feed_follows` and its `posts`.
  - Deleting a `feed_follow` removes only that follow record.
- **Many-to-many** between `users` and `feeds` is expressed through the `feed_follows` junction table (a user follows a feed → a row in `feed_follows`).
- Note the schema is slightly redundant: a feed already has an owning `user_id` (`feeds.user_id`), and `feed_follows` additionally tracks followership. Deleting a user therefore also cascades through both paths.

## Access Layer & Tooling

### Migrations (goose)

Each numbered file under `sql/schema/` contains `-- +goose Up` and `-- +goose Down` sections and is applied in filename order:

| Migration | Purpose |
| --------- | ------- |
| `001_users.sql` | Create `users` table |
| `002_users_apikey.sql` | Add `api_key` column to `users` |
| `003_feeds.sql` | Create `feeds` table |
| `004_feed_follows.sql` | Create `feed_follows` junction table |
| `005_feeds_lastfetchedat.sql` | Add `last_fetched_at` to `feeds` |
| `006_posts.sql` | Create `posts` table |
| `007_feeds_category.sql` | Add `category` array column to `feeds` |

### Code Generation (sqlc)

Configured in `sqlc.yaml`:
- `schema: "sql/schema"` — the goose migration files as the source of truth
- `queries: "sql/queries"` — hand-written SQL queries
- `engine: "postgresql"`
- Generated Go output: `internal/database/`

Query files:

| File                      | Functions |
| ------------------------- | --------- |
| `sql/queries/users.sql`   | `CreateUser`, `GetUserByApiKey` |
| `sql/queries/feeds.sql`   | `CreateFeed`, `GetFeeds`, `GetFeedsByUserId`, `GetFeedById`, `GetNextFeedsToFetch`, `MarkFeedAsFetched`, `DeleteFeed` |
| `sql/queries/feed_follows.sql` | `CreateFeedFollow`, `GetFeedFollows`, `DeleteFeedFollow` |
| `sql/queries/posts.sql`   | `CreatePost`, `GetPostsForUser` |

Generated Go artifacts in `internal/database/`:
- `db.go` — `DBTX` interface, `Queries` struct, `New()`, `WithTx()`
- `models.go` — Go structs mirroring each table (`User`, `Feed`, `FeedFollow`, `Post`)
- `users.sql.go`, `feeds.sql.go`, `feed_follows.sql.go`, `posts.sql.go` — type-safe query methods

### Scraper usage

The background scraper (`internal/scrapper/scrapper.go`) periodically selects feeds to fetch via `GetNextFeedsToFetch` (ordered by `last_fetched_at ASC NULLS FIRST`, limited to 10), marks them fetched with `MarkFeedAsFetched`, and inserts each article found via `CreatePost`. Duplicate detection relies on the `posts.url` / `posts.published_at` unique constraints.

## Notable Data Details

- **API keys** are 64-character SHA-256 hex digests, generated in SQL at user creation time and looked up during authentication (`GetUserByApiKey`).
- **Nullable fields in Go:** `Feed.LastFetchedAt` maps to `sql.NullTime`, `Post.Description` maps to `sql.NullString`. The `category` array column on `feeds` is not yet exposed in the generated model.
- **Duplicate posts** are naturally avoided by unique constraints on `posts.url` and `posts.published_at`, though `CreatePost` currently relies on these constraints to skip duplicates (duplicate inserts will error rather than be silently ignored).
