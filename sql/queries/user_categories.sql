-- name: CreateUserCategory :exec
INSERT INTO user_categories (user_id, category_id)
VALUES ($1, $2)
ON CONFLICT (user_id, category_id) DO NOTHING;

-- name: DeleteUserCategories :exec
DELETE FROM user_categories
WHERE user_id = $1;

-- name: GetCategoriesForUser :many
SELECT c.id, c.name, c.slug, c.created_at
FROM categories c
JOIN user_categories uc ON uc.category_id = c.id
WHERE uc.user_id = $1
ORDER BY c.name ASC;
