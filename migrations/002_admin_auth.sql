CREATE TABLE admin_sessions (
    id UUID PRIMARY KEY,
    admin_user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NULL,
    revoked_at TIMESTAMPTZ NULL
);

CREATE INDEX admin_sessions_admin_user_id_idx ON admin_sessions (admin_user_id);
CREATE INDEX admin_sessions_expires_at_idx ON admin_sessions (expires_at);
