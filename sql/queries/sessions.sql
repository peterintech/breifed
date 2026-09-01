-- name: CreateSession :one
INSERT INTO sessions (id, user_id, token, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, token, created_at, expires_at;

-- name: GetUserBySessionToken :one
SELECT u.id, u.created_at, u.updated_at, u.name, u.email, u.password_hash
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token = $1
AND s.expires_at > NOW();

-- name: DeleteSessionByToken :exec
DELETE FROM sessions
WHERE token = $1;
