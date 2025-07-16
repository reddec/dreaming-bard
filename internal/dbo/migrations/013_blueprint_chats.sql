-- +migrate Up

CREATE TABLE blueprint_chat
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, -- used for consistent sorting in UI
    blueprint_id INTEGER                           NOT NULL REFERENCES blueprint (id) ON DELETE CASCADE,
    chat_id      INTEGER                           NOT NULL REFERENCES chat (id) ON DELETE CASCADE,
    UNIQUE (blueprint_id, chat_id)
);