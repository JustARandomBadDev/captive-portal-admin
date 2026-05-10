CREATE TABLE admin_users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NULL,
    display_name TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE pitches (
    id UUID PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    label TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wifi_tickets (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    cleartext_password TEXT NOT NULL,
    pitch_id UUID NOT NULL REFERENCES pitches(id),
    status TEXT NOT NULL,
    valid_from TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ NOT NULL,
    created_by UUID NULL REFERENCES admin_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ NULL,
    revoked_by UUID NULL REFERENCES admin_users(id),
    radius_synced_at TIMESTAMPTZ NULL,
    CONSTRAINT wifi_tickets_status_check CHECK (status IN ('active', 'expired', 'revoked')),
    CONSTRAINT wifi_tickets_valid_dates_check CHECK (valid_until > valid_from)
);

CREATE INDEX pitches_code_idx ON pitches (code);
CREATE INDEX pitches_is_active_idx ON pitches (is_active);

CREATE INDEX wifi_tickets_username_idx ON wifi_tickets (username);
CREATE INDEX wifi_tickets_pitch_id_idx ON wifi_tickets (pitch_id);
CREATE INDEX wifi_tickets_status_idx ON wifi_tickets (status);
CREATE INDEX wifi_tickets_valid_until_idx ON wifi_tickets (valid_until);
CREATE INDEX wifi_tickets_radius_synced_at_idx ON wifi_tickets (radius_synced_at);
CREATE INDEX wifi_tickets_pitch_id_status_idx ON wifi_tickets (pitch_id, status);
CREATE INDEX wifi_tickets_status_valid_until_idx ON wifi_tickets (status, valid_until);
