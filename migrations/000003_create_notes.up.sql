CREATE TYPE note_status AS ENUM ('todo', 'in_progress', 'done');

CREATE TABLE notes (
                       id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                       space_id   BIGINT      NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
                       title      TEXT        NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
                       content    TEXT        NOT NULL DEFAULT '',
                       status     note_status NOT NULL DEFAULT 'todo',
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notes_space_id ON notes(space_id);

CREATE TRIGGER trg_notes_updated_at
    BEFORE UPDATE ON notes
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();