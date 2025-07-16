-- name: CreatePrompt :one
INSERT INTO prompt (summary, content, role_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetPrompt :one
SELECT *
FROM prompt
WHERE id = ?;


-- name: ListPrompts :many
SELECT sqlc.embed(prompt), role_id, role.purpose AS role_purpose, role.name AS role_name
FROM prompt
         INNER JOIN role ON role.id = prompt.role_id
ORDER BY prompt.pinned_at IS NULL, prompt.pinned_at, prompt.id;

-- name: ListPinnedPrompts :many
SELECT *
FROM prompt
WHERE pinned_at IS NOT NULL
ORDER BY pinned_at;

-- name: DeletePrompt :exec
DELETE
FROM prompt
WHERE id = ?;


-- name: UpdatePrompt :exec
UPDATE prompt
SET summary    = ?,
    content    = ?,
    role_id    = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdatePromptPin :exec
UPDATE prompt
SET pinned_at = ?
WHERE id = ?;