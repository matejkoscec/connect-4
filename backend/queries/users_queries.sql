-- name: GetUserByUsernameOrEmail :one
SELECT *
FROM users
WHERE username = $1
   OR email = $2
LIMIT 1;

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;