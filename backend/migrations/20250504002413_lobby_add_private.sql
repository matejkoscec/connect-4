-- +goose Up
ALTER TABLE lobby
    ADD COLUMN is_private bool NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE lobby
    DROP COLUMN is_private;
