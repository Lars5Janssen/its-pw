-- name: GetAllUsers :many
SELECT * FROM users;

-- name: CreateUser :exec
INSERT INTO users (id, display_name, name)
VALUES ($1, $2, $3);

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1
LIMIT 1;

-- name: GetUserCredsById :one
SELECT * from credentials
WHERE id = $1
LIMIT 1;
