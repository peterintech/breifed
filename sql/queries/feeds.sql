-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, submitted_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at, name, url, submitted_by, last_fetched_at;

-- name: GetFeeds :many
SELECT DISTINCT
    f.id,
    f.created_at,
    f.updated_at,
    f.name,
    f.url,
    f.submitted_by,
    f.last_fetched_at
FROM feeds f
LEFT JOIN feed_categories fc ON fc.feed_id = f.id
WHERE (
    sqlc.arg(category_ids)::text = ''
    OR fc.category_id = ANY(string_to_array(sqlc.arg(category_ids)::text, ',')::uuid[])
)
AND (
    sqlc.arg(search)::text = ''
    OR f.name ILIKE '%' || sqlc.arg(search)::text || '%'
)
ORDER BY f.name ASC
LIMIT sqlc.arg(result_limit) OFFSET sqlc.arg(result_offset);

-- name: GetFeedByID :one
SELECT id, created_at, updated_at, name, url, submitted_by, last_fetched_at
FROM feeds
WHERE id = $1;

-- name: GetFeedByURL :one
SELECT id, created_at, updated_at, name, url, submitted_by, last_fetched_at
FROM feeds
WHERE url = $1;

-- name: GetFeedsForUser :many
SELECT
    f.id,
    f.created_at,
    f.updated_at,
    f.name,
    f.url,
    f.submitted_by,
    f.last_fetched_at
FROM feeds f
JOIN feed_follows ff ON ff.feed_id = f.id
WHERE ff.user_id = $1
ORDER BY f.name ASC;

-- name: GetNextFeedsToFetch :many
SELECT id, created_at, updated_at, name, url, submitted_by, last_fetched_at
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT $1;

-- name: MarkFeedAsFetched :one
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING id, created_at, updated_at, name, url, submitted_by, last_fetched_at;
