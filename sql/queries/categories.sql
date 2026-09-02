-- name: GetCategories :many
SELECT id, name, slug, created_at
FROM categories
ORDER BY name ASC;
