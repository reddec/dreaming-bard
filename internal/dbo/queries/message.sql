-- name: CreateMessage :one
INSERT INTO message (chat_id, content, role, tool_id, tool_name)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMessage :one
SELECT *
FROM message
WHERE id = ?;

-- name: DeleteMessage :exec
DELETE
FROM message
WHERE id = ?;

-- name: UpdateMessageContent :exec
UPDATE message
SET content    = ?
WHERE id = ?;

-- name: ListMessagesByChat :many
SELECT *
FROM message
WHERE chat_id = ?
ORDER BY id;
