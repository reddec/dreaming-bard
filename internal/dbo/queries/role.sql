-- name: CreateRole :one
INSERT INTO role (name, system, model, purpose)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetRole :one
SELECT *
FROM role
WHERE id = ?;


-- name: DeleteRole :exec
DELETE
FROM role
WHERE id = ?;


-- name: ListRoles :many
SELECT *
FROM role
ORDER BY id;


-- name: UpdateRole :one
UPDATE role
SET name       = ?,
    system     = ?,
    model      = ?,
    purpose    = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;