-- Day 11: users and rotating refresh tokens.
-- Only the SHA-256 (hex, 64 chars) of a refresh token is stored, never the token itself.
CREATE TABLE users (
    username      text PRIMARY KEY CHECK (username ~ '^[A-Za-z0-9._:@-]{1,64}$'),
    password_hash text NOT NULL,
    role          text NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'staff')),
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    token_hash text PRIMARY KEY CHECK (char_length(token_hash) = 64),
    family_id  text NOT NULL,
    username   text NOT NULL REFERENCES users (username) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX refresh_tokens_family_idx ON refresh_tokens (family_id);
CREATE INDEX refresh_tokens_expires_idx ON refresh_tokens (expires_at);
