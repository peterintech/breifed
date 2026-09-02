-- +goose Up
ALTER TABLE posts
ADD COLUMN image_url TEXT;

-- +goose Down
ALTER TABLE posts
DROP COLUMN image_url;
