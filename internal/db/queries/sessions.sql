-- name: CreateSession :one
INSERT INTO sessions (id, user_id)
VALUES (?, ?)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions WHERE id = ? LIMIT 1;
