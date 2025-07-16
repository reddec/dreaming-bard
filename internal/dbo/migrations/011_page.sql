-- +migrate Up
DROP TABLE IF EXISTS page; -- dev mistake, didn't work anyway

CREATE TABLE page
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    created_at TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    summary    TEXT                              NOT NULL DEFAULT '',
    content    TEXT                              NOT NULL DEFAULT '',
    num        INT                               NOT NULL DEFAULT 1
);

CREATE INDEX page_num_idx ON page (num);