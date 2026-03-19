-- name: GetUserByEmailOrUsername :one
SELECT * FROM users WHERE email = ? OR username = ?;

-- name: RegisterUser :one
INSERT INTO users (username, email, password_hash, is_admin)
VALUES (?, ?, ?, ?)
RETURNING *;
