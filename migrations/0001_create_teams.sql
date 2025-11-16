-- +goose Up
CREATE TABLE teams
(
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE IF EXISTS teams;
