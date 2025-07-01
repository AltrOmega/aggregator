-- name: CreateFeeds :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $2, updated_at = $2
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST;

-- name: GetFeedByName :one

SELECT * FROM feeds
WHERE name = $1;

-- name: GetFeedByURL :one

SELECT * FROM feeds
WHERE url = $1;

-- name: ResetFeeds :exec

DELETE FROM feeds;

-- name: DeleteFeedById :exec

DELETE FROM feeds
WHERE id = $1;

-- name: GetFeedsRaw :many

SELECT * FROM feeds;

-- name: GetFeeds :many

SELECT f.name AS feed_name, f.url AS feed_URL, u.name AS user_name
FROM feeds AS f
JOIN users AS u ON f.user_id = u.id;