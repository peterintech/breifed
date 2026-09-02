-- name: CreatePost :exec
INSERT INTO posts (id, created_at, updated_at, title, description, published_at, url, feed_id, image_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (url) DO NOTHING;

-- name: GetGlobalPosts :many
SELECT
    p.id,
    p.created_at,
    p.updated_at,
    p.title,
    p.description,
    p.published_at,
    p.url,
    p.feed_id,
    p.image_url,
    f.name AS feed_name,
    f.url AS feed_url
FROM posts p
JOIN feeds f ON f.id = p.feed_id
WHERE (
    sqlc.arg(category_ids)::text = ''
    OR EXISTS (
        SELECT 1
        FROM feed_categories fc
        WHERE fc.feed_id = p.feed_id
          AND fc.category_id = ANY(string_to_array(sqlc.arg(category_ids)::text, ',')::uuid[])
    )
)
AND (
    sqlc.arg(search)::text = ''
    OR p.title ILIKE '%' || sqlc.arg(search)::text || '%'
    OR COALESCE(p.description, '') ILIKE '%' || sqlc.arg(search)::text || '%'
    OR f.name ILIKE '%' || sqlc.arg(search)::text || '%'
)
ORDER BY p.published_at DESC
LIMIT sqlc.arg(result_limit) OFFSET sqlc.arg(result_offset);

-- name: GetPostsForUser :many
SELECT
    p.id,
    p.created_at,
    p.updated_at,
    p.title,
    p.description,
    p.published_at,
    p.url,
    p.feed_id,
    p.image_url,
    f.name AS feed_name,
    f.url AS feed_url
FROM posts p
JOIN feed_follows ff ON ff.feed_id = p.feed_id
JOIN feeds f ON f.id = p.feed_id
WHERE ff.user_id = sqlc.arg(user_id)
AND (
    sqlc.arg(category_ids)::text = ''
    OR EXISTS (
        SELECT 1
        FROM feed_categories fc
        WHERE fc.feed_id = p.feed_id
          AND fc.category_id = ANY(string_to_array(sqlc.arg(category_ids)::text, ',')::uuid[])
    )
)
AND (
    sqlc.arg(search)::text = ''
    OR p.title ILIKE '%' || sqlc.arg(search)::text || '%'
    OR COALESCE(p.description, '') ILIKE '%' || sqlc.arg(search)::text || '%'
    OR f.name ILIKE '%' || sqlc.arg(search)::text || '%'
)
ORDER BY p.published_at DESC
LIMIT sqlc.arg(result_limit) OFFSET sqlc.arg(result_offset);

-- name: UserHasFeedFollows :one
SELECT EXISTS (
    SELECT 1
    FROM feed_follows
    WHERE user_id = $1
);

-- name: GetCategoriesForFeeds :many
SELECT
    fc.feed_id,
    c.id,
    c.name,
    c.slug,
    c.created_at
FROM feed_categories fc
JOIN categories c ON c.id = fc.category_id
WHERE fc.feed_id = ANY(string_to_array($1, ',')::uuid[])
ORDER BY fc.feed_id, c.name ASC;
