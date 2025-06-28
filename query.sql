-- name: GetPwUserByName :one
SELECT * FROM pwusers
WHERE username = $1
LIMIT 1;

-- name: AddPwUser :exec
INSERT INTO pwusers (username, pw, totp_secret)
VALUES ($1, $2, $3);

-- name: DeletePwUserByName :exec
DELETE from pwusers
WHERE username = $1;

-- name: UpdatePwUserPwByName :exec
UPDATE pwusers
SET pw = $2
WHERE username = $1;

-- name: UpdatePwUsertotpByName :exec
UPDATE pwusers
SET totp_secret = $2
WHERE username = $1;

-- name: GetPwUserSessionByName :one
SELECT * FROM pwsessions
WHERE username = $1
LIMIT 1;

-- name: GetPwUserSessionByUuid :many
SELECT * FROM pwsessions
WHERE uuid = $1;

-- name: DeletePwUserSessionByUuid :exec
DELETE FROM pwsessions
WHERE uuid = $1;

-- name: CreatePwUserSession :exec
INSERT INTO pwsessions (username, uuid, expires_at)
VALUES ($1, $2, $3);

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: CreateUser :exec
INSERT INTO users (id, display_name, name, credentials)
VALUES ($1, $2, $3, $4);

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1
LIMIT 1;

-- name: GetSessionByUserId :one
SELECT * FROM sessions
WHERE user_id = $1
LIMIT 1;

-- name: CreateSession :exec
INSERT INTO sessions (user_id, session_id, session_data)
VALUES ($1, $2, $3);

-- name: UpdateUserCredentials :exec
UPDATE users
SET credentials = $2
WHERE id = $1;

-- name: DeleteSessionByUserId :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: GetSessionBySessionId :one
SELECT * FROM sessions
WHERE session_id = $1
LIMIT 1;
