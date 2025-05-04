-- name: GetFirstFreeLobby :one
SELECT id
FROM lobby
WHERE player_1_id != $1
  AND player_2_id IS NULL
  AND is_private = false
LIMIT 1;

-- name: GetLobbyById :one
SELECT id
FROM lobby
WHERE id = $1
LIMIT 1;