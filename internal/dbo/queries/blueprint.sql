-- name: CreateBlueprint :one
INSERT INTO blueprint (note)
VALUES ('') -- dumb workaround to since SQLC doesnt support DEFAULT VALUES with RETURNING
RETURNING *;

-- name: GetBlueprint :one
SELECT *
FROM blueprint
WHERE id = ?;

-- name: ListBlueprints :many
SELECT *
FROM blueprint
ORDER BY id DESC;

-- name: DeleteBlueprint :exec
DELETE
FROM blueprint
WHERE id = ?;

-- name: UpdateBlueprint :exec
UPDATE blueprint
SET note       = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: CreateBlueprintStep :one
INSERT INTO blueprint_step (blueprint_id, content)
VALUES (?, ?)
RETURNING *;

-- name: ListBlueprintSteps :many
SELECT *
FROM blueprint_step
WHERE blueprint_id = ?
ORDER BY id;


-- name: ListBlueprintPreviousSteps :many
SELECT *
FROM blueprint_step
WHERE blueprint_id = ? AND id < ? -- FIXME: this is temporary while order is not yet configurable
ORDER BY id;


-- name: GetBlueprintStep :one
SELECT *
FROM blueprint_step
WHERE id = ?;

-- name: DeleteBlueprintStep :exec
DELETE
FROM blueprint_step
WHERE id = ?;

-- name: UpdateBlueprintStep :exec
UPDATE blueprint_step
SET content    = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: BlueprintLinkContext :exec
INSERT INTO blueprint_linked_context (blueprint_id, context_id)
VALUES (?, ?);

-- name: BlueprintUnlinkContext :exec
DELETE
FROM blueprint_linked_context
WHERE blueprint_id = ?
  AND context_id = ?;

-- name: ListBlueprintLinkedContexts :many
SELECT context.*
FROM context
         INNER JOIN blueprint_linked_context blc ON blc.context_id = context.id
WHERE blc.blueprint_id = ?
ORDER BY blc.id;

-- name: ListBlueprintUnlinkedContexts :many
SELECT context.*
FROM context
WHERE context.id NOT IN (SELECT context_id
                         FROM blueprint_linked_context blc
                         WHERE blc.blueprint_id = ?)
ORDER BY id;


-- name: BlueprintLinkPage :exec
INSERT INTO blueprint_linked_page (blueprint_id, page_id, inline)
VALUES (?, ?, ?);

-- name: BlueprintUnlinkPage :exec
DELETE
FROM blueprint_linked_page
WHERE blueprint_id = ?
  AND page_id = ?;

-- name: ListBlueprintPages :many
SELECT sqlc.embed(page), blp.inline
FROM page
         LEFT JOIN blueprint_linked_page blp ON page.id = blp.page_id AND blp.blueprint_id = ?
ORDER BY page.num DESC;


-- name: SetBlueprintLinkedPage :exec
INSERT INTO blueprint_linked_page (blueprint_id, page_id, inline)
VALUES (?, ?, ?)
ON CONFLICT DO UPDATE SET inline = excluded.inline;

-- name: ListBlueprintLinkedPages :many
SELECT sqlc.embed(page), blp.inline
FROM page
         INNER JOIN blueprint_linked_page blp ON blp.page_id = page.id
WHERE blp.blueprint_id = ?
ORDER BY page.num DESC;

-- name: ListBlueprintChats :many
SELECT chat.*
FROM chat
         INNER JOIN blueprint_chat bc ON chat.id = bc.chat_id
WHERE bc.blueprint_id = ?
ORDER BY bc.id;


-- name: LinkBlueprintChat :exec
INSERT INTO blueprint_chat (blueprint_id, chat_id)
VALUES (?, ?);