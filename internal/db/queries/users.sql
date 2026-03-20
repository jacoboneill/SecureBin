-- name: GetUserByEmailOrUsername :one
SELECT * FROM users WHERE email = sqlc.arg(identifier) OR username = sqlc.arg(identifier);

-- name: RegisterUser :one
INSERT INTO users (username, email, password_hash, is_admin)
VALUES (?, ?, ?, ?)
RETURNING *;
