-- +migrate Up

ALTER TABLE chat
    ADD COLUMN annotation TEXT NOT NULL DEFAULT '';
ALTER TABLE chat
    DROP COLUMN num_pages;

