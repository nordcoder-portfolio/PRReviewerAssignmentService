-- +goose Up
CREATE TYPE pr_status_type AS ENUM ('OPEN', 'MERGED');

-- +goose Down
DROP TYPE IF EXISTS pr_status_type;
