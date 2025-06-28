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

-- name: UpdateUserCredentials :exec
UPDATE users
SET credentials = $2
WHERE id = $1;
