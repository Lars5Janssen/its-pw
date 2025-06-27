CREATE TABLE IF NOT EXISTS users (
    id BYTEA PRIMARY KEY,
    display_name TEXT NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS credentials (
    id BYTEA PRIMARY KEY,
    user_id BYTEA REFERENCES users(id) ON DELETE CASCADE,

    public_key BYTEA NOT NULL,
    attestation_type TEXT NOT NULL,

    transport TEXT[] NOT NULL,

    flags JSONB NOT NULL,
    authenticator JSONB NOT NULL,
    attestation JSONB NOT NULL
);
