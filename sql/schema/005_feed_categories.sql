-- +goose Up
CREATE TABLE feed_categories (
    feed_id UUID NOT NULL REFERENCES feeds (id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    PRIMARY KEY (feed_id, category_id)
);

CREATE INDEX feed_categories_category_id_feed_id_idx
ON feed_categories (category_id, feed_id);

-- +goose Down
DROP TABLE feed_categories;
