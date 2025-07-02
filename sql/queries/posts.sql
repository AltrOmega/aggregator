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

-- name: CreatePost :one
INSERT INTO posts (
    id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *; -- is returning necesairy here?

-- name: GetPostsForUser :many
SELECT *        -- rewrite this as a join statement
FROM posts
WHERE feed_id IN (
    SELECT id
    FROM feeds
    WHERE user_id = $1
)
ORDER BY published_at DESC NULLS LAST
LIMIT $2 OFFSET $3;