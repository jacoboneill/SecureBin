-- name: GetUserByEmailOrUsername :one
SELECT
  *
FROM
  users
WHERE
  email = ?
  OR username = ?
