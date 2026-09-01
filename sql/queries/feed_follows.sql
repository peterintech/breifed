-- name: CreateFeedFollow :exec
INSERT INTO feed_follows (user_id, feed_id)
VALUES ($1, $2)
ON CONFLICT (user_id, feed_id) DO NOTHING;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;
