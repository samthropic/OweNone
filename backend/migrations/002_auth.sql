ALTER TABLE users ADD COLUMN password_hash text NOT NULL DEFAULT '';

CREATE TABLE sessions (
    token_hash text PRIMARY KEY CHECK (length(token_hash) = 64),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

DROP TABLE waitlist_entries;
