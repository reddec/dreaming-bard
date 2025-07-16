-- name: GetPreference :one
SELECT *
FROM preference
WHERE name = ?;


-- name: SetPreference :exec
INSERT INTO preference(name, value)
VALUES (?, ?)
ON CONFLICT DO UPDATE SET value      = EXCLUDED.value,
                          updated_at = CURRENT_TIMESTAMP;


-- name: DeletePreference :exec
DELETE
FROM preference
WHERE name = ?;