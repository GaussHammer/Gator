-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: SelectAllFeeds :many
SELECT feeds.name, feeds.url, users.name
FROM feeds
INNER JOIN users ON feeds.user_id = users.id;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS(
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)SELECT inserted_feed_follow.*, feeds.name AS feed_name, users.name AS user_name
FROM inserted_feed_follow
INNER JOIN users ON users.id = inserted_feed_follow.user_id
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id;

-- name: SelectFeedByUrl :one
SELECT id, name
FROM feeds
WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT users.name AS user_name, feeds.name AS feed_name
FROM feed_follows
INNER JOIN users ON feed_follows.user_id = users.id
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE users.name = $1;

-- name: DeleteFeedFollowRecord :exec
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1 
AND feed_follows.feed_id = (SELECT id FROM feeds WHERE url = $2);

-- name: MarkFeedFetched :exec
UPDATE feeds
SET updated_at = $1, last_fetched_at = $1
WHERE id = $2
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at NULLS FIRST
LIMIT 1;