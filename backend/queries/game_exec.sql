-- name: CreateGame :exec
INSERT INTO game (id, lobby_id, started_at_utc, state)
VALUES ($1, $2, $3, $4);
