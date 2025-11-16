-- +goose Up
CREATE TABLE users
(
    user_id   TEXT PRIMARY KEY,
    username  TEXT    NOT NULL,
    team_id   BIGINT  NOT NULL REFERENCES teams (id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_users_team_id ON users (team_id);
CREATE INDEX idx_users_team_active ON users (team_id, is_active);

-- +goose Down
DROP INDEX IF EXISTS idx_users_team_active;
DROP INDEX IF EXISTS idx_users_team_id;
DROP TABLE IF EXISTS users;
