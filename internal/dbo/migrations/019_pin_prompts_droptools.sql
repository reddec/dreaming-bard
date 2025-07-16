-- +migrate Up

ALTER TABLE prompt
    DROP COLUMN default_tools;

ALTER TABLE prompt
    ADD COLUMN pinned_at TIMESTAMP;
