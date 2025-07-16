-- +migrate Up

CREATE TABLE role
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    created_at TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name       TEXT                              NOT NULL DEFAULT '',
    system     TEXT                              NOT NULL DEFAULT '', -- system prompt
    model      TEXT                              NOT NULL DEFAULT '', -- model name
    purpose    TEXT                              NOT NULL DEFAULT ''  -- role category: writer, summary, etc... (see common.Purpose)
);

INSERT INTO role (name, system, purpose)
VALUES ('writer', 'You are a creative story writer.', 'write');
INSERT INTO role (name, system, purpose)
VALUES ('summariser', 'You are a text summarization specialist. Your task is to create extremely concise summaries that preserve essential information for LLM context windows.

INSTRUCTIONS:
1. Extract only the most critical information from the input
2. Maintain the same language as the input text
3. Limit output to 2-3 sentences maximum (preferably 1-2)
4. Focus on: main topic, key facts, and primary purpose/outcome
5. Omit: examples, elaborations, redundancies, and stylistic elements
6. Use neutral, factual tone
7. Preserve technical terms and proper nouns exactly as written

FORMAT: Write as a single dense paragraph without bullet points or formatting.

PRIORITY: Maximum information density with minimum word count.
', 'summary');


ALTER TABLE chat
    DROP COLUMN role;
ALTER TABLE chat
    ADD COLUMN role_id INTEGER NOT NULL DEFAULT 1 REFERENCES role (id) ON DELETE CASCADE;


ALTER TABLE prompt
    DROP COLUMN default_role;

ALTER TABLE prompt
    ADD COLUMN role_id INTEGER NOT NULL DEFAULT 1 REFERENCES role (id) ON DELETE CASCADE;

