-- name: CreateFeedCategory :exec
INSERT INTO feed_categories (feed_id, category_id)
VALUES ($1, $2)
ON CONFLICT (feed_id, category_id) DO NOTHING;

-- name: GetCategoriesForFeed :many
SELECT c.id, c.name, c.slug, c.created_at
FROM categories c
JOIN feed_categories fc ON fc.category_id = c.id
WHERE fc.feed_id = $1
ORDER BY c.name ASC;
