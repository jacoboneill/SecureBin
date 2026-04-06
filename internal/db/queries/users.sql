-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByEmailOrUsername :one
SELECT * FROM users WHERE email = sqlc.arg(identifier) OR username = sqlc.arg(identifier) LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, is_admin)
VALUES (?, ?, ?, ?)
RETURNING *;
