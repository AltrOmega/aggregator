-- +goose Up
CREATE TABLE feed_follows (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID,
    feed_id UUID,
    CONSTRAINT unique_user_feed_pair UNIQUE (user_id, feed_id)
);  -- missing auto delete on user of feed delete

-- +goose Down
DROP TABLE feed_follows;