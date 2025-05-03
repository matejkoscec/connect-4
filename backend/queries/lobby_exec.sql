-- name: CreateLobby :exec
INSERT INTO lobby (id, player_1_id, created_at_utc)
VALUES ($1, $2, $3);
