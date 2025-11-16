-- +goose Up
CREATE TABLE pr_reviewers
(
    pull_request_id TEXT NOT NULL REFERENCES pull_requests (pull_request_id) ON DELETE CASCADE,
    reviewer_id     TEXT NOT NULL REFERENCES users (user_id),
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX idx_pr_reviewers_reviewer ON pr_reviewers (reviewer_id);

-- +goose Down
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer;
DROP TABLE IF EXISTS pr_reviewers;
