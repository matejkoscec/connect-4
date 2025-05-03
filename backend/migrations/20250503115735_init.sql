-- +goose Up
CREATE TABLE users
(
    id             uuid PRIMARY KEY,
    username       varchar(100) UNIQUE NOT NULL,
    email          varchar(500) UNIQUE NOT NULL,
    password       bytea               NOT NULL,
    created_at_utc timestamptz         NOT NULL
);

CREATE TABLE lobby
(
    id             uuid PRIMARY KEY,
    player_1_id    uuid REFERENCES users (id) NOT NULL,
    player_2_id    uuid REFERENCES users (id),
    created_at_utc timestamptz                NOT NULL
);

CREATE TABLE game
(
    id             uuid PRIMARY KEY,
    lobby_id       uuid REFERENCES lobby (id) NOT NULL,
    started_at_utc timestamptz,
    ended_at_utc   timestamptz,
    state          text                       NOT NULL
);

CREATE TABLE message
(
    id          uuid PRIMARY KEY,
    lobby_id    uuid REFERENCES lobby (id) NOT NULL,
    sender_id   uuid                       NOT NULL,
    content     text                       NOT NULL,
    sent_at_utc timestamptz                NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS message;
DROP TABLE IF EXISTS game;
DROP TABLE IF EXISTS lobby;
DROP TABLE IF EXISTS users;
