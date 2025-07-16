-- +migrate Up
UPDATE role SET purpose = 'write' WHERE purpose = 'writer';

