-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title VARCHAR(255) NOT NULL UNIQUE,
    url VARCHAR(255) NOT NULL UNIQUE,
    description VARCHAR(255),
    published_at TIMESTAMP,
    feed_id UUID,
    CONSTRAINT fk_feed FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
