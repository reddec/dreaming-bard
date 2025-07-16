-- +migrate Up

ALTER TABLE chat
    ADD COLUMN draft TEXT NOT NULL DEFAULT '';

CREATE TABLE prompt
(
    id                INTEGER   NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    summary           TEXT      NOT NULL,
    content           TEXT      NOT NULL,
    default_role      TEXT      NOT NULL,
    default_num_pages INTEGER   NOT NULL DEFAULT 3,
    default_facts     JSONB     NOT NULL DEFAULT '[]',
    default_tools     JSONB     NOT NULL DEFAULT '[]'
);