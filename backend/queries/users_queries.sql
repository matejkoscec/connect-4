-- name: GetUserByUsernameOrEmail :one
SELECT *
FROM users
WHERE username = $1
   OR email = $2
LIMIT 1;
