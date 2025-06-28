CREATE TABLE IF NOT EXISTS pwusers (
    username TEXT PRIMARY KEY,
    pw BYTEA,
    totp_secret BYTEA
);

CREATE TABLE IF NOT EXISTS pwsessions (
    username TEXT PRIMARY KEY,
    uuid TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id BYTEA PRIMARY KEY,
    display_name TEXT NOT NULL,
    name TEXT NOT NULL,
    credentials JSONB
);

CREATE TABLE IF NOT EXISTS sessions (
    user_id BYTEA PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    session_id TEXT NOT NULL,
    session_data JSONB
);
