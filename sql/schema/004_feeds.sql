-- +goose Up
CREATE TABLE feeds (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    submitted_by UUID REFERENCES users (id) ON DELETE SET NULL,
    last_fetched_at TIMESTAMPTZ
);

CREATE INDEX feeds_last_fetched_at_idx
ON feeds (last_fetched_at ASC NULLS FIRST);

-- +goose Down
DROP TABLE feeds;
