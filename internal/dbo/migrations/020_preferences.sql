-- +migrate Up

-- one day it may become user-specific, but currently the application is strictly single-tenant
CREATE TABLE preference
(
    name       TEXT      NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    value      JSONB     NOT NULL
);

INSERT INTO preference (name, value)
VALUES ('help_chats', 'true'),
       ('help_prompts', 'true'),
       ('help_context', 'true'),
       ('help_pages', 'true'),
       ('help_blueprints', 'true'),
       ('help_roles', 'true');;