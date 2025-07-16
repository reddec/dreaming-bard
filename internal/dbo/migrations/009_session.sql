-- +migrate Up

CREATE TABLE session
(
    token      TEXT      NOT NULL PRIMARY KEY,
    content    BLOB      NOT NULL,
    expires_at TIMESTAMP NOT NULL
);