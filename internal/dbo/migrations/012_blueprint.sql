-- +migrate Up
CREATE TABLE blueprint
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    created_at TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    note       TEXT                              NOT NULL DEFAULT ''
);

CREATE TABLE blueprint_step
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    created_at   TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    blueprint_id INTEGER                           NOT NULL REFERENCES blueprint (id) ON DELETE CASCADE,
    content      TEXT                              NOT NULL DEFAULT ''
);

CREATE TABLE blueprint_linked_context
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, -- used for consistent sorting in UI
    blueprint_id INTEGER                           NOT NULL REFERENCES blueprint (id) ON DELETE CASCADE,
    context_id   INTEGER                           NOT NULL REFERENCES context (id) ON DELETE CASCADE,
    UNIQUE (blueprint_id, context_id)
);

CREATE TABLE blueprint_linked_page
(
    blueprint_id INTEGER NOT NULL REFERENCES blueprint (id) ON DELETE CASCADE,
    page_id      INTEGER NOT NULL REFERENCES page (id) ON DELETE CASCADE,
    inline       BOOL    NOT NULL DEFAULT FALSE, -- if TRUE - full page referenced, otherwise only summary/truncated
    PRIMARY KEY (blueprint_id, page_id)
);