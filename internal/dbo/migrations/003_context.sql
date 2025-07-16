-- +migrate Up
CREATE TABLE context
(
    id         INTEGER   NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    title      TEXT      NOT NULL DEFAULT '',
    category   TEXT      NOT NULL DEFAULT '',
    content    TEXT      NOT NULL DEFAULT '',
    -- should be deprecated after migration VVVV
    inline     BOOL      NOT NULL DEFAULT FALSE,
    summary    TEXT      NOT NULL DEFAULT ''
);

CREATE INDEX context_category_idx ON context (category);