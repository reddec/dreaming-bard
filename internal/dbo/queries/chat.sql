-- name: CreateChat :one
INSERT INTO chat (role_id, draft, annotation)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateChatDraft :exec
UPDATE chat
SET draft      = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;


-- name: GetChat :one
SELECT *
FROM chat
WHERE id = ?;


-- name: DeleteChat :exec
DELETE
FROM chat
WHERE id = ?;


-- name: ListChats :many
SELECT sqlc.embed(chat),
       r.name                                                 AS role_name,
       (SELECT COUNT(*) FROM message WHERE chat_id = chat.id) AS num_messages
FROM chat
         INNER JOIN role r ON chat.role_id = r.id
ORDER BY chat.id DESC;

-- name: ListLastChats :many
SELECT *
FROM chat
ORDER BY id DESC
LIMIT ?;


-- name: AddChatStats :exec
UPDATE chat
SET input_tokens  = input_tokens + ?,
    output_tokens = output_tokens + ?,
    updated_at    = CURRENT_TIMESTAMP
WHERE id = ?;