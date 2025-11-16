-- +goose Up
CREATE TABLE pull_requests
(
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT           NOT NULL,
    author_id         TEXT           NOT NULL REFERENCES users (user_id),
    status            pr_status_type NOT NULL DEFAULT 'OPEN',
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT now(),
    merged_at         TIMESTAMPTZ
);

CREATE INDEX idx_pull_requests_author_id ON pull_requests (author_id);
CREATE INDEX idx_pull_requests_status ON pull_requests (status);

-- +goose Down
DROP INDEX IF EXISTS idx_pull_requests_status;
DROP INDEX IF EXISTS idx_pull_requests_author_id;
DROP TABLE IF EXISTS pull_requests;
