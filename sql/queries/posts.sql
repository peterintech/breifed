-- name: CreatePost :exec
INSERT INTO posts (id, created_at, updated_at, title, description, published_at, url, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (url) DO NOTHING;

-- name: GetPostsForUser :many
SELECT
    p.id,
    p.created_at,
    p.updated_at,
    p.title,
    p.description,
    p.published_at,
    p.url,
    p.feed_id
FROM posts p
JOIN feed_follows ff ON ff.feed_id = p.feed_id
WHERE ff.user_id = $1
ORDER BY p.published_at DESC
LIMIT $2 OFFSET $3;
