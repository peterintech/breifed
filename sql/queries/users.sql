-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, email, password_hash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at, name, email, password_hash;

-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, name, email, password_hash
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, created_at, updated_at, name, email, password_hash
FROM users
WHERE id = $1;
