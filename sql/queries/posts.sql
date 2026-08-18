-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, description, published_at, url, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, updated_at, title, description, published_at, url, feed_id;

-- name: GetPostsForUser :many
SELECT p.id, p.created_at, p.updated_at, p.title, p.description, p.published_at, p.url, p.feed_id
FROM posts p
JOIN feeds f ON p.feed_id = f.id
WHERE f.user_id = $1
ORDER BY p.published_at DESC
LIMIT $2 OFFSET $3;