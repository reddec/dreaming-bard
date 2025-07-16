-- name: CreateContext :one
INSERT INTO context (title, category, content)
VALUES (?, ?, ?)
RETURNING *;


-- name: GetContext :one
SELECT *
FROM context
WHERE id = ?;

-- name: ListContextsByIDs :many
SELECT *
FROM context
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteContext :exec
DELETE
FROM context
WHERE id = ?;

-- name: UpdateContext :exec
UPDATE context
SET title      = ?,
    category   = ?,
    content    = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateContextArchivedStatus :exec
UPDATE context
SET archived = ?
WHERE id = ?;

-- name: ListContexts :many
SELECT *
FROM context
ORDER BY id;

-- name: ListContextsByCategory :many
SELECT *
FROM context
WHERE category = ?
ORDER BY id;

-- name: ListContextCategories :many
SELECT DISTINCT category
FROM context;
