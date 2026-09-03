# Briefed

Briefed is a source-first RSS and Atom reader that gives people useful news before asking them to create an account. Anyone can read the newest stories in the shared catalog; signed-in readers shape that same timeline by choosing interests and explicitly following publications they trust.

This project is more than an RSS parser. It explores how product decisions become data relationships, transaction boundaries, HTTP behavior, and interface architecture. Its central question is: **how small can the system remain without making its behavior accidental?**

The answer is a modular Go monolith: one deployable process serves HTML and JSON, owns authentication and business rules, and runs a bounded background collector against PostgreSQL. It favors explicit SQL and server-rendered HTML over abstractions its current scale does not require.

For the relational model, transaction analysis, constraints, and scaling triggers, read [the database design notes](database.md).

## Reading guide

- [Product thinking](#product-thinking) and [engineering highlights](#engineering-highlights)
- [Architecture](#architecture) and [design decisions](#design-decisions)
- [Atomicity](#atomicity-protects-product-invariants) and [ingestion](#feed-contribution-and-ingestion)
- [Deliberate boundaries](#deliberate-boundaries), [local setup](#local-development), and [testing](#testing)

## Product thinking

RSS readers often demand configuration before demonstrating value. Briefed reverses that sequence:

1. A visitor immediately receives a global, newest-first timeline.
2. Interests narrow source discovery but never silently subscribe the reader.
3. The reader chooses sources and registers only to preserve those choices.
4. A signed-in reader receives posts from followed feeds and can revise them.
5. A missing RSS/Atom source can be contributed and becomes available to everyone after successful parsing.

This moves progressively from **public value** to **intentional personalization**. Categories describe interests; follows express subscriptions. The visual product applies similar restraint through editorial hierarchy, visible sources, and no social engagement mechanics. See [PRODUCT.md](PRODUCT.md) and [DESIGN.md](DESIGN.md).

## Engineering highlights

- **One timeline, three modes.** The homepage and `GET /v1/posts` resolve to `global`, `personalized`, or `global_fallback`.
- **Server-owned UI state.** Go and templ render pages and fragments; HTMX requests targeted updates.
- **Shared catalog.** Feeds are global, while `feed_follows` represents personal subscription.
- **Selective atomicity.** Registration, preference replacement, and new-feed contribution are transactional.
- **Typed, visible SQL.** Queries remain handwritten and sqlc generates Go methods and types.
- **Normalized ingestion.** One parser handles RSS, Atom, date variations, excerpts, and images.
- **Retry-safe writes.** Composite keys and `ON CONFLICT DO NOTHING` make repeated work idempotent.
- **Bounded collection.** Small concurrent batches isolate individual publisher failures.

## Technology choices

| Concern | Choice | Why it fits Briefed |
|---|---|---|
| Application | Go + Chi | A small runtime, explicit concurrency, and composable `net/http` routing. |
| Database | PostgreSQL | Transactions, constraints, joins, and conflict handling match the domain. |
| SQL workflow | goose + sqlc | Reviewable migrations and queries with generated Go type safety. |
| HTML | templ | Typed, composable components compiled to Go. |
| Interactions | HTMX | The server returns HTML instead of maintaining duplicate client state. |
| Styling | Tailwind v4 | Static CSS from templ classes; the standalone CLI avoids a Node-only toolchain. |
| Authentication | Opaque sessions | Straightforward expiry and revocation without an unnecessary JWT claim lifecycle. |

## Architecture

```text
Browser
  ├─ pages ─────────────> handlers ─────> views
  ├─ HTMX fragments ────> handlers ─────> components
  └─ JSON /v1 ──────────> internal/api
                              │
                              v
                 shared application operations
        accounts · timeline · preferences · feedcatalog
                              │
                              v
                   sqlc queries → PostgreSQL
                              ^
                              │
             scraper ─────────┘
                │
                └─ fetch/parse through shared feed parser
```

### Why a modular monolith

Briefed has one cohesive domain, one database, and a modest workload. Splitting it into services would add network failures, distributed transactions, duplicated deployment configuration, and harder local development without creating a useful ownership or scaling boundary.

The monolith is still structured:

- `internal/accounts` owns registration and login.
- `internal/sessionauth` owns tokens, cookies, and session resolution.
- `internal/timeline` chooses the timeline and hydrates post categories.
- `internal/preferences` atomically replaces interests and follows.
- `internal/feedcatalog` owns shared-feed contribution.
- `internal/feedparser` normalizes RSS and Atom.
- `internal/scrapper` owns background collection.
- `internal/database` contains generated sqlc code only.

`handlers` serves pages and HTMX fragments; `internal/api` serves JSON. Both reuse the same application operations instead of reimplementing transaction and validation rules. This keeps responsibilities visible without placing a repository or service framework around generated queries.

### HTML, HTMX, and timeline flow

`GET /` renders the document and first timeline page. Search and category changes replace only the posts fragment; load-more appends rows while meaningful filters stay in the URL. Anonymous onboarding is a three-step dialog. Preferences and feed contribution use drawers. Small browser scripts handle focus, opening/closing overlays, and draft interaction state; persisted account and feed state stays in PostgreSQL. Server handlers validate submitted data. This is not a claim that HTMX removes JavaScript—it reduces the client-side application code needed for request/response interactions.

```text
GET / or GET /v1/posts
       ├─ no/invalid/expired session ──> global
       ├─ valid session + follows ─────> personalized
       └─ valid session + no follows ──> global_fallback
```

A stale cookie should not block a public page, so optional authentication treats it as anonymous. Protected endpoints return `401`. Unexpected database failures remain server errors rather than being hidden as anonymous state.

## Design decisions

### Global feeds, explicit follows

A feed URL identifies a publication, not a private user resource. A copy per subscriber would duplicate fetching, posts, metadata, and failure state. Briefed stores one feed and models subscriptions through `feed_follows`.

`submitted_by` records attribution, not ownership. `ON DELETE SET NULL` preserves a useful feed if its contributor leaves, while followers remain independent.

### Interests discover; follows personalize

Interest in “AI” should reveal AI publications, not subscribe somebody to every current and future AI source. `user_categories` and `feed_follows` encode different intentions and prevent surprising timeline changes.

### Sessions instead of JWTs

Registration or login creates a cryptographically random token, stores an expiring session, and sends an HTTP-only, SameSite=Lax cookie. The database lookup is an accepted cost: it enables immediate logout and revocation, exposes no identity claims to the client, and avoids JWT signing, rotation, and stale-claim concerns. JWTs become useful when independent services must verify identity without a shared store; Briefed has no such boundary.

### Simple pagination and ranking

Limit/offset keeps URLs and HTMX pagination understandable. Queries request `limit + 1` and derive `has_more` without `COUNT(*)`. Newest-first ranking is transparent. Cursor pagination and relevance scoring wait for measured deep-page cost or product evidence.

### Typed templates without duplicating application state

templ makes page and fragment contracts typed Go functions, so component inputs participate in compilation and refactoring. Go’s `html/template` would also be a valid, contextually escaped rendering choice; templ is selected for composition and editor ergonomics, not because the standard library is unsafe. The cost is a code-generation step and another tool version to manage.

HTMX fits operations that already map to HTTP: search, pagination, forms, and replacement fragments. A React client would add value for richer client-owned or offline state, but this reader primarily renders database-backed content. Keeping the render decision on the server avoids maintaining the same form and preference rules in two application layers.

### Tailwind without Node

The standalone Tailwind compiler watches development sources and emits minified `public/styles.css` for production. The project gets utility-first CSS without Node, `package.json`, or `node_modules` solely for compilation.

## Atomicity protects product invariants

| Operation | Writes committed together | Partial state prevented |
|---|---|---|
| Registration | user, categories, follows, session | An account with incomplete onboarding or no promised session. |
| Preference replacement | remove old selections, insert all replacements | A mixture of old and new interests and subscriptions. |
| New contribution | feed, categories, contributor follow | A public but partially classified or unfollowed source. |

Each path uses sqlc through `WithTx`, defers rollback, and commits only after establishing the complete invariant. The separate `PUT /v1/me/categories` endpoint also wraps its delete-and-reinsert operation in a transaction; replacing one collection still spans multiple statements. Password hashing, token generation, and network feed parsing happen before transactions so database resources are not held during unrelated work.

Login reads the user and verifies the password, then performs one database write to create the session. Follow and unfollow each perform one relationship mutation. PostgreSQL already makes each individual statement atomic. The rule is: **transaction boundaries follow business operations, not function boundaries.** See [the database analysis](database.md#transaction-boundaries).

## Feed contribution and ingestion

A contribution first checks the globally unique URL. Existing feeds are followed idempotently. New URLs are fetched and parsed before insertion; the feed, its categories, and the follow then commit together. A uniqueness race falls back to the now-existing feed.

The shared parser accepts RSS 2.0 and Atom, handles common ISO/RFC dates, selects Atom alternate links, normalizes HTML descriptions, discovers preview images across common conventions, and resolves relative image URLs. Posts without a title, URL, or publication time are skipped.

The collector orders work by `last_fetched_at ASC NULLS FIRST`, prioritizing never-attempted feeds. It runs a bounded concurrent batch and marks every fetch attempt, even a failure, so one broken source cannot remain first and starve others. Failures are logged per feed or article. Unique article URLs plus `ON CONFLICT DO NOTHING` prevent duplicate stored articles on repeated collection. This does not guarantee exactly-once fetching or update previously stored article content. The collector currently starts immediately, selects up to ten feeds, waits for the batch, and uses a one-minute ticker; it does not promise that every source refreshes once per minute.

## Deliberate boundaries

One process and a small catalog do not yet justify distributed leases, queues, conditional requests, exponential backoff, relevance ranking, cursor pagination, multi-instance coordination, or repository abstractions.

| Evidence that changes the trade-off | Likely response |
|---|---|
| Multiple collectors duplicate work | Dedicated worker plus database claims or leases. |
| Ingestion delay or retry volume grows | Queue, retry state, backoff, and conditional requests. |
| Global sorting dominates query time | Measure with `EXPLAIN (ANALYZE, BUFFERS)` and add a publication-led index if warranted. |
| Deep offsets become slow or unstable | Keyset pagination on `(published_at, id)`. |
| URL variants duplicate stories | Canonical URL or reliable publisher GUID identity. |
| Chronology no longer satisfies readers | Explainable ranking using follows and interests. |

Independent of scale, an internet-facing deployment needs a security review. Current cookie `Secure` behavior depends on `r.TLS`, so HTTPS terminated by a reverse proxy needs explicit verification; SameSite is not a substitute for a complete CSRF strategy. I would also hash session tokens at rest, clean expired sessions, rate-limit authentication and contribution, constrain user-supplied feed destinations and response sizes, add structured observability and graceful shutdown, expand PostgreSQL integration coverage, and define backup and migration rollout procedures.

These are known boundaries with concrete triggers, not claims that production concerns do not exist.

## Repository map

```text
components/          reusable templ and HTMX fragments
handlers/            HTML routes and rendering
internal/api/        JSON API and middleware
internal/database/   sqlc-generated code; never edit manually
internal/*           focused shared application operations
public/              compiled CSS and browser assets
sql/schema/          goose migrations
sql/queries/         handwritten sqlc queries
sql/seeds/           shared catalog seed
types/               view models
views/               full templ pages and Tailwind source
```

## Local development

Requirements: Go 1.26.4+, PostgreSQL, goose, sqlc, the module-tracked templ tool, and the [Tailwind standalone CLI](https://tailwindcss.com/docs/installation/tailwind-cli) available as `tailwindcss`. Node, Corepack, and pnpm are not required.

Create `.env` for the application:

```env
PORT=8080
DB_URL=postgresql://username:password@localhost:5432/briefed?sslmode=disable
```

The application loads `.env`, but your shell must have `DB_URL` set before expanding `"$DB_URL"` in a command. In Bash/WSL, export the same connection string in your terminal, create the database first, and run migrations using a role that owns the schema:

```bash
export DB_URL='postgresql://username:password@localhost:5432/briefed?sslmode=disable'
goose postgres "$DB_URL" -dir ./sql/schema up
goose postgres "$DB_URL" -dir ./sql/seeds -table goose_seed_version up
sqlc generate
go tool templ generate
tailwindcss -i ./views/css/styles.css -o ./public/styles.css --minify
go run .
```

Keep credentials out of Git. Use your provider’s connection string and TLS settings for a hosted database; `sslmode=disable` above is only a local-development example.

The seed has its own Goose migration directory and version table, separate from schema history. This prevents the seed’s version `001` from colliding with schema version `001`. Its SQL uses upserts; Goose tracks the applied seed and skips it on subsequent `up` calls. To change the catalog later, add a new seed migration rather than assuming an already-recorded file will rerun. Edit `sql/schema` and `sql/queries`, then regenerate; never manually edit `internal/database`.

Run the watchers in separate terminals. They generate files; `go run .` does not automatically restart when Go source changes. The existing `make air` target is available if Air is installed.

```bash
make templ       # watch templ generation
make tailwind    # watch CSS generation
make test
make build
```

## API surface

| Method | Path | Auth | Description |
|---|---|---:|---|
| GET | `/v1/health` | No | Health check |
| GET | `/v1/categories` | No | List categories |
| GET | `/v1/feeds` | No | Discover shared feeds |
| GET | `/v1/feeds/{feedID}` | No | Get a feed |
| POST | `/v1/feeds` | Yes | Validate, publish, and follow a feed |
| POST | `/v1/auth/register` | No | Register with onboarding choices |
| POST | `/v1/auth/login` | No | Create a session |
| POST | `/v1/auth/logout` | No | Delete the current session |
| GET | `/v1/me` | Yes | Read profile and preferences |
| PUT | `/v1/me/preferences` | Yes | Atomically replace preferences |
| PUT | `/v1/me/categories` | Yes | Replace interests |
| GET | `/v1/me/feeds` | Yes | List followed feeds |
| POST | `/v1/me/feeds/{feedID}` | Yes | Follow idempotently |
| DELETE | `/v1/me/feeds/{feedID}` | Yes | Unfollow |
| GET | `/v1/posts` | Optional | Global or personalized timeline |

The browser uses clean topic URLs such as `/?category=sports`; the JSON API accepts UUID category filters such as `/v1/posts?category_ids=ID1,ID2&q=ai&limit=20&offset=0`. The response contains `posts`, `mode`, `limit`, `offset`, and `has_more`. Each post carries feed identity, categories, and an optional image URL. Filters apply within the selected timeline; a personalized search with no matches does not silently widen to global results.

HTML also exposes `GET /partials/feeds/new` and `POST /partials/feeds` for contribution. Import [the Postman collection](docs/postman_collection.json) or inspect [OpenAPI](docs/openapi.json).

## Testing

Current tests cover deterministic boundaries: pagination and filter parsing, clean URL generation, RSS/Atom parsing, image extraction, date variations, and unsupported XML rejection.

```bash
go test ./...
```

The next valuable layer is PostgreSQL integration coverage for rollback, sessions, shared-feed races, cascades, and timeline modes. Those behaviors belong to the database and should be exercised end-to-end before calling the project production-ready.

## What this demonstrates

Briefed models different intentions as different relationships, lets constraints enforce durable truths, places transactions around product actions, reuses operations across HTML and JSON, and adds infrastructure when evidence creates a need. It is compact but not casual: current decisions are explainable and future decisions have recognizable triggers.
