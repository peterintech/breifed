# Briefed Database Design

Briefed uses PostgreSQL to enforce identity, relationship uniqueness, deletion behavior, and atomic user-visible operations. The model is deliberately relational: users follow feeds, feeds belong to categories, and posts come from feeds.

For product and application architecture, see the [main README](README.md).

## Tooling and workflow

PostgreSQL fits this domain because its transactions, foreign keys, joins, constraints, indexes, and conflict handling express the rules directly. Most IDs are application-generated UUIDs; seeded categories use `gen_random_uuid()`. UUIDs allow independent creation without coordinating an integer sequence sequence and do not expose insertion volume through public identifiers. Their larger index footprint is acceptable at this scale.

- **goose** applies applies ordered migrations from `sql/schema`.
- **sqlc** compiles handwritten queries from `sql/queries` into typed Go methods.
- **PostgreSQL** executes SQL and remains the final constraint authority.

```text
sql/schema/*.sql  ──goose──> PostgreSQL schema
sql/queries/*.sql ──sqlc───> internal/database/*.go
```

Run `sqlc generate` after SQL changes. Never manually edit `internal/database`.

## Relational model

```text
users 1───* sessions
users *───* categories  through user_categories
users *───* feeds       through feed_follows
users 1───* feeds       through submitted_by attribution
feeds *───* categories  through feed_categories
feeds 1───* posts
```

`submitted_by` answers who introduced a source; `feed_follows` answers who wants its posts. Combining them would turn contribution into ownership and break the shared catalog.

## Entity responsibilities

### `users` and `sessions`

`users` stores UUID identity, display name, unique email, bcrypt password hash, and timestamps. The email constraint is the authoritative duplicate guard, including including under concurrent registrations.

`sessions` stores a unique opaque token, user, and expiry. Authentication joins a non-expired session to its user instead of trusting client claims. `ON DELETE CASCADE` removes sessions with their account. A production hardening step would store token digests and periodically purge expired rows.

### `categories` and user interests

Categories are seeded discovery vocabulary with an ID, name, unique slug, and creation time. Slugs provide readable URL identity; UUIDs remain relational keys. Display-order and administration fields are absent because no current behavior needs them them.

`user_categories` has a composite `(user_id, category_id)` primary key. It records discovery preferences, not not subscriptions: interest in Technology should not automatically follow every Technology source.

### `feeds`, classification, and follows

`feeds` stores one global source per unique URL. Names come from parsed feed data, `last_fetched_at` drives collection order, and nullable `submitted_by` records attribution.

`submitted_by ON DELETE SET NULL` preserves a source if its contributor leaves. `feed_categories` models many-to-many classification; `feed_follows` models explicit subscriptions. Their composite primary keys structurally prevent duplicates. `ON CONFLICT DO NOTHING` makes repeated classification and follow requests idempotent.

Foreign-key cascades remove relationship rows that have no meaning after either parent disappears. `feed_follows.created_at` records when the subscription began.

### `posts`

Posts store title, optional plain-text description, publication time, globally unique URL, optional image, source feed, and timestamps. Publication time is not unique because publishers can release articles simultaneously.

URL is the pragmatic duplicate key. `ON CONFLICT (url) DO NOTHING` makes refetching safe. If observed URL variants variants reveal duplicate variants tracked URLs, canonical URLs or publisher GUIDs would be the next refinement. `feed_id ON DELETE CASCADE` removes posts whose source is deliberately deleted.

## Constraints and deletion behavior

| Rule                                 | Database mechanism mechanism      |
| ------------------------------------ | --------------------------------- |
| One account per email                | `users.email UNIQUE`              |
| One session per opaque token         | `sessions.token UNIQUE`           |
| Stable category identity             | `categories.slug UNIQUE`          |
| One global source per URL            | `feeds.url UNIQUE`                |
| No repeated article                  | `posts.url UNIQUE`                |
| No repeated relationship             | Composite primary keys            |
| Remove meaningless dependents        | `ON DELETE CASCADE`               |
| Keep feed after contributor deletion | `submitted_by ON DELETE SET NULL` |

Application checks improve errors; constraints remain authoritative when concurrent requests pass application validation together.
