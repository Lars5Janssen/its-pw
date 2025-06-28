-- name: GetAllUsers :many
SELECT * FROM users;

-- name: CreateUser :exec
INSERT INTO users (id, display_name, name, credentials)
VALUES ($1, $2, $3, $4);

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1
LIMIT 1;
