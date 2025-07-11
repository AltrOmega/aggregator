// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: feed_follows.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createFeedFollow = `-- name: CreateFeedFollow :many

WITH inserted_feed_follows AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING id, created_at, updated_at, user_id, feed_id
)
SELECT
    inserted_feed_follows.id, inserted_feed_follows.created_at, inserted_feed_follows.updated_at, inserted_feed_follows.user_id, inserted_feed_follows.feed_id,
    feeds.name as feed_name,
    users.name as user_name
FROM inserted_feed_follows
INNER JOIN feeds ON feeds.id = inserted_feed_follows.feed_id
INNER JOIN users ON users.id = inserted_feed_follows.user_id
`

type CreateFeedFollowParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.NullUUID
	FeedID    uuid.NullUUID
}

type CreateFeedFollowRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.NullUUID
	FeedID    uuid.NullUUID
	FeedName  string
	UserName  string
}

func (q *Queries) CreateFeedFollow(ctx context.Context, arg CreateFeedFollowParams) ([]CreateFeedFollowRow, error) {
	rows, err := q.db.QueryContext(ctx, createFeedFollow,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.FeedID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CreateFeedFollowRow
	for rows.Next() {
		var i CreateFeedFollowRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.UserID,
			&i.FeedID,
			&i.FeedName,
			&i.UserName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getFeedFollowsForUser = `-- name: GetFeedFollowsForUser :many
SELECT
    u.name AS user_name,
    f.name AS feed_name,
    f.url AS feed_url,
    f.id as feed_id,
    u.id as user_id
FROM feed_follows AS ff
INNER JOIN feeds AS f ON f.id = ff.feed_id
INNER JOIN users AS u ON u.id = ff.user_id
WHERE u.id = $1
`

type GetFeedFollowsForUserRow struct {
	UserName string
	FeedName string
	FeedUrl  string
	FeedID   uuid.UUID
	UserID   uuid.UUID
}

func (q *Queries) GetFeedFollowsForUser(ctx context.Context, id uuid.UUID) ([]GetFeedFollowsForUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedFollowsForUser, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedFollowsForUserRow
	for rows.Next() {
		var i GetFeedFollowsForUserRow
		if err := rows.Scan(
			&i.UserName,
			&i.FeedName,
			&i.FeedUrl,
			&i.FeedID,
			&i.UserID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getFeedFollowsForUserByName = `-- name: GetFeedFollowsForUserByName :one
SELECT
    u.name AS user_name,
    f.name AS feed_name,
    f.url AS feed_url,
    f.id as feed_id,
    u.id as user_id
FROM feed_follows AS ff
INNER JOIN feeds AS f ON f.id = ff.feed_id
INNER JOIN users AS u ON u.id = ff.user_id
WHERE u.id = $1 AND f.name = $2
`

type GetFeedFollowsForUserByNameParams struct {
	ID   uuid.UUID
	Name string
}

type GetFeedFollowsForUserByNameRow struct {
	UserName string
	FeedName string
	FeedUrl  string
	FeedID   uuid.UUID
	UserID   uuid.UUID
}

func (q *Queries) GetFeedFollowsForUserByName(ctx context.Context, arg GetFeedFollowsForUserByNameParams) (GetFeedFollowsForUserByNameRow, error) {
	row := q.db.QueryRowContext(ctx, getFeedFollowsForUserByName, arg.ID, arg.Name)
	var i GetFeedFollowsForUserByNameRow
	err := row.Scan(
		&i.UserName,
		&i.FeedName,
		&i.FeedUrl,
		&i.FeedID,
		&i.UserID,
	)
	return i, err
}

const getFeedFollowsForUserByURL = `-- name: GetFeedFollowsForUserByURL :one
SELECT
    u.name AS user_name,
    f.name AS feed_name,
    f.url AS feed_url,
    f.id as feed_id,
    u.id as user_id
FROM feed_follows AS ff
INNER JOIN feeds AS f ON f.id = ff.feed_id
INNER JOIN users AS u ON u.id = ff.user_id
WHERE u.id = $1 AND f.url = $2
`

type GetFeedFollowsForUserByURLParams struct {
	ID  uuid.UUID
	Url string
}

type GetFeedFollowsForUserByURLRow struct {
	UserName string
	FeedName string
	FeedUrl  string
	FeedID   uuid.UUID
	UserID   uuid.UUID
}

func (q *Queries) GetFeedFollowsForUserByURL(ctx context.Context, arg GetFeedFollowsForUserByURLParams) (GetFeedFollowsForUserByURLRow, error) {
	row := q.db.QueryRowContext(ctx, getFeedFollowsForUserByURL, arg.ID, arg.Url)
	var i GetFeedFollowsForUserByURLRow
	err := row.Scan(
		&i.UserName,
		&i.FeedName,
		&i.FeedUrl,
		&i.FeedID,
		&i.UserID,
	)
	return i, err
}

const resetFeedFollows = `-- name: ResetFeedFollows :exec

DELETE FROM feed_follows
`

func (q *Queries) ResetFeedFollows(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, resetFeedFollows)
	return err
}
