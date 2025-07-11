-- name: CreateFeedFollow :many

WITH inserted_feed_follows AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT
    inserted_feed_follows.*,
    feeds.name as feed_name,
    users.name as user_name
FROM inserted_feed_follows
INNER JOIN feeds ON feeds.id = inserted_feed_follows.feed_id
INNER JOIN users ON users.id = inserted_feed_follows.user_id;

-- name: ResetFeedFollows :exec

DELETE FROM feed_follows;

-- name: GetFeedFollowsForUser :many
SELECT
    u.name AS user_name,
    f.name AS feed_name,
    f.url AS feed_url,
    f.id as feed_id,
    u.id as user_id
FROM feed_follows AS ff
INNER JOIN feeds AS f ON f.id = ff.feed_id
INNER JOIN users AS u ON u.id = ff.user_id
WHERE u.id = $1;


-- name: GetFeedFollowsForUserByURL :one
SELECT
    u.name AS user_name,
    f.name AS feed_name,
    f.url AS feed_url,
    f.id as feed_id,
    u.id as user_id
FROM feed_follows AS ff
INNER JOIN feeds AS f ON f.id = ff.feed_id
INNER JOIN users AS u ON u.id = ff.user_id
WHERE u.id = $1 AND f.url = $2;

-- name: GetFeedFollowsForUserByName :one
SELECT
    u.name AS user_name,
    f.name AS feed_name,
    f.url AS feed_url,
    f.id as feed_id,
    u.id as user_id
FROM feed_follows AS ff
INNER JOIN feeds AS f ON f.id = ff.feed_id
INNER JOIN users AS u ON u.id = ff.user_id
WHERE u.id = $1 AND f.name = $2;