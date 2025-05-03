-- name: CreateUser :exec
INSERT INTO users (id, username, email, password, created_at_utc)
VALUES ($1, $2, $3, $4, $5);

