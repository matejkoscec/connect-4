-- name: CreateLobby :exec
INSERT INTO lobby (id, player_1_id, player_2_id, created_at_utc, is_private)
VALUES ($1, $2, $3, $4, false);
