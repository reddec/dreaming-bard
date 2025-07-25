// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: blueprint.sql

package dbo

import (
	"context"
)

const blueprintLinkContext = `-- name: BlueprintLinkContext :exec
INSERT INTO blueprint_linked_context (blueprint_id, context_id)
VALUES (?, ?)
`

type BlueprintLinkContextParams struct {
	BlueprintID int64 `json:"blueprint_id"`
	ContextID   int64 `json:"context_id"`
}

func (q *Queries) BlueprintLinkContext(ctx context.Context, arg BlueprintLinkContextParams) error {
	_, err := q.db.ExecContext(ctx, blueprintLinkContext, arg.BlueprintID, arg.ContextID)
	return err
}

const blueprintLinkPage = `-- name: BlueprintLinkPage :exec
INSERT INTO blueprint_linked_page (blueprint_id, page_id, inline)
VALUES (?, ?, ?)
`

type BlueprintLinkPageParams struct {
	BlueprintID int64 `json:"blueprint_id"`
	PageID      int64 `json:"page_id"`
	Inline      bool  `json:"inline"`
}

func (q *Queries) BlueprintLinkPage(ctx context.Context, arg BlueprintLinkPageParams) error {
	_, err := q.db.ExecContext(ctx, blueprintLinkPage, arg.BlueprintID, arg.PageID, arg.Inline)
	return err
}

const blueprintUnlinkContext = `-- name: BlueprintUnlinkContext :exec
DELETE
FROM blueprint_linked_context
WHERE blueprint_id = ?
  AND context_id = ?
`

type BlueprintUnlinkContextParams struct {
	BlueprintID int64 `json:"blueprint_id"`
	ContextID   int64 `json:"context_id"`
}

func (q *Queries) BlueprintUnlinkContext(ctx context.Context, arg BlueprintUnlinkContextParams) error {
	_, err := q.db.ExecContext(ctx, blueprintUnlinkContext, arg.BlueprintID, arg.ContextID)
	return err
}

const blueprintUnlinkPage = `-- name: BlueprintUnlinkPage :exec
DELETE
FROM blueprint_linked_page
WHERE blueprint_id = ?
  AND page_id = ?
`

type BlueprintUnlinkPageParams struct {
	BlueprintID int64 `json:"blueprint_id"`
	PageID      int64 `json:"page_id"`
}

func (q *Queries) BlueprintUnlinkPage(ctx context.Context, arg BlueprintUnlinkPageParams) error {
	_, err := q.db.ExecContext(ctx, blueprintUnlinkPage, arg.BlueprintID, arg.PageID)
	return err
}

const createBlueprint = `-- name: CreateBlueprint :one
INSERT INTO blueprint (note)
VALUES ('') -- dumb workaround to since SQLC doesnt support DEFAULT VALUES with RETURNING
RETURNING id, created_at, updated_at, note
`

func (q *Queries) CreateBlueprint(ctx context.Context) (Blueprint, error) {
	row := q.db.QueryRowContext(ctx, createBlueprint)
	var i Blueprint
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Note,
	)
	return i, err
}

const createBlueprintStep = `-- name: CreateBlueprintStep :one
INSERT INTO blueprint_step (blueprint_id, content)
VALUES (?, ?)
RETURNING id, created_at, updated_at, blueprint_id, content
`

type CreateBlueprintStepParams struct {
	BlueprintID int64  `json:"blueprint_id"`
	Content     string `json:"content"`
}

func (q *Queries) CreateBlueprintStep(ctx context.Context, arg CreateBlueprintStepParams) (BlueprintStep, error) {
	row := q.db.QueryRowContext(ctx, createBlueprintStep, arg.BlueprintID, arg.Content)
	var i BlueprintStep
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.BlueprintID,
		&i.Content,
	)
	return i, err
}

const deleteBlueprint = `-- name: DeleteBlueprint :exec
DELETE
FROM blueprint
WHERE id = ?
`

func (q *Queries) DeleteBlueprint(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteBlueprint, id)
	return err
}

const deleteBlueprintStep = `-- name: DeleteBlueprintStep :exec
DELETE
FROM blueprint_step
WHERE id = ?
`

func (q *Queries) DeleteBlueprintStep(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteBlueprintStep, id)
	return err
}

const getBlueprint = `-- name: GetBlueprint :one
SELECT id, created_at, updated_at, note
FROM blueprint
WHERE id = ?
`

func (q *Queries) GetBlueprint(ctx context.Context, id int64) (Blueprint, error) {
	row := q.db.QueryRowContext(ctx, getBlueprint, id)
	var i Blueprint
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Note,
	)
	return i, err
}

const getBlueprintStep = `-- name: GetBlueprintStep :one
SELECT id, created_at, updated_at, blueprint_id, content
FROM blueprint_step
WHERE id = ?
`

func (q *Queries) GetBlueprintStep(ctx context.Context, id int64) (BlueprintStep, error) {
	row := q.db.QueryRowContext(ctx, getBlueprintStep, id)
	var i BlueprintStep
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.BlueprintID,
		&i.Content,
	)
	return i, err
}

const linkBlueprintChat = `-- name: LinkBlueprintChat :exec
INSERT INTO blueprint_chat (blueprint_id, chat_id)
VALUES (?, ?)
`

type LinkBlueprintChatParams struct {
	BlueprintID int64 `json:"blueprint_id"`
	ChatID      int64 `json:"chat_id"`
}

func (q *Queries) LinkBlueprintChat(ctx context.Context, arg LinkBlueprintChatParams) error {
	_, err := q.db.ExecContext(ctx, linkBlueprintChat, arg.BlueprintID, arg.ChatID)
	return err
}

const listBlueprintChats = `-- name: ListBlueprintChats :many
SELECT chat.id, chat.created_at, chat.updated_at, chat.input_tokens, chat.output_tokens, chat.draft, chat.role_id, chat.annotation
FROM chat
         INNER JOIN blueprint_chat bc ON chat.id = bc.chat_id
WHERE bc.blueprint_id = ?
ORDER BY bc.id
`

func (q *Queries) ListBlueprintChats(ctx context.Context, blueprintID int64) ([]Chat, error) {
	rows, err := q.db.QueryContext(ctx, listBlueprintChats, blueprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Chat{}
	for rows.Next() {
		var i Chat
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.InputTokens,
			&i.OutputTokens,
			&i.Draft,
			&i.RoleID,
			&i.Annotation,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBlueprintLinkedContexts = `-- name: ListBlueprintLinkedContexts :many
SELECT context.id, context.created_at, context.updated_at, context.title, context.category, context.content, context.archived
FROM context
         INNER JOIN blueprint_linked_context blc ON blc.context_id = context.id
WHERE blc.blueprint_id = ?
ORDER BY blc.id
`

func (q *Queries) ListBlueprintLinkedContexts(ctx context.Context, blueprintID int64) ([]Context, error) {
	rows, err := q.db.QueryContext(ctx, listBlueprintLinkedContexts, blueprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Context{}
	for rows.Next() {
		var i Context
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Title,
			&i.Category,
			&i.Content,
			&i.Archived,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBlueprintLinkedPages = `-- name: ListBlueprintLinkedPages :many
SELECT page.id, page.created_at, page.updated_at, page.summary, page.content, page.num, blp.inline
FROM page
         INNER JOIN blueprint_linked_page blp ON blp.page_id = page.id
WHERE blp.blueprint_id = ?
ORDER BY page.num DESC
`

type ListBlueprintLinkedPagesRow struct {
	Page   Page `json:"page"`
	Inline bool `json:"inline"`
}

func (q *Queries) ListBlueprintLinkedPages(ctx context.Context, blueprintID int64) ([]ListBlueprintLinkedPagesRow, error) {
	rows, err := q.db.QueryContext(ctx, listBlueprintLinkedPages, blueprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListBlueprintLinkedPagesRow{}
	for rows.Next() {
		var i ListBlueprintLinkedPagesRow
		if err := rows.Scan(
			&i.Page.ID,
			&i.Page.CreatedAt,
			&i.Page.UpdatedAt,
			&i.Page.Summary,
			&i.Page.Content,
			&i.Page.Num,
			&i.Inline,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBlueprintPages = `-- name: ListBlueprintPages :many
SELECT page.id, page.created_at, page.updated_at, page.summary, page.content, page.num, blp.inline
FROM page
         LEFT JOIN blueprint_linked_page blp ON page.id = blp.page_id AND blp.blueprint_id = ?
ORDER BY page.num DESC
`

type ListBlueprintPagesRow struct {
	Page   Page  `json:"page"`
	Inline *bool `json:"inline"`
}

func (q *Queries) ListBlueprintPages(ctx context.Context, blueprintID int64) ([]ListBlueprintPagesRow, error) {
	rows, err := q.db.QueryContext(ctx, listBlueprintPages, blueprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListBlueprintPagesRow{}
	for rows.Next() {
		var i ListBlueprintPagesRow
		if err := rows.Scan(
			&i.Page.ID,
			&i.Page.CreatedAt,
			&i.Page.UpdatedAt,
			&i.Page.Summary,
			&i.Page.Content,
			&i.Page.Num,
			&i.Inline,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBlueprintPreviousSteps = `-- name: ListBlueprintPreviousSteps :many
SELECT id, created_at, updated_at, blueprint_id, content
FROM blueprint_step
WHERE blueprint_id = ? AND id < ? -- FIXME: this is temporary while order is not yet configurable
ORDER BY id
`

type ListBlueprintPreviousStepsParams struct {
	BlueprintID int64 `json:"blueprint_id"`
	ID          int64 `json:"id"`
}

func (q *Queries) ListBlueprintPreviousSteps(ctx context.Context, arg ListBlueprintPreviousStepsParams) ([]BlueprintStep, error) {
	rows, err := q.db.QueryContext(ctx, listBlueprintPreviousSteps, arg.BlueprintID, arg.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []BlueprintStep{}
	for rows.Next() {
		var i BlueprintStep
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.BlueprintID,
			&i.Content,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBlueprintSteps = `-- name: ListBlueprintSteps :many
SELECT id, created_at, updated_at, blueprint_id, content
FROM blueprint_step
WHERE blueprint_id = ?
ORDER BY id
`

func (q *Queries) ListBlueprintSteps(ctx context.Context, blueprintID int64) ([]BlueprintStep, error) {
	rows, err := q.db.QueryContext(ctx, listBlueprintSteps, blueprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []BlueprintStep{}
	for rows.Next() {
		var i BlueprintStep
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.BlueprintID,
			&i.Content,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBlueprintUnlinkedContexts = `-- name: ListBlueprintUnlinkedContexts :many
SELECT context.id, context.created_at, context.updated_at, context.title, context.category, context.content, context.archived
FROM context
WHERE context.id NOT IN (SELECT context_id
                         FROM blueprint_linked_context blc
                         WHERE blc.blueprint_id = ?)
ORDER BY id
`

func (q *Queries) ListBlueprintUnlinkedContexts(ctx context.Context, blueprintID int64) ([]Context, error) {
	rows, err := q.db.QueryContext(ctx, listBlueprintUnlinkedContexts, blueprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Context{}
	for rows.Next() {
		var i Context
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Title,
			&i.Category,
			&i.Content,
			&i.Archived,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBlueprints = `-- name: ListBlueprints :many
SELECT id, created_at, updated_at, note
FROM blueprint
ORDER BY id DESC
`

func (q *Queries) ListBlueprints(ctx context.Context) ([]Blueprint, error) {
	rows, err := q.db.QueryContext(ctx, listBlueprints)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Blueprint{}
	for rows.Next() {
		var i Blueprint
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Note,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setBlueprintLinkedPage = `-- name: SetBlueprintLinkedPage :exec
INSERT INTO blueprint_linked_page (blueprint_id, page_id, inline)
VALUES (?, ?, ?)
ON CONFLICT DO UPDATE SET inline = excluded.inline
`

type SetBlueprintLinkedPageParams struct {
	BlueprintID int64 `json:"blueprint_id"`
	PageID      int64 `json:"page_id"`
	Inline      bool  `json:"inline"`
}

func (q *Queries) SetBlueprintLinkedPage(ctx context.Context, arg SetBlueprintLinkedPageParams) error {
	_, err := q.db.ExecContext(ctx, setBlueprintLinkedPage, arg.BlueprintID, arg.PageID, arg.Inline)
	return err
}

const updateBlueprint = `-- name: UpdateBlueprint :exec
UPDATE blueprint
SET note       = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`

type UpdateBlueprintParams struct {
	Note string `json:"note"`
	ID   int64  `json:"id"`
}

func (q *Queries) UpdateBlueprint(ctx context.Context, arg UpdateBlueprintParams) error {
	_, err := q.db.ExecContext(ctx, updateBlueprint, arg.Note, arg.ID)
	return err
}

const updateBlueprintStep = `-- name: UpdateBlueprintStep :exec
UPDATE blueprint_step
SET content    = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`

type UpdateBlueprintStepParams struct {
	Content string `json:"content"`
	ID      int64  `json:"id"`
}

func (q *Queries) UpdateBlueprintStep(ctx context.Context, arg UpdateBlueprintStepParams) error {
	_, err := q.db.ExecContext(ctx, updateBlueprintStep, arg.Content, arg.ID)
	return err
}
