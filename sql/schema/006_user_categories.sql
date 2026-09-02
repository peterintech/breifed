-- +goose Up
CREATE TABLE user_categories (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, category_id)
);

-- +goose Down
DROP TABLE user_categories;
