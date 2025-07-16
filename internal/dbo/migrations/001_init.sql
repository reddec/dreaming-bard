-- +migrate Up

CREATE TABLE chat
(
    id            INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    created_at    TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role          TEXT                              NOT NULL, -- role name: reference to role definition (TBD: replace to table)
    inline_facts  JSONB                             NOT NULL DEFAULT '[]',
    num_pages     INTEGER                           NOT NULL DEFAULT 3,
    input_tokens  INTEGER                           NOT NULL DEFAULT 0,
    output_tokens INTEGER                           NOT NULL DEFAULT 0
);

CREATE TABLE message
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    chat_id    INTEGER                           NOT NULL REFERENCES chat (id) ON DELETE CASCADE,
    created_at TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    content    TEXT                              NOT NULL,
    role       TEXT                              NOT NULL, -- user, assistant, tool_call, tool_result - see common.role
    tool_id    TEXT                              NOT NULL DEFAULT '',
    tool_name  TEXT                              NOT NULL DEFAULT ''
);