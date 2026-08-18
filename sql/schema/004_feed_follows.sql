-- +goose Up
CREATE TABLE
    feed_follows (
        id UUID PRIMARY KEY,
        created_at TIMESTAMP NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW (),
        feed_id UUID NOT NULL REFERENCES feeds (id) ON DELETE CASCADE,
        user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        UNIQUE (feed_id, user_id)
    );

-- +goose Down