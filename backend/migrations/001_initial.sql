CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email      TEXT UNIQUE NOT NULL,
  name       TEXT NOT NULL,
  avatar_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE campaigns (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_by UUID NOT NULL REFERENCES users(id),
  name       TEXT NOT NULL,
  archived   BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE campaign_members (
  campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role        TEXT NOT NULL CHECK (role IN ('owner', 'editor')),
  PRIMARY KEY (campaign_id, user_id)
);

CREATE TABLE task_lists (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  color       TEXT NOT NULL,
  position    INTEGER NOT NULL
);

CREATE TABLE task_groups (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_list_id UUID NOT NULL REFERENCES task_lists(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  position     INTEGER NOT NULL,
  collapsed    BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE tasks (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_group_id UUID NOT NULL REFERENCES task_groups(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  description   JSONB,
  status        TEXT NOT NULL CHECK (status IN ('todo', 'in_progress', 'done')) DEFAULT 'todo',
  due_date      DATE,
  position      INTEGER NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE subtasks (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id     UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  is_complete BOOLEAN NOT NULL DEFAULT false,
  position    INTEGER NOT NULL
);

-- Note: sessions are stored in cookies via gorilla/sessions, no DB session table needed

-- Indexes for common queries
CREATE INDEX idx_campaign_members_user ON campaign_members(user_id);
CREATE INDEX idx_task_lists_campaign ON task_lists(campaign_id);
CREATE INDEX idx_task_groups_list ON task_groups(task_list_id);
CREATE INDEX idx_tasks_group ON tasks(task_group_id);
CREATE INDEX idx_subtasks_task ON subtasks(task_id);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
