-- +migrate Up
ALTER TABLE chat
    DROP COLUMN inline_facts;

ALTER TABLE prompt
    DROP COLUMN default_facts;