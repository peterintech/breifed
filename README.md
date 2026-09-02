# Briefed

Briefed is a small RSS/Atom aggregator with a public editorial homepage and a personalized newest-first timeline for signed-in readers. Users choose interests, discover sources in those categories, and explicitly follow the feeds they want. Every feed is global: a source contributed by one user can be followed by everyone.

## Stack

- Go and Chi
- PostgreSQL, goose, and sqlc
- templ-rendered HTML and HTMX interactions
- Tailwind CSS v4 compiled to a static stylesheet
- Cookie sessions with bcrypt passwords

The frontend follows a simple separation inspired by `fullstack-go-htmx`: full pages live in `views`, reusable fragments live in `components`, HTML handlers live in `handlers`, and static assets live in `public`. JSON endpoints remain in `internal/api`.

## Setup

Requirements: Go 1.26+, PostgreSQL, goose, sqlc, Node.js, and Corepack.

Create `.env`:

```env
PORT=8080
DB_URL=postgresql://username:password@localhost:5432/briefed?sslmode=disable
```

Then run:

```bash
goose -dir sql/schema postgres "$DB_URL" up
sqlc generate
corepack pnpm install
corepack pnpm run build
go tool templ generate
go run .
```

Seed the global catalog after migrations:

```bash
psql "$DB_URL" -f sql/seeds/001_catalog.sql
```

Do not pass Goose migration files directly to `psql`: they contain both Up and Down statements. Use Goose for `sql/schema` and `psql` only for the standalone seed.

## Browser experience

- `/` renders public stories immediately. A valid session personalizes the same page; a signed-in user with no follows receives the global fallback.
- Topic filters, debounced search, and load-more pagination update the timeline through HTMX.
- Anonymous visitors can open the three-step interest → source → account modal. It also appears after 15 seconds or 30% scroll unless dismissed for the browser session.
- Signed-in readers use the right-side preferences drawer. Save atomically replaces interests and followed sources; Cancel, Escape, the backdrop, or Close discards changes.
- `/login` renders the cookie-session login flow.

Build commands:

```bash
make templ          # generate *_templ.go
make tailwind       # watch Tailwind
make test
make build
```

## JSON API flow

1. `GET /v1/categories`
2. `GET /v1/feeds?category_ids=...`
3. `POST /v1/auth/register` with selected category/feed IDs
4. Browser stores the HTTP-only `briefed_session` cookie
5. `GET /v1/posts` returns global, personalized, or global-fallback results
6. Follow existing feeds, update preferences, or contribute a shared RSS/Atom URL

`GET /v1/posts?category_ids=ID1,ID2&q=ai&limit=20&offset=0` is public. If a valid session cookie is present, personalization applies automatically. Its response includes `posts`, `mode`, `limit`, `offset`, and `has_more`; each post carries feed identity, categories, and an optional `image_url`.

## Endpoints

| Method | Path | Auth | Description |
|---|---|---:|---|
| GET | `/v1/health` | No | Health check |
| GET | `/v1/categories` | No | List seeded categories |
| GET | `/v1/feeds` | No | Search/filter shared feeds |
| GET | `/v1/feeds/{feedID}` | No | Get a shared feed |
| POST | `/v1/feeds` | Yes | Validate, publish, and follow a feed |
| POST | `/v1/auth/register` | No | Register with onboarding choices |
| POST | `/v1/auth/login` | No | Create a session |
| POST | `/v1/auth/logout` | No | Delete the current session if present |
| GET | `/v1/me` | Yes | Get profile, interests, and follows |
| PUT | `/v1/me/preferences` | Yes | Atomically replace interests and follows |
| PUT | `/v1/me/categories` | Yes | Replace interests |
| GET | `/v1/me/feeds` | Yes | List followed feeds |
| POST | `/v1/me/feeds/{feedID}` | Yes | Follow a feed |
| DELETE | `/v1/me/feeds/{feedID}` | Yes | Unfollow a feed |
| GET | `/v1/posts` | Optional | Public/personalized timeline |

Import [docs/postman_collection.json](docs/postman_collection.json) directly into Postman. The complete contract is in [docs/openapi.json](docs/openapi.json).

## Development notes

Write schema changes under `sql/schema` and queries under `sql/queries`. Never manually edit `internal/database`; run `sqlc generate` after changing either SQL surface. Categories and starter feeds are seeded separately and there is no category administration UI in this version.
