-- name: GetSessionByUserId :one
SELECT * FROM sessions
WHERE user_id = $1
LIMIT 1;

-- name: CreateSession :exec
INSERT INTO sessions (user_id, session_id, session_data)
VALUES ($1, $2, $3);


-- name: DeleteSessionByUserId :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: GetSessionBySessionId :one
SELECT * FROM sessions
WHERE session_id = $1
LIMIT 1;
