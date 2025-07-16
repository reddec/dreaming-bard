-- +migrate Up
ALTER TABLE context DROP COLUMN
    inline;

ALTER TABLE context DROP COLUMN
    summary;