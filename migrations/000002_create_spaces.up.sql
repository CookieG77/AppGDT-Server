CREATE TABLE spaces (
                        id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                        user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        name        TEXT        NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
                        description TEXT        NOT NULL DEFAULT '',
                        created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
                        updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_spaces_user_id ON spaces(user_id);

CREATE TRIGGER trg_spaces_updated_at
    BEFORE UPDATE ON spaces
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();