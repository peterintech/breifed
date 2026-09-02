# Briefed Database

Briefed uses PostgreSQL, goose migrations, and sqlc-generated Go code.

The migrations describe a clean database and must be applied in filename order.

## Tables

### `users`

Stores account identity:

- `id UUID` primary key
- `name TEXT`
- unique `email TEXT`
- `password_hash TEXT`
- creation and update timestamps

### `sessions`

Stores login sessions:

- `id UUID` primary key
- `user_id` referencing users with cascade deletion
- unique session `token`
- `created_at` and `expires_at`

### `categories`

Seeded interest categories:

- `id UUID` primary key
- `name`
- unique `slug`
- `created_at`

Categories are intentionally small and have no display ordering or administration fields.

### `feeds`

The global RSS/Atom catalog:

- `id UUID` primary key
- unique `url`
- extracted `name`
- nullable `submitted_by` referencing users with `ON DELETE SET NULL`
- `last_fetched_at`
- creation and update timestamps

`submitted_by` records attribution, not ownership.

### `feed_categories`

Many-to-many relationship between feeds and categories with primary key `(feed_id, category_id)`.

### `user_categories`

Stores user interests with primary key `(user_id, category_id)`.

### `feed_follows`

Stores explicit subscriptions with primary key `(user_id, feed_id)` and a `created_at` timestamp.

### `posts`

Stores fetched articles:

- `id UUID` primary key
- title and nullable description
- `published_at`
- globally unique article `url`
- nullable article preview `image_url`
- `feed_id` referencing feeds with cascade deletion
- creation and update timestamps

Publication timestamps are not unique. Duplicate article URLs are ignored by the insert query.

## Relationships

```text
users 1---* sessions
users *---* categories  through user_categories
users *---* feeds       through feed_follows
users 1---* feeds       through nullable submitted_by attribution
feeds *---* categories  through feed_categories
feeds 1---* posts
```

## Indexes

- `feeds(last_fetched_at ASC NULLS FIRST)` for scraper selection.
- `feed_categories(category_id, feed_id)` for category discovery.
- `posts(feed_id, published_at DESC)` for feed timelines.
- The `feed_follows` primary key begins with `user_id` and supports followed-feed lookup.

## Query behavior

- `GetFeeds` filters by an optional comma-separated category list, optional name search, limit, and offset.
- `GetFeedsForUser` joins feeds through `feed_follows`.
- `GetGlobalPosts` returns the public newest-first timeline.
- `GetPostsForUser` joins posts through `feed_follows`; both timeline queries support category/search filters.
- Timeline queries return feed identity, fetch `limit + 1`, and use `GetCategoriesForFeeds` to hydrate categories in one batch.
- Follow and category inserts use `ON CONFLICT DO NOTHING`.
- Post inserts use `ON CONFLICT (url) DO NOTHING`.
- Registration and preference replacement combine generated queries inside SQL transactions.

## Workflow

Edit only:

- `sql/schema/*.sql` for schema changes.
- `sql/queries/*.sql` for query changes.

Then run:

```bash
sqlc generate
```

Files under `internal/database` are generated and should never be manually edited.
