-- +goose Up
ALTER TABLE lobby
    ALTER COLUMN player_2_id SET NOT NULL;

-- +goose Down
ALTER TABLE lobby
    ALTER COLUMN player_2_id DROP NOT NULL;
