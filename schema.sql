CREATE TABLE IF NOT EXISTS pwusers (
    username TEXT PRIMARY KEY,
    pw TEXT,
)

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
