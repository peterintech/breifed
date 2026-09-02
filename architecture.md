# Briefed Architecture

## Overview

Briefed is a deliberately small Go monolith. Chi serves both JSON APIs and server-rendered HTML while a background scraper refreshes the shared RSS/Atom catalog.

```text
Browser
  ├─ full page ───────> handlers + views
  ├─ HTMX fragments ──> handlers + components
  └─ JSON /v1 ────────> internal/api
                            │
                            v
shared helpers ───────> sqlc queries ───────> PostgreSQL
       ^                                           ^
       └──── feed parser <──── background scraper ┘
```

## Package boundaries

- `views`: full `.templ` pages and the shared document shell.
- `components`: reusable page sections and HTMX fragments.
- `handlers`: HTML routes, rendering, form parsing, and UI redirects.
- `public`: compiled Tailwind CSS, vendored HTMX browser asset, and small interaction JavaScript.
- `types`: view models shared by views/components and HTML handlers.
- `internal/api`: `/v1` JSON routes and response models.
- `internal/accounts`: shared registration/login operations used by JSON and HTML handlers.
- `internal/preferences`: the atomic interest/follow replacement operation.
- `internal/sessionauth`: random token generation, cookie handling, and optional user resolution.
- `internal/timeline`: global/personalized selection and category hydration.
- `internal/feedparser`: shared RSS/Atom fetch and parsing.
- `internal/feedcatalog`: shared global feed contribution and automatic-follow transaction.
- `internal/scrapper`: least-recently-fetched collection loop.
- `internal/database`: sqlc-generated only; never manually edit.
- `sql/schema`, `sql/queries`, `sql/seeds`: migrations, handwritten sqlc queries, and standalone catalog seed.

## Timeline selection

```text
GET / or GET /v1/posts
       │
       ├─ no/invalid session ──────────────> global
       ├─ valid session + follows ─────────> personalized
       └─ valid session + zero follows ────> global_fallback
```

All modes apply the same optional category/search filters and `limit`/`offset` pagination. Queries fetch `limit + 1` rows to calculate `has_more`. Feed categories are batch-loaded for the returned feed IDs.

## HTML and HTMX flow

`GET /` renders the shell, category navigation, and initial timeline. HTMX replaces only the timeline for search/category changes and appends story rows for load-more. URL filters remain shareable through `HX-Push-Url`.

Anonymous onboarding uses a native `<dialog>` with three server-rendered steps. The final form calls the same registration transaction as `/v1/auth/register`. Authenticated personalization is a fixed overlay drawer; the page rails never move. The drawer's single save calls the same atomic preference operation as `/v1/me/preferences`.

Feed contribution uses the same overlay-drawer vocabulary. `GET /partials/feeds/new` renders the form and `POST /partials/feeds` calls `internal/feedcatalog`. The shared package is also used by `POST /v1/feeds`, so HTML and JSON clients receive the same parsing, duplicate handling, transaction, and automatic-follow behavior.

## Transactions

- Registration creates the user, interests, follows, and session together.
- New feed contribution creates the global feed, category relationships, and submitter follow together; an existing feed is simply followed.
- Preference save deletes/recreates interests and follows together.

These transactions prevent partially applied user-visible state without adding a repository/service framework.

## Scraper and parser

The scraper selects global feeds ordered by `last_fetched_at ASC NULLS FIRST`, fetches small concurrent batches, parses RSS or Atom through `internal/feedparser`, and marks each fetch attempt. Article URL conflicts are ignored. Descriptions become plain-text excerpts; images are selected from media elements, image enclosures, or the first HTML image and may remain null.

This version intentionally omits distributed leases, conditional requests, exponential backoff, and multi-instance coordination.

## Authentication

Passwords use bcrypt. A successful register/login stores a random 30-day token in `sessions` and sends it as an HTTP-only, SameSite=Lax cookie. Protected endpoints require it. Public posts treat missing, invalid, and expired cookies as anonymous, while unexpected database errors still return `500`.
