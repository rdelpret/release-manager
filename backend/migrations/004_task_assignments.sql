-- Add assignee support to tasks
ALTER TABLE tasks ADD COLUMN assigned_to UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
