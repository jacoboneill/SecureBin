-- name: CreateSession :one
INSERT INTO sessions (id, user_id)
VALUES (?, ?)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions WHERE id = ? LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = ?;
