-- name: CreateImplSession :exec
INSERT INTO implsessions (sid, username, client_nounce, own_nounce, session_key)
VALUES ($1, $2, $3, $4, $5);

-- name: GetSessionKeyBySID :one
SELECT session_key FROM implsessions
WHERE sid = $1
LIMIT 1;

-- name: DeleteAllSessions :exec
DELETE FROM implsessions;

-- name: GetImplUserNameFromSID :one
SELECT username FROM implsessions
WHERE sid = $1
LIMIT 1;

-- name: GetSIDbyUserName :one
SELECT sid FROM implsessions
WHERE username = $1
LIMIT 1;

-- name: GetClientNouncebyUserName :one
SELECT client_nounce FROM implsessions
WHERE username = $1
LIMIT 1;

-- name: GetOwnNouncebyUserName :one
SELECT own_nounce FROM implsessions
WHERE username = $1
LIMIT 1;

-- name: GetSessionKeybyUserName :one
SELECT session_key FROM implsessions
WHERE username = $1
LIMIT 1;

-- name: DeleteImplSessionByUsername :exec
DELETE FROM implsessions
WHERE username = $1;
