-- +goose Up
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_campaigns_archived ON campaigns(archived);

-- +goose Down
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_campaigns_archived;
