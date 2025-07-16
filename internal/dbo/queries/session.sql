-- name: SetSession :exec
INSERT INTO session (token, content, expires_at)
VALUES (?, ?, ?)
ON CONFLICT DO UPDATE SET content    = excluded.content,
                          expires_at = excluded.expires_at;


-- name: DeleteSession :exec
DELETE
FROM session
WHERE token = ?;

-- name: DeleteSessionExpired :exec
DELETE
FROM session
WHERE expires_at < CURRENT_TIMESTAMP;

-- name: FindSession :one
SELECT *
FROM session
WHERE token = ?;