-- +migrate Up

ALTER TABLE prompt
    DROP COLUMN default_num_pages;
