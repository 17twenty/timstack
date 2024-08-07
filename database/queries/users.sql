-- name: CreateUser :one
INSERT INTO users (id, display_name, name, credentials)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1 LIMIT 1;

-- name: InsertIntoSessions :one
INSERT INTO sessions (id, user_id, data, expires)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1 LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;


-- name: GetUserByUsername :one
SELECT * FROM users
WHERE name = $1 LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;
