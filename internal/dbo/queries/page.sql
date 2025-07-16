-- name: CreatePage :one
INSERT INTO page (summary, content, num)
VALUES (?, ?, (SELECT COALESCE(MAX(num), 0) + 1 FROM page))
RETURNING *;

-- name: GetPage :one
SELECT *
FROM page
WHERE id = ?;

-- name: GetPageNum :one
SELECT num
FROM page
WHERE id = ?;

-- name: GetPageByNum :one
SELECT *
FROM page
WHERE num = ?
LIMIT 1;

-- name: UpdatePage :exec
UPDATE page
SET summary    = ?,
    content    = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdatePageSummary :exec
UPDATE page
SET summary    = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;


-- name: deletePage :exec
DELETE
FROM page
WHERE id = ?;

-- name: ListPages :many
SELECT *
FROM page
ORDER BY num;

-- name: ListPagesByIDs :many
SELECT *
FROM page
WHERE id IN (sqlc.slice('ids'))
ORDER BY num;


-- name: ListLastPages :many
SELECT *
FROM page
ORDER BY num DESC
LIMIT ?;

-- name: ListPagesIDs :many
SELECT id
FROM page
ORDER BY num;

-- name: movePages :exec
UPDATE page
SET num = num + (CASE WHEN num >= ? THEN 1 ELSE 0 END);

-- name: compressPagesSequence :exec
UPDATE page
SET num = (SELECT rn
           FROM (SELECT id,
                        ROW_NUMBER() OVER (ORDER BY num, id) AS rn
                 FROM page) AS seq
           WHERE seq.id = page.id);
-- all this ugly thing because of limitations of SQLc

-- name: setPageNum :exec
UPDATE page
SET num = ?
WHERE id = ?;