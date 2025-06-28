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
