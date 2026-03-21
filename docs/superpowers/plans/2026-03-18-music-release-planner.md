# Music Release Planner Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an internal release campaign management tool for the Subwave team — Go API backend + Next.js frontend, deployed on Cloudflare.

**Architecture:** Go monolith (chi router, pgx, Neon PostgreSQL) serves a REST API. Next.js frontend (App Router, TailwindCSS, shadcn/ui) consumes it. Google OAuth with email whitelist for 2 users. Session-based auth with cookies. Cloudflare Workers serves static frontend and proxies `/api/*` to a Go container.

**Tech Stack:** Go 1.24+, chi/v5, pgx/v5, gorilla/sessions, golang.org/x/oauth2 | Next.js 16+, React 19, TailwindCSS 4, shadcn/ui, @dnd-kit, Tiptap, TanStack Query | Neon PostgreSQL | Cloudflare Workers + Containers

**Spec:** `docs/superpowers/specs/2026-03-18-music-release-planner-design.md`

---

## File Structure

```
/backend
  cmd/server/main.go                    — Entry point, env loading, init auth/db, start server
  internal/
    auth/
      auth.go                           — Google OAuth flow, session store, middleware
      auth_test.go                      — Auth handler + middleware tests
    handler/
      handler.go                        — Server struct, writeJSON/writeError helpers
      routes.go                         — Chi router setup, route registration
      campaign.go                       — Campaign CRUD handlers
      campaign_test.go                  — Campaign handler tests
      task.go                           — Task + subtask handlers
      task_test.go                      — Task handler tests
      reorder.go                        — Drag-and-drop reorder handlers
      reorder_test.go                   — Reorder handler tests
    middleware/
      cors.go                           — CORS middleware
    model/
      model.go                          — All Go structs (User, Campaign, TaskList, etc.)
    store/
      store.go                          — DB pool init, Store struct
      user.go                           — User queries
      campaign.go                       — Campaign queries (create, list, get full, duplicate, archive, delete)
      campaign_test.go                  — Campaign store integration tests
      task.go                           — Task + subtask queries
      task_test.go                      — Task store integration tests
      reorder.go                        — Reorder queries (position swaps)
      reorder_test.go                   — Reorder store integration tests
    template/
      template.go                       — Default campaign template data (5 lists, groups, tasks)
  migrations/
    001_initial.sql                     — Full schema DDL
  Dockerfile                            — Multi-stage Go build
  go.mod
  go.sum
  .air.toml                             — Hot reload config

/frontend
  src/
    app/
      layout.tsx                        — Root layout, QueryClientProvider, auth context, fonts, theme
      login/page.tsx                    — Google OAuth login page
      dashboard/page.tsx                — Campaign list grid
      campaign/[id]/page.tsx            — Campaign board (tabbed task lists)
      campaign/[id]/calendar/page.tsx   — Calendar view
      globals.css                       — TailwindCSS + Subwave theme variables
    components/
      ui/                               — shadcn/ui components (restyled)
      campaign-card.tsx                 — Dashboard campaign card
      task-list-tabs.tsx                — Tab bar for task lists
      task-group.tsx                    — Collapsible task group
      task-item.tsx                     — Task row with status indicator
      task-detail.tsx                   — Slide-out panel (rich text, subtasks, due date)
      subtask-item.tsx                  — Checkbox subtask row
      calendar-view.tsx                 — Calendar with tasks by due date
      hide-done-toggle.tsx              — Toggle to hide completed tasks
    lib/
      api.ts                            — Fetch wrapper for Go backend
      auth.tsx                          — Auth context + hooks
      types.ts                          — TypeScript types matching Go models
    hooks/
      use-campaign.ts                   — Campaign data fetching + mutations (TanStack Query)
      use-drag-drop.ts                  — @dnd-kit drag-and-drop logic
  next.config.ts                        — API rewrite to Go backend
  package.json
  components.json                       — shadcn/ui config
  tsconfig.json
  postcss.config.mjs

/src
  index.ts                              — Cloudflare Worker entry point
wrangler.toml                           — Cloudflare config
Makefile                                — Build orchestration
```

---

## Task 1: Project Scaffolding + Go Module

**Files:**
- Create: `backend/cmd/server/main.go`
- Create: `backend/go.mod`
- Create: `backend/internal/handler/handler.go`
- Create: `backend/internal/handler/routes.go`
- Create: `backend/.air.toml`
- Create: `Makefile`

- [ ] **Step 1: Initialize Go module**

```bash
cd backend
go mod init github.com/rdelpret/music-release-planner/backend
```

- [ ] **Step 2: Create main.go entry point**

Create `backend/cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rdelpret/music-release-planner/backend/internal/handler"
)

func main() {
	// Load .env from backend/ or project root
	godotenv.Load()
	godotenv.Load("../.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := handler.NewServer()

	fmt.Printf("Server running at http://localhost:%s\n", port)
	if err := srv.Start(port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

- [ ] **Step 3: Create handler scaffolding**

Create `backend/internal/handler/handler.go`:

```go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	router chi.Router
}

func NewServer() *Server {
	s := &Server{}
	s.router = s.routes()
	return s
}

func (s *Server) Start(port string) error {
	return http.ListenAndServe(fmt.Sprintf(":%s", port), s.router)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
```

Create `backend/internal/handler/routes.go`:

```go
package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func (s *Server) routes() chi.Router {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return r
}
```

- [ ] **Step 4: Create .air.toml for hot reload**

Create `backend/.air.toml`:

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/server ./cmd/server/main.go"
  bin = "./tmp/server"
  delay = 1000
  exclude_dir = ["tmp", "vendor", "migrations"]
  exclude_regex = ["_test\\.go$"]
  include_ext = ["go"]
  kill_delay = "0s"

[misc]
  clean_on_exit = true
```

- [ ] **Step 5: Create Makefile**

Create `Makefile`:

```makefile
.PHONY: dev dev-backend dev-frontend test test-backend test-frontend build clean migrate

dev:
	@echo "Starting backend and frontend..."
	@make dev-backend & make dev-frontend & wait

dev-backend:
	cd backend && $(HOME)/go/bin/air

dev-frontend:
	cd frontend && npm run dev

test: test-backend test-frontend

test-backend:
	cd backend && go test ./... -v -count=1

test-frontend:
	cd frontend && npm test

build:
	cd backend && go build -o ../bin/server ./cmd/server/main.go
	cd frontend && npm run build

clean:
	rm -rf bin/
	rm -rf frontend/.next

migrate:
	@echo "Run: psql $$DATABASE_URL -f backend/migrations/001_initial.sql"
```

- [ ] **Step 6: Install Go dependencies and verify**

```bash
cd backend
go get github.com/go-chi/chi/v5
go get github.com/joho/godotenv
go mod tidy
go build ./cmd/server/
```

Expected: Build succeeds with no errors.

- [ ] **Step 7: Commit**

```bash
git add backend/cmd/ backend/internal/handler/ backend/go.mod backend/go.sum backend/.air.toml Makefile
git commit -m "feat: scaffold Go backend with chi router and health endpoint"
```

---

## Task 2: Database Schema + Store Initialization

**Files:**
- Create: `backend/migrations/001_initial.sql`
- Create: `backend/internal/store/store.go`
- Create: `backend/internal/model/model.go`

- [ ] **Step 1: Write the migration file**

Create `backend/migrations/001_initial.sql`:

```sql
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
```

- [ ] **Step 2: Create Go models**

Create `backend/internal/model/model.go`:

```go
package model

import (
	"encoding/json"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Campaign struct {
	ID        string     `json:"id"`
	CreatedBy string     `json:"created_by"`
	Name      string     `json:"name"`
	Archived  bool       `json:"archived"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	TaskLists []TaskList `json:"task_lists,omitempty"`
}

type CampaignMember struct {
	CampaignID string `json:"campaign_id"`
	UserID     string `json:"user_id"`
	Role       string `json:"role"`
}

type TaskList struct {
	ID         string      `json:"id"`
	CampaignID string      `json:"campaign_id"`
	Name       string      `json:"name"`
	Color      string      `json:"color"`
	Position   int         `json:"position"`
	TaskGroups []TaskGroup `json:"task_groups,omitempty"`
}

type TaskGroup struct {
	ID         string `json:"id"`
	TaskListID string `json:"task_list_id"`
	Name       string `json:"name"`
	Position   int    `json:"position"`
	Collapsed  bool   `json:"collapsed"`
	Tasks      []Task `json:"tasks,omitempty"`
}

type Task struct {
	ID          string           `json:"id"`
	TaskGroupID string           `json:"task_group_id"`
	Name        string           `json:"name"`
	Description *json.RawMessage `json:"description,omitempty"`
	Status      string           `json:"status"`
	DueDate     *string          `json:"due_date,omitempty"`
	Position    int              `json:"position"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Subtasks    []Subtask        `json:"subtasks,omitempty"`
}

type Subtask struct {
	ID         string `json:"id"`
	TaskID     string `json:"task_id"`
	Name       string `json:"name"`
	IsComplete bool   `json:"is_complete"`
	Position   int    `json:"position"`
}
```

- [ ] **Step 3: Create store initialization**

Create `backend/internal/store/store.go`:

```go
package store

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New() (*Store, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parsing database config: %w", err)
	}
	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}
```

- [ ] **Step 4: Install pgx dependency**

```bash
cd backend
go get github.com/jackc/pgx/v5
go mod tidy
go build ./...
```

Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/ backend/internal/model/ backend/internal/store/store.go backend/go.mod backend/go.sum
git commit -m "feat: add database schema, Go models, and store initialization"
```

---

## Task 3: User Store + Campaign Store (Core Queries)

**Files:**
- Create: `backend/internal/store/user.go`
- Create: `backend/internal/store/campaign.go`
- Create: `backend/internal/store/campaign_test.go`

- [ ] **Step 1: Write failing campaign store tests**

Create `backend/internal/store/campaign_test.go`:

```go
//go:build integration

package store

import (
	"context"
	"os"
	"testing"

	"github.com/rdelpret/music-release-planner/backend/internal/model"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}
	s, err := New()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func createTestUser(t *testing.T, s *Store) *model.User {
	t.Helper()
	user, err := s.UpsertUser(context.Background(), "test@example.com", "Test User", nil)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestCreateCampaign(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)

	campaign, err := s.CreateCampaign(context.Background(), user.ID, "Test Release")
	if err != nil {
		t.Fatalf("failed to create campaign: %v", err)
	}
	if campaign.Name != "Test Release" {
		t.Errorf("expected name 'Test Release', got '%s'", campaign.Name)
	}
	if campaign.CreatedBy != user.ID {
		t.Errorf("expected created_by '%s', got '%s'", user.ID, campaign.CreatedBy)
	}

	// Cleanup
	s.DeleteCampaign(context.Background(), campaign.ID)
}

func TestListCampaigns(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)

	c1, _ := s.CreateCampaign(context.Background(), user.ID, "Campaign 1")
	c2, _ := s.CreateCampaign(context.Background(), user.ID, "Campaign 2")
	defer s.DeleteCampaign(context.Background(), c1.ID)
	defer s.DeleteCampaign(context.Background(), c2.ID)

	campaigns, err := s.ListCampaigns(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("failed to list campaigns: %v", err)
	}
	if len(campaigns) < 2 {
		t.Errorf("expected at least 2 campaigns, got %d", len(campaigns))
	}
}

func TestGetFullCampaign(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)

	campaign, _ := s.CreateCampaign(context.Background(), user.ID, "Full Test")
	defer s.DeleteCampaign(context.Background(), campaign.ID)

	full, err := s.GetFullCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("failed to get full campaign: %v", err)
	}
	if full.Name != "Full Test" {
		t.Errorf("expected name 'Full Test', got '%s'", full.Name)
	}
	// Template population is a stub until Task 7 — no task lists expected yet
	// After Task 7, this test should verify len(full.TaskLists) == 5
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend
go test ./internal/store/ -v -count=1 -tags=integration -run TestCreateCampaign
```

Expected: FAIL — functions don't exist yet.

- [ ] **Step 3: Implement user store**

Create `backend/internal/store/user.go`:

```go
package store

import (
	"context"

	"github.com/rdelpret/music-release-planner/backend/internal/model"
)

func (s *Store) UpsertUser(ctx context.Context, email, name string, avatarURL *string) (*model.User, error) {
	var user model.User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, name, avatar_url)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET name = $2, avatar_url = $3
		RETURNING id, email, name, avatar_url, created_at
	`, email, name, avatarURL).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, name, avatar_url, created_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, name, avatar_url, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
```

- [ ] **Step 4: Implement campaign store**

Create `backend/internal/store/campaign.go`:

```go
package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rdelpret/music-release-planner/backend/internal/model"
)

func (s *Store) CreateCampaign(ctx context.Context, userID, name string) (*model.Campaign, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var campaign model.Campaign
	err = tx.QueryRow(ctx, `
		INSERT INTO campaigns (created_by, name)
		VALUES ($1, $2)
		RETURNING id, created_by, name, archived, created_at, updated_at
	`, userID, name).Scan(&campaign.ID, &campaign.CreatedBy, &campaign.Name,
		&campaign.Archived, &campaign.CreatedAt, &campaign.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Add creator as owner
	_, err = tx.Exec(ctx, `
		INSERT INTO campaign_members (campaign_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, campaign.ID, userID)
	if err != nil {
		return nil, err
	}

	// Populate default template (stub — replaced with real implementation in Task 7)
	if err := s.populateTemplate(ctx, tx, campaign.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (s *Store) ListCampaigns(ctx context.Context, userID string) ([]model.Campaign, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.created_by, c.name, c.archived, c.created_at, c.updated_at
		FROM campaigns c
		JOIN campaign_members cm ON cm.campaign_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []model.Campaign
	for rows.Next() {
		var c model.Campaign
		if err := rows.Scan(&c.ID, &c.CreatedBy, &c.Name, &c.Archived, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, rows.Err()
}

func (s *Store) GetFullCampaign(ctx context.Context, campaignID string) (*model.Campaign, error) {
	var campaign model.Campaign
	err := s.pool.QueryRow(ctx, `
		SELECT id, created_by, name, archived, created_at, updated_at
		FROM campaigns WHERE id = $1
	`, campaignID).Scan(&campaign.ID, &campaign.CreatedBy, &campaign.Name,
		&campaign.Archived, &campaign.CreatedAt, &campaign.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Fetch task lists
	listRows, err := s.pool.Query(ctx, `
		SELECT id, campaign_id, name, color, position
		FROM task_lists WHERE campaign_id = $1
		ORDER BY position
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer listRows.Close()

	for listRows.Next() {
		var tl model.TaskList
		if err := listRows.Scan(&tl.ID, &tl.CampaignID, &tl.Name, &tl.Color, &tl.Position); err != nil {
			return nil, err
		}
		campaign.TaskLists = append(campaign.TaskLists, tl)
	}
	if err := listRows.Err(); err != nil {
		return nil, err
	}

	// Fetch groups and tasks for each list
	for i := range campaign.TaskLists {
		groups, err := s.getGroupsForList(ctx, campaign.TaskLists[i].ID)
		if err != nil {
			return nil, err
		}
		campaign.TaskLists[i].TaskGroups = groups
	}

	return &campaign, nil
}

func (s *Store) getGroupsForList(ctx context.Context, listID string) ([]model.TaskGroup, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, task_list_id, name, position, collapsed
		FROM task_groups WHERE task_list_id = $1
		ORDER BY position
	`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.TaskGroup
	for rows.Next() {
		var g model.TaskGroup
		if err := rows.Scan(&g.ID, &g.TaskListID, &g.Name, &g.Position, &g.Collapsed); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch tasks + subtasks for each group
	for i := range groups {
		tasks, err := s.getTasksForGroup(ctx, groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].Tasks = tasks
	}
	return groups, nil
}

func (s *Store) getTasksForGroup(ctx context.Context, groupID string) ([]model.Task, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, task_group_id, name, description, status, due_date, position, created_at, updated_at
		FROM tasks WHERE task_group_id = $1
		ORDER BY position
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.TaskGroupID, &t.Name, &t.Description, &t.Status,
			&t.DueDate, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch subtasks
	for i := range tasks {
		subtasks, err := s.getSubtasksForTask(ctx, tasks[i].ID)
		if err != nil {
			return nil, err
		}
		tasks[i].Subtasks = subtasks
	}
	return tasks, nil
}

func (s *Store) getSubtasksForTask(ctx context.Context, taskID string) ([]model.Subtask, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, task_id, name, is_complete, position
		FROM subtasks WHERE task_id = $1
		ORDER BY position
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subtasks []model.Subtask
	for rows.Next() {
		var st model.Subtask
		if err := rows.Scan(&st.ID, &st.TaskID, &st.Name, &st.IsComplete, &st.Position); err != nil {
			return nil, err
		}
		subtasks = append(subtasks, st)
	}
	return subtasks, rows.Err()
}

func (s *Store) ArchiveCampaign(ctx context.Context, campaignID string, archived bool) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE campaigns SET archived = $2, updated_at = $3
		WHERE id = $1
	`, campaignID, archived, time.Now())
	return err
}

func (s *Store) DeleteCampaign(ctx context.Context, campaignID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM campaigns WHERE id = $1`, campaignID)
	return err
}

func (s *Store) DuplicateCampaign(ctx context.Context, sourceCampaignID, userID string) (*model.Campaign, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get source campaign name
	var sourceName string
	err = tx.QueryRow(ctx, `SELECT name FROM campaigns WHERE id = $1`, sourceCampaignID).Scan(&sourceName)
	if err != nil {
		return nil, err
	}

	// Create new campaign
	var campaign model.Campaign
	err = tx.QueryRow(ctx, `
		INSERT INTO campaigns (created_by, name)
		VALUES ($1, $2)
		RETURNING id, created_by, name, archived, created_at, updated_at
	`, userID, sourceName+" (Copy)").Scan(&campaign.ID, &campaign.CreatedBy, &campaign.Name,
		&campaign.Archived, &campaign.CreatedAt, &campaign.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Add duplicator as owner (no other members copied)
	_, err = tx.Exec(ctx, `
		INSERT INTO campaign_members (campaign_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, campaign.ID, userID)
	if err != nil {
		return nil, err
	}

	// Copy task lists
	listRows, err := tx.Query(ctx, `
		SELECT id, name, color, position FROM task_lists
		WHERE campaign_id = $1 ORDER BY position
	`, sourceCampaignID)
	if err != nil {
		return nil, err
	}
	defer listRows.Close()

	type listMapping struct {
		oldID string
		newID string
	}
	var listMappings []listMapping

	for listRows.Next() {
		var oldID, name, color string
		var position int
		if err := listRows.Scan(&oldID, &name, &color, &position); err != nil {
			return nil, err
		}
		var newID string
		err = tx.QueryRow(ctx, `
			INSERT INTO task_lists (campaign_id, name, color, position)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, campaign.ID, name, color, position).Scan(&newID)
		if err != nil {
			return nil, err
		}
		listMappings = append(listMappings, listMapping{oldID, newID})
	}
	if err := listRows.Err(); err != nil {
		return nil, err
	}

	// Copy groups and tasks for each list
	for _, lm := range listMappings {
		if err := s.duplicateGroupsForList(ctx, tx, lm.oldID, lm.newID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (s *Store) duplicateGroupsForList(ctx context.Context, tx pgx.Tx, oldListID, newListID string) error {
	groupRows, err := tx.Query(ctx, `
		SELECT id, name, position, collapsed FROM task_groups
		WHERE task_list_id = $1 ORDER BY position
	`, oldListID)
	if err != nil {
		return err
	}
	defer groupRows.Close()

	type groupMapping struct {
		oldID string
		newID string
	}
	var groupMappings []groupMapping

	for groupRows.Next() {
		var oldID, name string
		var position int
		var collapsed bool
		if err := groupRows.Scan(&oldID, &name, &position, &collapsed); err != nil {
			return err
		}
		var newID string
		err = tx.QueryRow(ctx, `
			INSERT INTO task_groups (task_list_id, name, position, collapsed)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, newListID, name, position, collapsed).Scan(&newID)
		if err != nil {
			return err
		}
		groupMappings = append(groupMappings, groupMapping{oldID, newID})
	}
	if err := groupRows.Err(); err != nil {
		return err
	}

	for _, gm := range groupMappings {
		if err := s.duplicateTasksForGroup(ctx, tx, gm.oldID, gm.newID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) duplicateTasksForGroup(ctx context.Context, tx pgx.Tx, oldGroupID, newGroupID string) error {
	taskRows, err := tx.Query(ctx, `
		SELECT id, name, description, status, due_date, position
		FROM tasks WHERE task_group_id = $1 ORDER BY position
	`, oldGroupID)
	if err != nil {
		return err
	}
	defer taskRows.Close()

	type taskMapping struct {
		oldID string
		newID string
	}
	var taskMappings []taskMapping

	for taskRows.Next() {
		var t model.Task
		if err := taskRows.Scan(&t.ID, &t.Name, &t.Description, &t.Status, &t.DueDate, &t.Position); err != nil {
			return err
		}
		var newID string
		err = tx.QueryRow(ctx, `
			INSERT INTO tasks (task_group_id, name, description, status, due_date, position)
			VALUES ($1, $2, $3, 'todo', NULL, $4)
			RETURNING id
		`, newGroupID, t.Name, t.Description, t.Position).Scan(&newID)
		if err != nil {
			return err
		}
		taskMappings = append(taskMappings, taskMapping{t.ID, newID})
	}
	if err := taskRows.Err(); err != nil {
		return err
	}

	// Copy subtasks
	for _, tm := range taskMappings {
		_, err := tx.Exec(ctx, `
			INSERT INTO subtasks (task_id, name, is_complete, position)
			SELECT $1, name, false, position
			FROM subtasks WHERE task_id = $2
		`, tm.newID, tm.oldID)
		if err != nil {
			return err
		}
	}
	return nil
}

// populateTemplate is a no-op stub. Replaced with real implementation in Task 7.
func (s *Store) populateTemplate(ctx context.Context, tx pgx.Tx, campaignID string) error {
	return nil
}
```

- [ ] **Step 5: Run tests (requires DATABASE_URL)**

```bash
cd backend
go test ./internal/store/ -v -count=1 -tags=integration -run TestCreateCampaign
```

Expected: PASS (if DATABASE_URL is set and schema is applied).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/store/ backend/internal/model/
git commit -m "feat: add user and campaign store with CRUD queries"
```

---

## Task 4: Task + Subtask Store

**Files:**
- Create: `backend/internal/store/task.go`
- Create: `backend/internal/store/task_test.go`

- [ ] **Step 1: Write failing task store tests**

Create `backend/internal/store/task_test.go`:

```go
//go:build integration

package store

import (
	"context"
	"testing"
)

// createTestGroup creates a campaign with a manually inserted task list and group for testing.
// This avoids depending on the template stub (real template added in Task 7).
func createTestGroup(t *testing.T, s *Store, userID string) (campaignID, groupID string) {
	t.Helper()
	ctx := context.Background()
	campaign, err := s.CreateCampaign(ctx, userID, "Test Campaign")
	if err != nil {
		t.Fatalf("failed to create campaign: %v", err)
	}
	var listID string
	err = s.pool.QueryRow(ctx, `
		INSERT INTO task_lists (campaign_id, name, color, position) VALUES ($1, 'Test List', '#3B82F6', 100) RETURNING id
	`, campaign.ID).Scan(&listID)
	if err != nil {
		t.Fatalf("failed to create task list: %v", err)
	}
	var gID string
	err = s.pool.QueryRow(ctx, `
		INSERT INTO task_groups (task_list_id, name, position) VALUES ($1, 'Test Group', 100) RETURNING id
	`, listID).Scan(&gID)
	if err != nil {
		t.Fatalf("failed to create task group: %v", err)
	}
	return campaign.ID, gID
}

func TestCreateTask(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)
	campaignID, groupID := createTestGroup(t, s, user.ID)
	defer s.DeleteCampaign(context.Background(), campaignID)

	task, err := s.CreateTask(context.Background(), groupID, "New Task")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	if task.Name != "New Task" {
		t.Errorf("expected name 'New Task', got '%s'", task.Name)
	}
	if task.Status != "todo" {
		t.Errorf("expected status 'todo', got '%s'", task.Status)
	}
}

func TestUpdateTask(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)
	campaignID, groupID := createTestGroup(t, s, user.ID)
	defer s.DeleteCampaign(context.Background(), campaignID)

	task, _ := s.CreateTask(context.Background(), groupID, "To Update")
	updates := TaskUpdate{Status: strPtr("done")}
	updated, err := s.UpdateTask(context.Background(), task.ID, updates)
	if err != nil {
		t.Fatalf("failed to update task: %v", err)
	}
	if updated.Status != "done" {
		t.Errorf("expected status 'done', got '%s'", updated.Status)
	}
}

func TestSubtasks(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)
	campaignID, groupID := createTestGroup(t, s, user.ID)
	defer s.DeleteCampaign(context.Background(), campaignID)

	task, _ := s.CreateTask(context.Background(), groupID, "With Subtasks")
	taskID := task.ID

	subtask, err := s.CreateSubtask(context.Background(), taskID, "Sub 1")
	if err != nil {
		t.Fatalf("failed to create subtask: %v", err)
	}
	if subtask.Name != "Sub 1" {
		t.Errorf("expected 'Sub 1', got '%s'", subtask.Name)
	}

	// Toggle complete
	updated, err := s.UpdateSubtask(context.Background(), subtask.ID, nil, boolPtr(true))
	if err != nil {
		t.Fatalf("failed to toggle subtask: %v", err)
	}
	if !updated.IsComplete {
		t.Error("expected subtask to be complete")
	}
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend
go test ./internal/store/ -v -count=1 -tags=integration -run TestCreateTask
```

Expected: FAIL — functions don't exist.

- [ ] **Step 3: Implement task + subtask store**

Create `backend/internal/store/task.go`:

```go
package store

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/rdelpret/music-release-planner/backend/internal/model"
)

type TaskUpdate struct {
	Name        *string          `json:"name,omitempty"`
	Description *json.RawMessage `json:"description,omitempty"`
	Status      *string          `json:"status,omitempty"`
	DueDate     *string          `json:"due_date,omitempty"`
}

func (s *Store) CreateTask(ctx context.Context, groupID, name string) (*model.Task, error) {
	// Get next position
	var maxPos *int
	s.pool.QueryRow(ctx, `
		SELECT MAX(position) FROM tasks WHERE task_group_id = $1
	`, groupID).Scan(&maxPos)

	nextPos := 100
	if maxPos != nil {
		nextPos = *maxPos + 100
	}

	var task model.Task
	err := s.pool.QueryRow(ctx, `
		INSERT INTO tasks (task_group_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, task_group_id, name, description, status, due_date, position, created_at, updated_at
	`, groupID, name, nextPos).Scan(&task.ID, &task.TaskGroupID, &task.Name, &task.Description,
		&task.Status, &task.DueDate, &task.Position, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) UpdateTask(ctx context.Context, taskID string, updates TaskUpdate) (*model.Task, error) {
	// Build dynamic update query
	setClauses := []string{"updated_at = now()"}
	args := []interface{}{}
	argN := 1

	if updates.Name != nil {
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argN))
		args = append(args, *updates.Name)
		argN++
	}
	if updates.Description != nil {
		setClauses = append(setClauses, "description = $"+strconv.Itoa(argN))
		args = append(args, updates.Description)
		argN++
	}
	if updates.Status != nil {
		setClauses = append(setClauses, "status = $"+strconv.Itoa(argN))
		args = append(args, *updates.Status)
		argN++
	}
	if updates.DueDate != nil {
		if *updates.DueDate == "" {
			setClauses = append(setClauses, "due_date = NULL")
		} else {
			setClauses = append(setClauses, "due_date = $"+strconv.Itoa(argN))
			args = append(args, *updates.DueDate)
			argN++
		}
	}

	args = append(args, taskID)

	query := "UPDATE tasks SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE id = $" + strconv.Itoa(argN)
	query += " RETURNING id, task_group_id, name, description, status, due_date, position, created_at, updated_at"

	var task model.Task
	err := s.pool.QueryRow(ctx, query, args...).Scan(&task.ID, &task.TaskGroupID, &task.Name,
		&task.Description, &task.Status, &task.DueDate, &task.Position, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) DeleteTask(ctx context.Context, taskID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, taskID)
	return err
}

func (s *Store) CreateSubtask(ctx context.Context, taskID, name string) (*model.Subtask, error) {
	var maxPos *int
	s.pool.QueryRow(ctx, `
		SELECT MAX(position) FROM subtasks WHERE task_id = $1
	`, taskID).Scan(&maxPos)

	nextPos := 100
	if maxPos != nil {
		nextPos = *maxPos + 100
	}

	var subtask model.Subtask
	err := s.pool.QueryRow(ctx, `
		INSERT INTO subtasks (task_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, task_id, name, is_complete, position
	`, taskID, name, nextPos).Scan(&subtask.ID, &subtask.TaskID, &subtask.Name, &subtask.IsComplete, &subtask.Position)
	if err != nil {
		return nil, err
	}
	return &subtask, nil
}

func (s *Store) UpdateSubtask(ctx context.Context, subtaskID string, name *string, isComplete *bool) (*model.Subtask, error) {
	setClauses := []string{}
	args := []interface{}{}
	argN := 1

	if name != nil {
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argN))
		args = append(args, *name)
		argN++
	}
	if isComplete != nil {
		setClauses = append(setClauses, "is_complete = $"+strconv.Itoa(argN))
		args = append(args, *isComplete)
		argN++
	}

	if len(setClauses) == 0 {
		// Nothing to update, just return current
		return s.getSubtask(ctx, subtaskID)
	}

	args = append(args, subtaskID)

	query := "UPDATE subtasks SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE id = $" + strconv.Itoa(argN)
	query += " RETURNING id, task_id, name, is_complete, position"

	var subtask model.Subtask
	err := s.pool.QueryRow(ctx, query, args...).Scan(&subtask.ID, &subtask.TaskID, &subtask.Name, &subtask.IsComplete, &subtask.Position)
	if err != nil {
		return nil, err
	}
	return &subtask, nil
}

func (s *Store) DeleteSubtask(ctx context.Context, subtaskID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM subtasks WHERE id = $1`, subtaskID)
	return err
}

func (s *Store) getSubtask(ctx context.Context, subtaskID string) (*model.Subtask, error) {
	var subtask model.Subtask
	err := s.pool.QueryRow(ctx, `
		SELECT id, task_id, name, is_complete, position
		FROM subtasks WHERE id = $1
	`, subtaskID).Scan(&subtask.ID, &subtask.TaskID, &subtask.Name, &subtask.IsComplete, &subtask.Position)
	if err != nil {
		return nil, err
	}
	return &subtask, nil
}

// GetTasksByDueDate fetches tasks with due dates for calendar view
func (s *Store) GetTasksByDueDate(ctx context.Context, campaignID string) ([]model.Task, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, t.task_group_id, t.name, t.description, t.status, t.due_date, t.position, t.created_at, t.updated_at
		FROM tasks t
		JOIN task_groups tg ON tg.id = t.task_group_id
		JOIN task_lists tl ON tl.id = tg.task_list_id
		WHERE tl.campaign_id = $1 AND t.due_date IS NOT NULL
		ORDER BY t.due_date
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.TaskGroupID, &t.Name, &t.Description, &t.Status,
			&t.DueDate, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend
go test ./internal/store/ -v -count=1 -tags=integration -run "TestCreateTask|TestUpdateTask|TestSubtasks"
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/store/task.go backend/internal/store/task_test.go
git commit -m "feat: add task and subtask store with CRUD queries"
```

---

## Task 5: Reorder Store

**Files:**
- Create: `backend/internal/store/reorder.go`
- Create: `backend/internal/store/reorder_test.go`

- [ ] **Step 1: Write failing reorder tests**

Create `backend/internal/store/reorder_test.go`:

```go
//go:build integration

package store

import (
	"context"
	"testing"
)

func TestReorderTask(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)
	campaign, _ := s.CreateCampaign(context.Background(), user.ID, "Reorder Test")
	defer s.DeleteCampaign(context.Background(), campaign.ID)

	full, _ := s.GetFullCampaign(context.Background(), campaign.ID)
	groupID := full.TaskLists[0].TaskGroups[0].ID
	tasks := full.TaskLists[0].TaskGroups[0].Tasks

	if len(tasks) < 2 {
		t.Fatal("need at least 2 tasks for reorder test")
	}

	// Move second task to position before first (within same group)
	err := s.ReorderTask(context.Background(), tasks[1].ID, groupID, tasks[0].Position-50)
	if err != nil {
		t.Fatalf("failed to reorder task: %v", err)
	}

	// Verify new order
	refreshed, _ := s.GetFullCampaign(context.Background(), campaign.ID)
	firstTask := refreshed.TaskLists[0].TaskGroups[0].Tasks[0]
	if firstTask.ID != tasks[1].ID {
		t.Errorf("expected task %s to be first, got %s", tasks[1].ID, firstTask.ID)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend
go test ./internal/store/ -v -count=1 -tags=integration -run TestReorderTask
```

Expected: FAIL.

- [ ] **Step 3: Implement reorder store**

Create `backend/internal/store/reorder.go`:

```go
package store

import (
	"context"
)

// ReorderTask moves a task to a new position, optionally in a different group
func (s *Store) ReorderTask(ctx context.Context, taskID, targetGroupID string, newPosition int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE tasks SET task_group_id = $2, position = $3, updated_at = now()
		WHERE id = $1
	`, taskID, targetGroupID, newPosition)
	return err
}

// ReorderTaskList moves a task list to a new position
func (s *Store) ReorderTaskList(ctx context.Context, listID string, newPosition int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE task_lists SET position = $2
		WHERE id = $1
	`, listID, newPosition)
	return err
}

// ReorderTaskGroup moves a task group to a new position
func (s *Store) ReorderTaskGroup(ctx context.Context, groupID string, newPosition int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE task_groups SET position = $2
		WHERE id = $1
	`, groupID, newPosition)
	return err
}
```

- [ ] **Step 4: Run test**

```bash
cd backend
go test ./internal/store/ -v -count=1 -tags=integration -run TestReorderTask
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/store/reorder.go backend/internal/store/reorder_test.go
git commit -m "feat: add reorder store for tasks, groups, and lists"
```

---

## Task 6: Google OAuth + Auth Middleware

**Files:**
- Create: `backend/internal/auth/auth.go`
- Create: `backend/internal/auth/auth_test.go`

- [ ] **Step 1: Write failing auth middleware tests**

Create `backend/internal/auth/auth_test.go`:

```go
package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuth_NoSession(t *testing.T) {
	initTestStore()

	handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/campaigns", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_WithValidSession(t *testing.T) {
	initTestStore()

	handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := GetUserEmail(r)
		if email != "test@subwave.music" {
			t.Errorf("expected email 'test@subwave.music', got '%s'", email)
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Create request with session
	req := httptest.NewRequest("GET", "/api/campaigns", nil)
	rr := httptest.NewRecorder()

	session, _ := sessionStore.Get(req, sessionName)
	session.Values["user_email"] = "test@subwave.music"
	session.Values["user_id"] = "test-uuid"
	session.Save(req, rr)

	// Make new request with session cookie
	req2 := httptest.NewRequest("GET", "/api/campaigns", nil)
	for _, cookie := range rr.Result().Cookies() {
		req2.AddCookie(cookie)
	}
	rr2 := httptest.NewRecorder()

	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr2.Code)
	}
}

func initTestStore() {
	sessionStore = newCookieStore("test-secret-key-for-testing-only")
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend
go test ./internal/auth/ -v -count=1 -run TestRequireAuth
```

Expected: FAIL.

- [ ] **Step 3: Implement auth**

Create `backend/internal/auth/auth.go`:

```go
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const sessionName = "mrp-session"

var (
	oauthConfig  *oauth2.Config
	sessionStore *sessions.CookieStore
	allowedEmails map[string]bool
)

type contextKey string

const (
	ctxUserEmail contextKey = "user_email"
	ctxUserID    contextKey = "user_id"
)

func Initialize() {
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	if oauthConfig.RedirectURL == "" {
		oauthConfig.RedirectURL = "http://localhost:8080/auth/google/callback"
	}

	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		log.Fatal("SESSION_SECRET is required")
	}
	sessionStore = newCookieStore(secret)

	// Parse whitelist from comma-separated env var
	allowedEmails = make(map[string]bool)
	emails := os.Getenv("ALLOWED_EMAILS")
	for _, email := range strings.Split(emails, ",") {
		email = strings.TrimSpace(email)
		if email != "" {
			allowedEmails[email] = true
		}
	}
}

func newCookieStore(secret string) *sessions.CookieStore {
	hash := sha256.Sum256([]byte(secret))
	store := sessions.NewCookieStore(hash[:])
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HttpOnly: true,
		Secure:   os.Getenv("ENV") != "development",
		SameSite: http.SameSiteLaxMode,
	}
	return store
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	session, _ := sessionStore.Get(r, sessionName)
	session.Values["oauth_state"] = state
	session.Save(r, w)

	url := oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// OnAuthSuccess is called by the server to upsert the user after successful OAuth.
// Returns the user_id to store in the session. This keeps auth package decoupled from store.
type UserUpsertFunc func(ctx context.Context, email, name string, avatarURL *string) (userID string, err error)

var userUpsertFn UserUpsertFunc

func SetUserUpsertFunc(fn UserUpsertFunc) {
	userUpsertFn = fn
}

func HandleCallbackWithUpsert(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, sessionName)
	expectedState, _ := session.Values["oauth_state"].(string)
	if r.URL.Query().Get("state") != expectedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	delete(session.Values, "oauth_state")

	code := r.URL.Query().Get("code")
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("OAuth exchange error: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	if !allowedEmails[userInfo.Email] {
		http.Error(w, "Access denied — this tool is for Subwave team members only", http.StatusForbidden)
		return
	}

	// Upsert user in DB
	var avatarURL *string
	if userInfo.Picture != "" {
		avatarURL = &userInfo.Picture
	}

	userID := ""
	if userUpsertFn != nil {
		userID, err = userUpsertFn(r.Context(), userInfo.Email, userInfo.Name, avatarURL)
		if err != nil {
			log.Printf("Failed to upsert user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	session.Values["user_email"] = userInfo.Email
	session.Values["user_id"] = userID
	session.Save(r, w)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	http.Redirect(w, r, frontendURL+"/dashboard", http.StatusTemporaryRedirect)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, sessionName)
	session.Options.MaxAge = -1
	session.Save(r, w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, sessionName)
		email, ok := session.Values["user_email"].(string)
		if !ok || email == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "not authenticated"})
			return
		}

		userID, _ := session.Values["user_id"].(string)
		ctx := context.WithValue(r.Context(), ctxUserEmail, email)
		ctx = context.WithValue(ctx, ctxUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserEmail(r *http.Request) string {
	email, _ := r.Context().Value(ctxUserEmail).(string)
	return email
}

func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(ctxUserID).(string)
	return id
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, sessionName)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"email":      session.Values["user_email"],
		"user_id":    session.Values["user_id"],
	})
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
```

- [ ] **Step 4: Run auth tests**

```bash
cd backend
go get github.com/gorilla/sessions
go get golang.org/x/oauth2
go mod tidy
go test ./internal/auth/ -v -count=1 -run TestRequireAuth
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/auth/
git commit -m "feat: add Google OAuth flow and session-based auth middleware"
```

---

## Task 7: Default Campaign Template

**Files:**
- Create: `backend/internal/template/template.go`
- Modify: `backend/internal/store/campaign.go` (wire up `populateTemplate`)

- [ ] **Step 1: Create template data**

Create `backend/internal/template/template.go`:

```go
package template

// TemplateList defines a task list with its groups and tasks
type TemplateList struct {
	Name   string
	Color  string
	Groups []TemplateGroup
}

type TemplateGroup struct {
	Name  string
	Tasks []TemplateTask
}

type TemplateTask struct {
	Name     string
	Subtasks []string
}

// DefaultTemplate returns the 5 pre-loaded task lists for a new campaign
func DefaultTemplate() []TemplateList {
	return []TemplateList{
		{
			Name:  "Campaign Assets",
			Color: "#3B82F6",
			Groups: []TemplateGroup{
				{
					Name: "Song Assets",
					Tasks: []TemplateTask{
						{Name: "Make New Asset Folder"},
						{Name: "Release Date"},
						{Name: "Final Master"},
						{Name: "Final Artwork"},
						{Name: "Instrumental"},
						{Name: "Stems"},
						{Name: "Lyrics"},
						{Name: "Press Photos"},
						{Name: "Music Video/Lyric Video"},
						{Name: "Folder for Promo Content"},
					},
				},
			},
		},
		{
			Name:  "Campaign Tasklist",
			Color: "#22C55E",
			Groups: []TemplateGroup{
				{
					Name: "Pre-Release",
					Tasks: []TemplateTask{
						{Name: "Campaign Setup + Timeline"},
						{Name: "Contracts"},
						{Name: "Photoshoot for release"},
						{Name: "Make content plan"},
						{Name: "Create content or work with a graphic designer"},
						{Name: "Create pre-save Link"},
						{Name: "Pre-save campaign/contest"},
						{Name: "Create linktree"},
						{Name: "Write out hashtags"},
						{Name: "Update website - coming soon"},
						{Name: "Update all SM profiles - coming soon"},
						{Name: "Update streaming profiles"},
					},
				},
			},
		},
		{
			Name:  "Social Media Content",
			Color: "#A78BFA",
			Groups: []TemplateGroup{
				{
					Name: "Pre-Release Content",
					Tasks: []TemplateTask{
						{Name: "Teaser post"},
						{Name: "Banner - release date - presave now"},
						{Name: "Single art - story sized - release date or 'pre-save now'"},
						{Name: "Single art edit #1 (with release date)"},
						{Name: "Pre-save Post"},
						{Name: "Song poster #1 - storytelling"},
						{Name: "Song teaser #1"},
						{Name: "\"Midnight\" or \"Tomorrow\""},
					},
				},
				{
					Name: "Post-Release Content",
					Tasks: []TemplateTask{
						{Name: "Banner - out now"},
						{Name: "Song Clip - out now"},
						{Name: "Single art edit #2 (\"out now\")"},
						{Name: "Single art - out now - story sized"},
						{Name: "Song poster #2"},
						{Name: "Studio content #1 - storytelling"},
						{Name: "Single art edit #3 - storytelling"},
						{Name: "Credits page"},
						{Name: "Lyric poster/Lyric Reel"},
					},
				},
				{
					Name: "Extra Content",
					Tasks: []TemplateTask{
						{Name: "Pre-save Post #2"},
						{Name: "Song clip/teaser #3"},
						{Name: "Song poster #3"},
						{Name: "Studio content #2"},
						{Name: "Tiktok/Reels"},
						{Name: "Clips from Music video/lyric video"},
						{Name: "Countdowns"},
						{Name: "Acoustic version/demo version/live version"},
					},
				},
			},
		},
		{
			Name:  "PR",
			Color: "#EAB308",
			Groups: []TemplateGroup{
				{
					Name: "Media Outreach",
					Tasks: []TemplateTask{
						{Name: "Research new outlets & develop CRM"},
						{Name: "Create press release"},
						{Name: "Create e-mail outreach pitch"},
						{Name: "Traditional Media outreach"},
						{Name: "New Media outreach"},
						{Name: "Submit to third party platforms"},
						{Name: "Press follow ups"},
					},
				},
				{
					Name: "Playlisting",
					Tasks: []TemplateTask{
						{Name: "Research playlists"},
						{Name: "Write playlist pitch"},
						{Name: "Pitch to independent curators"},
						{Name: "Submit to third party platforms"},
					},
				},
				{
					Name: "Community Engagement",
					Tasks: []TemplateTask{
						{Name: "Send DMs on IG/Twitter/FB with link"},
						{Name: "Post on discord chats/reddit threads"},
						{Name: "Send mailing list e-mail - presave"},
						{Name: "Send mailing list e-mail - out now"},
					},
				},
			},
		},
		{
			Name:  "Distribution",
			Color: "#EF4444",
			Groups: []TemplateGroup{
				{
					Name: "To Do",
					Tasks: []TemplateTask{
						{Name: "Obtain license for cover song"},
						{Name: "Upload to distributor", Subtasks: []string{"Fill out credits", "Upload lyrics"}},
						{Name: "Pitch song to Spotify editors"},
						{Name: "Upload Spotify Canvas"},
						{Name: "Register song with PROs"},
						{Name: "For physical releases"},
					},
				},
			},
		},
	}
}
```

- [ ] **Step 2: Add populateTemplate to store**

Add to `backend/internal/store/campaign.go` — replace the placeholder `populateTemplate` with the real implementation. Add this import and method:

```go
// Add to imports:
import "github.com/rdelpret/music-release-planner/backend/internal/template"

// Add to imports (ensure pgx.Tx is available):
import "github.com/jackc/pgx/v5"

func (s *Store) populateTemplate(ctx context.Context, tx pgx.Tx, campaignID string) error {
	tmpl := template.DefaultTemplate()

	for listPos, list := range tmpl {
		var listID string
		err := tx.QueryRow(ctx, `
			INSERT INTO task_lists (campaign_id, name, color, position)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, campaignID, list.Name, list.Color, (listPos+1)*100).Scan(&listID)
		if err != nil {
			return fmt.Errorf("inserting task list %s: %w", list.Name, err)
		}

		for groupPos, group := range list.Groups {
			var groupID string
			err := tx.QueryRow(ctx, `
				INSERT INTO task_groups (task_list_id, name, position)
				VALUES ($1, $2, $3)
				RETURNING id
			`, listID, group.Name, (groupPos+1)*100).Scan(&groupID)
			if err != nil {
				return fmt.Errorf("inserting task group %s: %w", group.Name, err)
			}

			for taskPos, task := range group.Tasks {
				var taskID string
				err := tx.QueryRow(ctx, `
					INSERT INTO tasks (task_group_id, name, position)
					VALUES ($1, $2, $3)
					RETURNING id
				`, groupID, task.Name, (taskPos+1)*100).Scan(&taskID)
				if err != nil {
					return fmt.Errorf("inserting task %s: %w", task.Name, err)
				}

				for subPos, subtask := range task.Subtasks {
					_, err := tx.Exec(ctx, `
						INSERT INTO subtasks (task_id, name, position)
						VALUES ($1, $2, $3)
					`, taskID, subtask, (subPos+1)*100)
					if err != nil {
						return fmt.Errorf("inserting subtask %s: %w", subtask, err)
					}
				}
			}
		}
	}
	return nil
}
```

- [ ] **Step 3: Verify build compiles**

```bash
cd backend
go build ./...
```

Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/template/ backend/internal/store/campaign.go
git commit -m "feat: add default campaign template with 5 task lists"
```

---

## Task 8: Campaign + Task HTTP Handlers

**Files:**
- Create: `backend/internal/handler/campaign.go`
- Create: `backend/internal/handler/campaign_test.go`
- Create: `backend/internal/handler/task.go`
- Create: `backend/internal/handler/task_test.go`
- Modify: `backend/internal/handler/handler.go` (add Store dependency)
- Modify: `backend/internal/handler/routes.go` (register all routes)

- [ ] **Step 1: Update Server struct to accept Store**

Modify `backend/internal/handler/handler.go` — add Store field:

```go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rdelpret/music-release-planner/backend/internal/store"
)

type Server struct {
	router chi.Router
	store  *store.Store
}

func NewServer(s *store.Store) *Server {
	srv := &Server{store: s}
	srv.router = srv.routes()
	return srv
}

func (s *Server) Start(port string) error {
	return http.ListenAndServe(fmt.Sprintf(":%s", port), s.router)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
```

- [ ] **Step 2: Register all routes**

Update `backend/internal/handler/routes.go`:

```go
package handler

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rdelpret/music-release-planner/backend/internal/auth"
)

func (s *Server) routes() chi.Router {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(corsMiddleware)

	// Auth routes (public)
	r.Get("/auth/google", auth.HandleLogin)
	r.Get("/auth/google/callback", auth.HandleCallbackWithUpsert)
	r.Post("/auth/logout", auth.HandleLogout)

	// Protected API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(auth.RequireAuth)

		r.Get("/me", auth.HandleMe)

		// Campaigns
		r.Get("/campaigns", s.handleListCampaigns)
		r.Post("/campaigns", s.handleCreateCampaign)
		r.Get("/campaigns/{id}", s.handleGetCampaign)
		r.Post("/campaigns/{id}/duplicate", s.handleDuplicateCampaign)
		r.Patch("/campaigns/{id}/archive", s.handleArchiveCampaign)
		r.Delete("/campaigns/{id}", s.handleDeleteCampaign)

		// Tasks
		r.Post("/task-groups/{id}/tasks", s.handleCreateTask)
		r.Patch("/tasks/{id}", s.handleUpdateTask)
		r.Delete("/tasks/{id}", s.handleDeleteTask)
		r.Patch("/tasks/{id}/reorder", s.handleReorderTask)

		// Task lists & groups reorder
		r.Patch("/task-lists/{id}/reorder", s.handleReorderTaskList)
		r.Patch("/task-groups/{id}/reorder", s.handleReorderTaskGroup)

		// Subtasks
		r.Post("/tasks/{id}/subtasks", s.handleCreateSubtask)
		r.Patch("/subtasks/{id}", s.handleUpdateSubtask)
		r.Delete("/subtasks/{id}", s.handleDeleteSubtask)
	})

	// Health check (public)
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigin := os.Getenv("FRONTEND_URL")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == allowedOrigin || origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 3: Implement campaign handlers**

Create `backend/internal/handler/campaign.go`:

```go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rdelpret/music-release-planner/backend/internal/auth"
	"github.com/rdelpret/music-release-planner/backend/internal/model"
)

func (s *Server) handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	campaigns, err := s.store.ListCampaigns(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list campaigns")
		return
	}
	if campaigns == nil {
		campaigns = []model.Campaign{}
	}
	writeJSON(w, http.StatusOK, campaigns)
}

func (s *Server) handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	campaign, err := s.store.CreateCampaign(r.Context(), userID, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create campaign")
		return
	}
	writeJSON(w, http.StatusCreated, campaign)
}

func (s *Server) handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	campaign, err := s.store.GetFullCampaign(r.Context(), campaignID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Campaign not found")
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (s *Server) handleDuplicateCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	userID := auth.GetUserID(r)

	campaign, err := s.store.DuplicateCampaign(r.Context(), campaignID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to duplicate campaign")
		return
	}
	writeJSON(w, http.StatusCreated, campaign)
}

func (s *Server) handleArchiveCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")

	var req struct {
		Archived bool `json:"archived"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.store.ArchiveCampaign(r.Context(), campaignID, req.Archived); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to archive campaign")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeleteCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	if err := s.store.DeleteCampaign(r.Context(), campaignID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete campaign")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
```

- [ ] **Step 4: Implement task handlers**

Create `backend/internal/handler/task.go`:

```go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rdelpret/music-release-planner/backend/internal/store"
)

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	task, err := s.store.CreateTask(r.Context(), groupID, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	var updates store.TaskUpdate
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	task, err := s.store.UpdateTask(r.Context(), taskID, updates)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update task")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if err := s.store.DeleteTask(r.Context(), taskID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleCreateSubtask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	subtask, err := s.store.CreateSubtask(r.Context(), taskID, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create subtask")
		return
	}
	writeJSON(w, http.StatusCreated, subtask)
}

func (s *Server) handleUpdateSubtask(w http.ResponseWriter, r *http.Request) {
	subtaskID := chi.URLParam(r, "id")

	var req struct {
		Name       *string `json:"name,omitempty"`
		IsComplete *bool   `json:"is_complete,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	subtask, err := s.store.UpdateSubtask(r.Context(), subtaskID, req.Name, req.IsComplete)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update subtask")
		return
	}
	writeJSON(w, http.StatusOK, subtask)
}

func (s *Server) handleDeleteSubtask(w http.ResponseWriter, r *http.Request) {
	subtaskID := chi.URLParam(r, "id")
	if err := s.store.DeleteSubtask(r.Context(), subtaskID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete subtask")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Reorder handlers
func (s *Server) handleReorderTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	var req struct {
		TargetGroupID string `json:"target_group_id"`
		Position      int    `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.store.ReorderTask(r.Context(), taskID, req.TargetGroupID, req.Position); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to reorder task")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReorderTaskList(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "id")

	var req struct {
		Position int `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.store.ReorderTaskList(r.Context(), listID, req.Position); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to reorder task list")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReorderTaskGroup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	var req struct {
		Position int `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.store.ReorderTaskGroup(r.Context(), groupID, req.Position); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to reorder task group")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

- [ ] **Step 5: Update main.go to wire everything together**

Update `backend/cmd/server/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rdelpret/music-release-planner/backend/internal/auth"
	"github.com/rdelpret/music-release-planner/backend/internal/handler"
	"github.com/rdelpret/music-release-planner/backend/internal/store"
)

func main() {
	godotenv.Load()
	godotenv.Load("../.env")

	// Initialize auth
	auth.Initialize()

	// Initialize database
	db, err := store.New()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Wire up user upsert for OAuth callback
	auth.SetUserUpsertFunc(func(ctx context.Context, email, name string, avatarURL *string) (string, error) {
		user, err := db.UpsertUser(ctx, email, name, avatarURL)
		if err != nil {
			return "", err
		}
		return user.ID, nil
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := handler.NewServer(db)
	fmt.Printf("Server running at http://localhost:%s\n", port)
	if err := srv.Start(port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

- [ ] **Step 6: Verify build**

```bash
cd backend
go mod tidy
go build ./...
```

Expected: Build succeeds.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/ backend/cmd/server/main.go
git commit -m "feat: add campaign and task HTTP handlers with full route registration"
```

---

## Task 9: Frontend Scaffolding + Subwave Theme

**Files:**
- Create: `frontend/` (Next.js project)
- Create: `frontend/src/app/globals.css` (Subwave theme)
- Create: `frontend/next.config.ts` (API proxy)
- Create: `frontend/src/lib/types.ts`
- Create: `frontend/src/lib/api.ts`

- [ ] **Step 1: Create Next.js project**

```bash
cd /Users/robbiedelprete/github.com/rdelpret/music-release-planner
npx create-next-app@latest frontend --typescript --tailwind --eslint --app --src-dir --no-import-alias --use-npm
```

Accept defaults. This creates the base Next.js + TailwindCSS setup.

- [ ] **Step 2: Install additional dependencies**

```bash
cd frontend
npm install @tanstack/react-query sonner lucide-react clsx tailwind-merge class-variance-authority
npm install -D shadcn
```

- [ ] **Step 3: Initialize shadcn/ui**

```bash
cd frontend
npx shadcn@latest init --style new-york --base-color neutral --css-variables
```

- [ ] **Step 4: Add required shadcn components**

```bash
cd frontend
npx shadcn@latest add button card tabs dialog sheet input textarea badge toast
```

- [ ] **Step 5: Configure Subwave theme**

Replace `frontend/src/app/globals.css` with:

```css
@import "tailwindcss";
@import "tw-animate-css";
@import "shadcn/tailwind.css";

@custom-variant dark (&:is(.dark *));

@theme inline {
  --color-bg-base: #0B0D10;
  --color-bg-surface: #15121f;
  --color-text-primary: #E8EEF6;
  --color-text-muted: #6B7280;
  --color-accent: #B7FF4A;
  --color-accent-dark: #8FCC3A;
  --color-secondary: #A78BFA;
  --color-secondary-dark: #8B6DF5;
  --color-accent-glow: rgba(183, 255, 74, 0.4);

  --font-heading: "Space Grotesk", sans-serif;
  --font-body: "Inter", sans-serif;
}

:root {
  --radius: 0.625rem;
  --background: #0B0D10;
  --foreground: #E8EEF6;
  --card: #15121f;
  --card-foreground: #E8EEF6;
  --popover: #15121f;
  --popover-foreground: #E8EEF6;
  --primary: #B7FF4A;
  --primary-foreground: #0B0D10;
  --secondary: #A78BFA;
  --secondary-foreground: #E8EEF6;
  --muted: #1a1a2e;
  --muted-foreground: #6B7280;
  --accent: #1a1a2e;
  --accent-foreground: #E8EEF6;
  --destructive: #EF4444;
  --destructive-foreground: #E8EEF6;
  --border: rgba(255, 255, 255, 0.08);
  --input: rgba(255, 255, 255, 0.08);
  --ring: #B7FF4A;
}

@layer base {
  * {
    @apply border-border outline-ring/50;
  }
  body {
    @apply bg-bg-base text-text-primary font-body;
  }
  h1, h2, h3, h4, h5, h6 {
    font-family: var(--font-heading);
  }
}

/* Glassmorphism utility */
.glass {
  backdrop-filter: blur(30px) saturate(180%);
  background: rgba(21, 18, 31, 0.8);
}

/* Accent glow on hover */
.glow-hover:hover {
  box-shadow: 0 0 20px rgba(183, 255, 74, 0.4);
}

/* Smooth transitions */
.transition-smooth {
  transition: all 200ms cubic-bezier(0.4, 0, 0.2, 1);
}
```

- [ ] **Step 6: Configure API proxy**

Create/update `frontend/next.config.ts`:

```ts
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export", // Static export for Cloudflare Workers (outputs to /out)
};

export default nextConfig;
```

- [ ] **Step 7: Create TypeScript types**

Create `frontend/src/lib/types.ts`:

```ts
export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url?: string;
  created_at: string;
}

export interface Campaign {
  id: string;
  created_by: string;
  name: string;
  archived: boolean;
  created_at: string;
  updated_at: string;
  task_lists?: TaskList[];
}

export interface TaskList {
  id: string;
  campaign_id: string;
  name: string;
  color: string;
  position: number;
  task_groups?: TaskGroup[];
}

export interface TaskGroup {
  id: string;
  task_list_id: string;
  name: string;
  position: number;
  collapsed: boolean;
  tasks?: Task[];
}

export interface Task {
  id: string;
  task_group_id: string;
  name: string;
  description?: Record<string, unknown>;
  status: "todo" | "in_progress" | "done";
  due_date?: string;
  position: number;
  created_at: string;
  updated_at: string;
  subtasks?: Subtask[];
}

export interface Subtask {
  id: string;
  task_id: string;
  name: string;
  is_complete: boolean;
  position: number;
}
```

- [ ] **Step 8: Create API wrapper**

Create `frontend/src/lib/api.ts`:

```ts
import type { Campaign, Task, Subtask } from "./types";

// In dev, points to Go backend on :8080. In production, same origin (Worker proxies /api/*).
const BASE = process.env.NEXT_PUBLIC_API_URL ?? "";

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + url, {
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }
  return res.json();
}

// Auth
export const getMe = () => fetchJSON<{ email: string; user_id: string }>("/api/me");
export const logout = () => fetchJSON<void>("/auth/logout", { method: "POST" });

// Campaigns
export const listCampaigns = () => fetchJSON<Campaign[]>("/api/campaigns");
export const createCampaign = (name: string) =>
  fetchJSON<Campaign>("/api/campaigns", { method: "POST", body: JSON.stringify({ name }) });
export const getCampaign = (id: string) => fetchJSON<Campaign>(`/api/campaigns/${id}`);
export const duplicateCampaign = (id: string) =>
  fetchJSON<Campaign>(`/api/campaigns/${id}/duplicate`, { method: "POST" });
export const archiveCampaign = (id: string, archived: boolean) =>
  fetchJSON<void>(`/api/campaigns/${id}/archive`, { method: "PATCH", body: JSON.stringify({ archived }) });
export const deleteCampaign = (id: string) =>
  fetchJSON<void>(`/api/campaigns/${id}`, { method: "DELETE" });

// Tasks
export const createTask = (groupId: string, name: string) =>
  fetchJSON<Task>(`/api/task-groups/${groupId}/tasks`, { method: "POST", body: JSON.stringify({ name }) });
export const updateTask = (id: string, updates: Partial<Task>) =>
  fetchJSON<Task>(`/api/tasks/${id}`, { method: "PATCH", body: JSON.stringify(updates) });
export const deleteTask = (id: string) =>
  fetchJSON<void>(`/api/tasks/${id}`, { method: "DELETE" });
export const reorderTask = (id: string, targetGroupId: string, position: number) =>
  fetchJSON<void>(`/api/tasks/${id}/reorder`, {
    method: "PATCH",
    body: JSON.stringify({ target_group_id: targetGroupId, position }),
  });

// Reorder
export const reorderTaskList = (id: string, position: number) =>
  fetchJSON<void>(`/api/task-lists/${id}/reorder`, { method: "PATCH", body: JSON.stringify({ position }) });
export const reorderTaskGroup = (id: string, position: number) =>
  fetchJSON<void>(`/api/task-groups/${id}/reorder`, { method: "PATCH", body: JSON.stringify({ position }) });

// Subtasks
export const createSubtask = (taskId: string, name: string) =>
  fetchJSON<Subtask>(`/api/tasks/${taskId}/subtasks`, { method: "POST", body: JSON.stringify({ name }) });
export const updateSubtask = (id: string, updates: { name?: string; is_complete?: boolean }) =>
  fetchJSON<Subtask>(`/api/subtasks/${id}`, { method: "PATCH", body: JSON.stringify(updates) });
export const deleteSubtask = (id: string) =>
  fetchJSON<void>(`/api/subtasks/${id}`, { method: "DELETE" });
```

- [ ] **Step 9: Add Google Fonts to layout**

Update `frontend/src/app/layout.tsx`:

```tsx
import type { Metadata } from "next";
import { Space_Grotesk, Inter } from "next/font/google";
import "./globals.css";

const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  variable: "--font-heading",
});

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-body",
});

export const metadata: Metadata = {
  title: "Subwave Release Planner",
  description: "Internal release campaign management for Subwave",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${spaceGrotesk.variable} ${inter.variable} antialiased min-h-screen`}
      >
        {children}
      </body>
    </html>
  );
}
```

- [ ] **Step 10: Verify frontend builds**

```bash
cd frontend
npm run build
```

Expected: Build succeeds.

- [ ] **Step 11: Commit**

```bash
git add frontend/
git commit -m "feat: scaffold Next.js frontend with Subwave theme, API client, and types"
```

---

## Task 10: Auth Context + Login Page

**Files:**
- Create: `frontend/src/lib/auth.tsx`
- Create: `frontend/src/app/login/page.tsx`
- Modify: `frontend/src/app/layout.tsx` (add providers)

- [ ] **Step 1: Create auth context**

Create `frontend/src/lib/auth.tsx`:

```tsx
"use client";

import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { getMe } from "./api";

interface AuthState {
  email: string | null;
  userId: string | null;
  loading: boolean;
}

const AuthContext = createContext<AuthState>({
  email: null,
  userId: null,
  loading: true,
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [auth, setAuth] = useState<AuthState>({
    email: null,
    userId: null,
    loading: true,
  });

  useEffect(() => {
    getMe()
      .then((data) => setAuth({ email: data.email, userId: data.user_id, loading: false }))
      .catch(() => setAuth({ email: null, userId: null, loading: false }));
  }, []);

  return <AuthContext.Provider value={auth}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  return useContext(AuthContext);
}
```

- [ ] **Step 2: Create login page**

Create `frontend/src/app/login/page.tsx`:

```tsx
"use client";

import { useAuth } from "@/lib/auth";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function LoginPage() {
  const { email, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && email) {
      router.replace("/dashboard");
    }
  }, [loading, email, router]);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-text-muted">Loading...</div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-8">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-accent mb-2">Subwave</h1>
        <p className="text-text-muted">Release Campaign Planner</p>
      </div>
      <a
        href="/auth/google"
        className="inline-flex items-center gap-3 rounded-lg bg-accent px-6 py-3 text-sm font-semibold text-bg-base transition-smooth glow-hover"
      >
        <svg className="h-5 w-5" viewBox="0 0 24 24">
          <path
            fill="currentColor"
            d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
          />
          <path
            fill="currentColor"
            d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
          />
          <path
            fill="currentColor"
            d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
          />
          <path
            fill="currentColor"
            d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
          />
        </svg>
        Sign in with Google
      </a>
    </div>
  );
}
```

- [ ] **Step 3: Update layout with providers**

Update `frontend/src/app/layout.tsx` to wrap children with `AuthProvider` and `QueryClientProvider`:

```tsx
import type { Metadata } from "next";
import { Space_Grotesk, Inter } from "next/font/google";
import { Providers } from "./providers";
import "./globals.css";

const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  variable: "--font-heading",
});

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-body",
});

export const metadata: Metadata = {
  title: "Subwave Release Planner",
  description: "Internal release campaign management for Subwave",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${spaceGrotesk.variable} ${inter.variable} antialiased min-h-screen`}
      >
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
```

Create `frontend/src/app/providers.tsx`:

```tsx
"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState, type ReactNode } from "react";
import { Toaster } from "sonner";
import { AuthProvider } from "@/lib/auth";

export function Providers({ children }: { children: ReactNode }) {
  const [queryClient] = useState(() => new QueryClient());

  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        {children}
        <Toaster theme="dark" />
      </AuthProvider>
    </QueryClientProvider>
  );
}
```

- [ ] **Step 4: Verify build**

```bash
cd frontend
npm run build
```

Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/auth.tsx frontend/src/app/login/ frontend/src/app/layout.tsx frontend/src/app/providers.tsx
git commit -m "feat: add auth context, login page, and app providers"
```

---

## Task 11: Dashboard Page (Campaign List)

**Files:**
- Create: `frontend/src/app/dashboard/page.tsx`
- Create: `frontend/src/components/campaign-card.tsx`
- Create: `frontend/src/hooks/use-campaign.ts`

- [ ] **Step 1: Create campaign data hook**

Create `frontend/src/hooks/use-campaign.ts`:

```tsx
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import * as api from "@/lib/api";
import type { Campaign } from "@/lib/types";

export function useCampaigns() {
  return useQuery<Campaign[]>({
    queryKey: ["campaigns"],
    queryFn: api.listCampaigns,
  });
}

export function useCampaign(id: string) {
  return useQuery<Campaign>({
    queryKey: ["campaign", id],
    queryFn: () => api.getCampaign(id),
  });
}

export function useCreateCampaign() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => api.createCampaign(name),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["campaigns"] }),
  });
}

export function useDuplicateCampaign() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.duplicateCampaign(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["campaigns"] }),
  });
}

export function useArchiveCampaign() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, archived }: { id: string; archived: boolean }) =>
      api.archiveCampaign(id, archived),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["campaigns"] }),
  });
}

export function useDeleteCampaign() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.deleteCampaign(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["campaigns"] }),
  });
}
```

- [ ] **Step 2: Create campaign card component**

Create `frontend/src/components/campaign-card.tsx`:

```tsx
"use client";

import type { Campaign } from "@/lib/types";
import { useRouter } from "next/navigation";
import { Copy, Archive, Trash2, MoreVertical } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useDuplicateCampaign, useArchiveCampaign, useDeleteCampaign } from "@/hooks/use-campaign";
import { toast } from "sonner";

export function CampaignCard({ campaign }: { campaign: Campaign }) {
  const router = useRouter();
  const duplicate = useDuplicateCampaign();
  const archive = useArchiveCampaign();
  const del = useDeleteCampaign();

  const handleDuplicate = (e: React.MouseEvent) => {
    e.stopPropagation();
    duplicate.mutate(campaign.id, {
      onSuccess: () => toast.success("Campaign duplicated"),
      onError: (err) => toast.error(err.message),
    });
  };

  const handleArchive = (e: React.MouseEvent) => {
    e.stopPropagation();
    archive.mutate(
      { id: campaign.id, archived: !campaign.archived },
      {
        onSuccess: () => toast.success(campaign.archived ? "Campaign unarchived" : "Campaign archived"),
        onError: (err) => toast.error(err.message),
      }
    );
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm("Permanently delete this campaign?")) return;
    del.mutate(campaign.id, {
      onSuccess: () => toast.success("Campaign deleted"),
      onError: (err) => toast.error(err.message),
    });
  };

  return (
    <div
      onClick={() => router.push(`/campaign/${campaign.id}`)}
      className="group cursor-pointer rounded-xl bg-bg-surface p-5 transition-smooth glow-hover border border-transparent hover:border-accent/20"
    >
      <div className="flex items-start justify-between">
        <h3 className="font-heading text-lg font-semibold text-text-primary">
          {campaign.name}
        </h3>
        <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-smooth">
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleDuplicate}>
            <Copy className="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleArchive}>
            <Archive className="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={handleDelete}>
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>
      <p className="mt-2 text-sm text-text-muted">
        Updated {new Date(campaign.updated_at).toLocaleDateString()}
      </p>
    </div>
  );
}
```

- [ ] **Step 3: Create dashboard page**

Create `frontend/src/app/dashboard/page.tsx`:

```tsx
"use client";

import { useAuth } from "@/lib/auth";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useCampaigns, useCreateCampaign } from "@/hooks/use-campaign";
import { CampaignCard } from "@/components/campaign-card";
import { Button } from "@/components/ui/button";
import { Plus, LogOut } from "lucide-react";
import { logout } from "@/lib/api";
import { toast } from "sonner";

export default function DashboardPage() {
  const { email, loading } = useAuth();
  const router = useRouter();
  const { data: campaigns, isLoading } = useCampaigns();
  const createCampaign = useCreateCampaign();
  const [newName, setNewName] = useState("");
  const [showCreate, setShowCreate] = useState(false);

  useEffect(() => {
    if (!loading && !email) {
      router.replace("/login");
    }
  }, [loading, email, router]);

  if (loading || isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-text-muted">Loading...</div>
      </div>
    );
  }

  const activeCampaigns = campaigns?.filter((c) => !c.archived) ?? [];
  const archivedCampaigns = campaigns?.filter((c) => c.archived) ?? [];

  const handleCreate = () => {
    if (!newName.trim()) return;
    createCampaign.mutate(newName.trim(), {
      onSuccess: () => {
        setNewName("");
        setShowCreate(false);
        toast.success("Campaign created");
      },
      onError: (err) => toast.error(err.message),
    });
  };

  const handleLogout = async () => {
    await logout();
    router.replace("/login");
  };

  return (
    <div className="min-h-screen p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-heading font-bold text-accent">Subwave</h1>
        <Button variant="ghost" size="sm" onClick={handleLogout}>
          <LogOut className="h-4 w-4 mr-2" />
          Sign out
        </Button>
      </div>

      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-heading font-semibold">Campaigns</h2>
        <Button
          onClick={() => setShowCreate(true)}
          className="bg-accent text-bg-base hover:bg-accent-dark"
        >
          <Plus className="h-4 w-4 mr-2" />
          New Campaign
        </Button>
      </div>

      {showCreate && (
        <div className="mb-6 flex gap-3 items-center bg-bg-surface p-4 rounded-xl">
          <input
            autoFocus
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleCreate()}
            placeholder="Campaign name..."
            className="flex-1 bg-transparent border border-border rounded-lg px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:ring-1 focus:ring-accent"
          />
          <Button onClick={handleCreate} className="bg-accent text-bg-base hover:bg-accent-dark">
            Create
          </Button>
          <Button variant="ghost" onClick={() => setShowCreate(false)}>
            Cancel
          </Button>
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {activeCampaigns.map((campaign) => (
          <CampaignCard key={campaign.id} campaign={campaign} />
        ))}
      </div>

      {archivedCampaigns.length > 0 && (
        <>
          <h3 className="text-lg font-heading font-semibold text-text-muted mt-10 mb-4">Archived</h3>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 opacity-60">
            {archivedCampaigns.map((campaign) => (
              <CampaignCard key={campaign.id} campaign={campaign} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Redirect root to dashboard**

Update `frontend/src/app/page.tsx`:

```tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function Home() {
  const router = useRouter();
  useEffect(() => {
    router.replace("/dashboard");
  }, [router]);
  return null;
}
```

- [ ] **Step 5: Verify build**

```bash
cd frontend
npm run build
```

Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/dashboard/ frontend/src/app/page.tsx frontend/src/components/campaign-card.tsx frontend/src/hooks/use-campaign.ts
git commit -m "feat: add dashboard page with campaign cards and CRUD"
```

---

## Task 12: Campaign Board (Tabbed Task Lists)

**Files:**
- Create: `frontend/src/app/campaign/[id]/page.tsx`
- Create: `frontend/src/components/task-list-tabs.tsx`
- Create: `frontend/src/components/task-group.tsx`
- Create: `frontend/src/components/task-item.tsx`
- Create: `frontend/src/components/hide-done-toggle.tsx`

- [ ] **Step 1: Create task item component**

Create `frontend/src/components/task-item.tsx`:

```tsx
"use client";

import type { Task } from "@/lib/types";
import { Circle, CircleDot, CheckCircle2 } from "lucide-react";

const statusConfig = {
  todo: { icon: Circle, color: "text-text-muted" },
  in_progress: { icon: CircleDot, color: "text-yellow-400" },
  done: { icon: CheckCircle2, color: "text-green-400" },
};

interface TaskItemProps {
  task: Task;
  onSelect: (task: Task) => void;
  onStatusChange: (taskId: string, status: Task["status"]) => void;
}

export function TaskItem({ task, onSelect, onStatusChange }: TaskItemProps) {
  const { icon: StatusIcon, color } = statusConfig[task.status];

  const cycleStatus = (e: React.MouseEvent) => {
    e.stopPropagation();
    const next: Record<string, Task["status"]> = {
      todo: "in_progress",
      in_progress: "done",
      done: "todo",
    };
    onStatusChange(task.id, next[task.status]);
  };

  return (
    <div
      onClick={() => onSelect(task)}
      className="flex items-center gap-3 rounded-lg px-3 py-2.5 cursor-pointer transition-smooth hover:bg-white/[0.03] group"
    >
      <button onClick={cycleStatus} className={`${color} transition-smooth`}>
        <StatusIcon className="h-4 w-4" />
      </button>
      <span className={`flex-1 text-sm ${task.status === "done" ? "text-text-muted line-through" : "text-text-primary"}`}>
        {task.name}
      </span>
      {task.due_date && (
        <span className="text-xs text-text-muted">
          {new Date(task.due_date).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
        </span>
      )}
      {task.subtasks && task.subtasks.length > 0 && (
        <span className="text-xs text-text-muted">
          {task.subtasks.filter((s) => s.is_complete).length}/{task.subtasks.length}
        </span>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Create task group component**

Create `frontend/src/components/task-group.tsx`:

```tsx
"use client";

import { useState } from "react";
import type { Task, TaskGroup as TaskGroupType } from "@/lib/types";
import { TaskItem } from "./task-item";
import { ChevronDown, ChevronRight, Plus } from "lucide-react";
import { createTask } from "@/lib/api";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

interface TaskGroupProps {
  group: TaskGroupType;
  campaignId: string;
  hideDone: boolean;
  onSelectTask: (task: Task) => void;
  onStatusChange: (taskId: string, status: Task["status"]) => void;
}

export function TaskGroup({ group, campaignId, hideDone, onSelectTask, onStatusChange }: TaskGroupProps) {
  const [collapsed, setCollapsed] = useState(group.collapsed);
  const [adding, setAdding] = useState(false);
  const [newTaskName, setNewTaskName] = useState("");
  const queryClient = useQueryClient();

  const visibleTasks = hideDone
    ? (group.tasks ?? []).filter((t) => t.status !== "done")
    : (group.tasks ?? []);

  const handleAddTask = async () => {
    if (!newTaskName.trim()) return;
    try {
      await createTask(group.id, newTaskName.trim());
      setNewTaskName("");
      setAdding(false);
      queryClient.invalidateQueries({ queryKey: ["campaign", campaignId] });
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  return (
    <div className="mb-4">
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="flex items-center gap-2 text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 hover:text-text-primary transition-smooth"
      >
        {collapsed ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
        {group.name}
        <span className="text-text-muted font-normal">({visibleTasks.length})</span>
      </button>

      {!collapsed && (
        <div className="space-y-0.5">
          {visibleTasks.map((task) => (
            <TaskItem
              key={task.id}
              task={task}
              onSelect={onSelectTask}
              onStatusChange={onStatusChange}
            />
          ))}

          {adding ? (
            <div className="flex gap-2 px-3 py-2">
              <input
                autoFocus
                value={newTaskName}
                onChange={(e) => setNewTaskName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleAddTask();
                  if (e.key === "Escape") setAdding(false);
                }}
                placeholder="Task name..."
                className="flex-1 bg-transparent border-b border-border text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent"
              />
            </div>
          ) : (
            <button
              onClick={() => setAdding(true)}
              className="flex items-center gap-2 px-3 py-1.5 text-xs text-text-muted hover:text-accent transition-smooth"
            >
              <Plus className="h-3 w-3" />
              Add task
            </button>
          )}
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 3: Create task list tabs**

Create `frontend/src/components/task-list-tabs.tsx`:

```tsx
"use client";

import type { TaskList } from "@/lib/types";

interface TaskListTabsProps {
  lists: TaskList[];
  activeId: string;
  onSelect: (id: string) => void;
}

export function TaskListTabs({ lists, activeId, onSelect }: TaskListTabsProps) {
  return (
    <div className="flex gap-0 bg-bg-surface rounded-lg overflow-hidden">
      {lists.map((list) => (
        <button
          key={list.id}
          onClick={() => onSelect(list.id)}
          className={`px-4 py-3 text-xs font-bold uppercase tracking-wider transition-smooth ${
            activeId === list.id
              ? "border-b-2"
              : "text-text-muted hover:text-text-primary"
          }`}
          style={
            activeId === list.id
              ? { color: list.color, borderBottomColor: list.color }
              : undefined
          }
        >
          {list.name}
        </button>
      ))}
    </div>
  );
}
```

- [ ] **Step 4: Create hide done toggle**

Create `frontend/src/components/hide-done-toggle.tsx`:

```tsx
"use client";

import { Eye, EyeOff } from "lucide-react";

export function HideDoneToggle({
  hidden,
  onToggle,
}: {
  hidden: boolean;
  onToggle: () => void;
}) {
  return (
    <button
      onClick={onToggle}
      className="flex items-center gap-2 text-xs text-text-muted hover:text-text-primary transition-smooth"
    >
      {hidden ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
      {hidden ? "Show completed" : "Hide completed"}
    </button>
  );
}
```

- [ ] **Step 5: Create campaign board page**

Create `frontend/src/app/campaign/[id]/page.tsx`:

```tsx
"use client";

import { use, useState } from "react";
import { useRouter } from "next/navigation";
import { useCampaign } from "@/hooks/use-campaign";
import { TaskListTabs } from "@/components/task-list-tabs";
import { TaskGroup } from "@/components/task-group";
import { HideDoneToggle } from "@/components/hide-done-toggle";
import { ArrowLeft, Calendar } from "lucide-react";
import { Button } from "@/components/ui/button";
import { updateTask } from "@/lib/api";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { Task } from "@/lib/types";

export default function CampaignPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const { data: campaign, isLoading } = useCampaign(id);
  const queryClient = useQueryClient();
  const [activeListId, setActiveListId] = useState<string | null>(null);
  const [hideDone, setHideDone] = useState(false);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);

  if (isLoading || !campaign) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-text-muted">Loading campaign...</div>
      </div>
    );
  }

  const lists = campaign.task_lists ?? [];
  const activeList = lists.find((l) => l.id === activeListId) ?? lists[0];

  if (activeList && !activeListId) {
    setActiveListId(activeList.id);
  }

  const handleStatusChange = async (taskId: string, status: Task["status"]) => {
    try {
      await updateTask(taskId, { status });
      queryClient.invalidateQueries({ queryKey: ["campaign", id] });
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  return (
    <div className="min-h-screen p-6 max-w-5xl mx-auto">
      {/* Top bar */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push("/dashboard")}>
            <ArrowLeft className="h-4 w-4 text-accent" />
          </Button>
          <h1 className="text-2xl font-heading font-bold text-text-primary">{campaign.name}</h1>
        </div>
        <div className="flex items-center gap-3">
          <HideDoneToggle hidden={hideDone} onToggle={() => setHideDone(!hideDone)} />
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.push(`/campaign/${id}/calendar`)}
          >
            <Calendar className="h-4 w-4 mr-2" />
            Calendar
          </Button>
        </div>
      </div>

      {/* Tabs */}
      {lists.length > 0 && (
        <TaskListTabs
          lists={lists}
          activeId={activeList?.id ?? ""}
          onSelect={setActiveListId}
        />
      )}

      {/* Active list content */}
      {activeList && (
        <div className="mt-4 bg-bg-surface rounded-xl p-5">
          {(activeList.task_groups ?? []).map((group) => (
            <TaskGroup
              key={group.id}
              group={group}
              campaignId={id}
              hideDone={hideDone}
              onSelectTask={setSelectedTask}
              onStatusChange={handleStatusChange}
            />
          ))}
        </div>
      )}

      {/* Task detail panel will be added in Task 13 */}
    </div>
  );
}
```

- [ ] **Step 6: Verify build**

```bash
cd frontend
npm run build
```

Expected: Build succeeds.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/app/campaign/ frontend/src/components/task-list-tabs.tsx frontend/src/components/task-group.tsx frontend/src/components/task-item.tsx frontend/src/components/hide-done-toggle.tsx
git commit -m "feat: add campaign board with tabbed task lists and task groups"
```

---

## Task 13: Task Detail Slide-Out Panel

**Files:**
- Create: `frontend/src/components/task-detail.tsx`
- Create: `frontend/src/components/subtask-item.tsx`
- Modify: `frontend/src/app/campaign/[id]/page.tsx` (wire up panel)

- [ ] **Step 1: Create subtask item**

Create `frontend/src/components/subtask-item.tsx`:

```tsx
"use client";

import type { Subtask } from "@/lib/types";
import { Trash2 } from "lucide-react";
import { updateSubtask, deleteSubtask } from "@/lib/api";
import { toast } from "sonner";

interface SubtaskItemProps {
  subtask: Subtask;
  onUpdate: () => void;
}

export function SubtaskItem({ subtask, onUpdate }: SubtaskItemProps) {
  const handleToggle = async () => {
    try {
      await updateSubtask(subtask.id, { is_complete: !subtask.is_complete });
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  const handleDelete = async () => {
    try {
      await deleteSubtask(subtask.id);
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  return (
    <div className="flex items-center gap-3 group py-1">
      <input
        type="checkbox"
        checked={subtask.is_complete}
        onChange={handleToggle}
        className="h-4 w-4 rounded border-border accent-accent"
      />
      <span className={`flex-1 text-sm ${subtask.is_complete ? "text-text-muted line-through" : "text-text-primary"}`}>
        {subtask.name}
      </span>
      <button
        onClick={handleDelete}
        className="opacity-0 group-hover:opacity-100 text-text-muted hover:text-destructive transition-smooth"
      >
        <Trash2 className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}
```

- [ ] **Step 2: Create task detail panel**

Create `frontend/src/components/task-detail.tsx`:

```tsx
"use client";

import { useState } from "react";
import type { Task } from "@/lib/types";
import { SubtaskItem } from "./subtask-item";
import { X, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { updateTask, deleteTask, createSubtask } from "@/lib/api";
import { toast } from "sonner";

const statusOptions = [
  { value: "todo", label: "To Do", color: "text-text-muted" },
  { value: "in_progress", label: "In Progress", color: "text-yellow-400" },
  { value: "done", label: "Done", color: "text-green-400" },
] as const;

interface TaskDetailProps {
  task: Task;
  onClose: () => void;
  onUpdate: () => void;
}

export function TaskDetail({ task, onClose, onUpdate }: TaskDetailProps) {
  const [name, setName] = useState(task.name);
  const [dueDate, setDueDate] = useState(task.due_date ?? "");
  const [newSubtaskName, setNewSubtaskName] = useState("");

  const handleNameBlur = async () => {
    if (name !== task.name && name.trim()) {
      try {
        await updateTask(task.id, { name: name.trim() } as any);
        onUpdate();
      } catch (err: any) {
        toast.error(err.message);
      }
    }
  };

  const handleStatusChange = async (status: string) => {
    try {
      await updateTask(task.id, { status } as any);
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  const handleDueDateChange = async (date: string) => {
    setDueDate(date);
    try {
      await updateTask(task.id, { due_date: date || undefined } as any);
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  const handleAddSubtask = async () => {
    if (!newSubtaskName.trim()) return;
    try {
      await createSubtask(task.id, newSubtaskName.trim());
      setNewSubtaskName("");
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Delete this task?")) return;
    try {
      await deleteTask(task.id);
      onClose();
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  return (
    <div className="fixed inset-y-0 right-0 w-96 bg-bg-surface border-l border-border glass p-6 overflow-y-auto z-50">
      <div className="flex items-center justify-between mb-6">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          onBlur={handleNameBlur}
          className="text-lg font-heading font-semibold bg-transparent text-text-primary border-none focus:outline-none focus:ring-0 w-full"
        />
        <Button variant="ghost" size="icon" onClick={onClose}>
          <X className="h-4 w-4" />
        </Button>
      </div>

      {/* Status */}
      <div className="mb-5">
        <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
          Status
        </label>
        <div className="flex gap-2">
          {statusOptions.map((opt) => (
            <button
              key={opt.value}
              onClick={() => handleStatusChange(opt.value)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-smooth ${
                task.status === opt.value
                  ? `${opt.color} bg-white/10`
                  : "text-text-muted hover:text-text-primary"
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      {/* Due date */}
      <div className="mb-5">
        <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
          Due Date
        </label>
        <input
          type="date"
          value={dueDate}
          onChange={(e) => handleDueDateChange(e.target.value)}
          className="bg-transparent border border-border rounded-lg px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-1 focus:ring-accent"
        />
      </div>

      {/* Description placeholder (Tiptap will be added in Task 14) */}
      <div className="mb-5">
        <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
          Description
        </label>
        <div className="text-sm text-text-muted italic p-3 border border-border rounded-lg">
          Rich text editor (Tiptap) — coming in next task
        </div>
      </div>

      {/* Subtasks */}
      <div className="mb-5">
        <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
          Subtasks
        </label>
        <div className="space-y-1">
          {(task.subtasks ?? []).map((subtask) => (
            <SubtaskItem key={subtask.id} subtask={subtask} onUpdate={onUpdate} />
          ))}
        </div>
        <div className="flex gap-2 mt-2">
          <input
            value={newSubtaskName}
            onChange={(e) => setNewSubtaskName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleAddSubtask()}
            placeholder="Add subtask..."
            className="flex-1 bg-transparent border-b border-border text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent"
          />
          <button onClick={handleAddSubtask} className="text-accent">
            <Plus className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Delete */}
      <Button
        variant="ghost"
        size="sm"
        className="text-destructive hover:text-destructive mt-4"
        onClick={handleDelete}
      >
        <Trash2 className="h-4 w-4 mr-2" />
        Delete task
      </Button>
    </div>
  );
}
```

- [ ] **Step 3: Wire task detail into campaign page**

In `frontend/src/app/campaign/[id]/page.tsx`, add the TaskDetail import and render it when `selectedTask` is set. Add at the bottom of the component JSX (before the final closing `</div>`):

```tsx
import { TaskDetail } from "@/components/task-detail";

// ... inside JSX, after the active list content:
{selectedTask && (
  <TaskDetail
    task={selectedTask}
    onClose={() => setSelectedTask(null)}
    onUpdate={() => {
      queryClient.invalidateQueries({ queryKey: ["campaign", id] });
      // Refresh selected task
      if (selectedTask) {
        const updated = campaign?.task_lists
          ?.flatMap((l) => l.task_groups ?? [])
          ?.flatMap((g) => g.tasks ?? [])
          ?.find((t) => t.id === selectedTask.id);
        if (updated) setSelectedTask(updated);
      }
    }}
  />
)}
```

- [ ] **Step 4: Verify build**

```bash
cd frontend
npm run build
```

Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/task-detail.tsx frontend/src/components/subtask-item.tsx frontend/src/app/campaign/
git commit -m "feat: add task detail slide-out panel with subtasks"
```

---

## Task 14: Tiptap Rich Text Editor

**Files:**
- Modify: `frontend/src/components/task-detail.tsx` (replace placeholder)

- [ ] **Step 1: Install Tiptap**

```bash
cd frontend
npm install @tiptap/react @tiptap/starter-kit @tiptap/extension-placeholder
```

- [ ] **Step 2: Create Tiptap editor component**

Create `frontend/src/components/rich-text-editor.tsx`:

```tsx
"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { useEffect } from "react";

interface RichTextEditorProps {
  content: Record<string, unknown> | null | undefined;
  onUpdate: (content: Record<string, unknown>) => void;
}

export function RichTextEditor({ content, onUpdate }: RichTextEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder: "Add a description..." }),
    ],
    content: content ?? undefined,
    editorProps: {
      attributes: {
        class:
          "prose prose-invert prose-sm max-w-none min-h-[100px] p-3 focus:outline-none text-text-primary [&_p]:my-1 [&_ul]:my-1 [&_ol]:my-1",
      },
    },
    onBlur: ({ editor }) => {
      onUpdate(editor.getJSON() as Record<string, unknown>);
    },
  });

  useEffect(() => {
    if (editor && content) {
      const current = JSON.stringify(editor.getJSON());
      const incoming = JSON.stringify(content);
      if (current !== incoming) {
        editor.commands.setContent(content);
      }
    }
  }, [content, editor]);

  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <EditorContent editor={editor} />
    </div>
  );
}
```

- [ ] **Step 3: Replace placeholder in task detail**

In `frontend/src/components/task-detail.tsx`, replace the description placeholder section:

```tsx
import { RichTextEditor } from "./rich-text-editor";

// Replace the description placeholder with:
<div className="mb-5">
  <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
    Description
  </label>
  <RichTextEditor
    content={task.description ?? null}
    onUpdate={async (content) => {
      try {
        await updateTask(task.id, { description: content } as any);
        onUpdate();
      } catch (err: any) {
        toast.error(err.message);
      }
    }}
  />
</div>
```

- [ ] **Step 4: Verify build**

```bash
cd frontend
npm run build
```

Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/rich-text-editor.tsx frontend/src/components/task-detail.tsx
git commit -m "feat: add Tiptap rich text editor for task descriptions"
```

---

## Task 15: Drag-and-Drop with @dnd-kit

**Files:**
- Create: `frontend/src/hooks/use-drag-drop.ts`
- Modify: `frontend/src/components/task-group.tsx` (wrap with sortable)
- Modify: `frontend/src/app/campaign/[id]/page.tsx` (add DndContext)

- [ ] **Step 1: Install @dnd-kit**

```bash
cd frontend
npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
```

- [ ] **Step 2: Create drag-drop hook**

Create `frontend/src/hooks/use-drag-drop.ts`:

```tsx
import { useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { reorderTask } from "@/lib/api";
import type { DragEndEvent } from "@dnd-kit/core";
import { toast } from "sonner";

export function useTaskDragDrop(campaignId: string) {
  const queryClient = useQueryClient();

  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      const { active, over } = event;
      if (!over || active.id === over.id) return;

      const taskId = active.id as string;
      const activeData = active.data.current as { groupId: string; position: number };
      const overData = over.data.current as { groupId: string; position: number };

      const targetGroupId = overData?.groupId ?? activeData.groupId;
      const newPosition = overData?.position ?? 0;

      try {
        await reorderTask(taskId, targetGroupId, newPosition);
        queryClient.invalidateQueries({ queryKey: ["campaign", campaignId] });
      } catch (err: any) {
        toast.error(err.message);
      }
    },
    [campaignId, queryClient]
  );

  return { handleDragEnd };
}
```

- [ ] **Step 3: Wrap task items with sortable**

Update `frontend/src/components/task-item.tsx` to add sortable support:

Add to imports:
```tsx
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical } from "lucide-react";
```

Update the component to use `useSortable`:

```tsx
export function TaskItem({ task, onSelect, onStatusChange }: TaskItemProps) {
  const { icon: StatusIcon, color } = statusConfig[task.status];
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: task.id,
    data: { groupId: task.task_group_id, position: task.position },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const cycleStatus = (e: React.MouseEvent) => {
    e.stopPropagation();
    const next: Record<string, Task["status"]> = {
      todo: "in_progress",
      in_progress: "done",
      done: "todo",
    };
    onStatusChange(task.id, next[task.status]);
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      onClick={() => onSelect(task)}
      className="flex items-center gap-3 rounded-lg px-3 py-2.5 cursor-pointer transition-smooth hover:bg-white/[0.03] group"
    >
      <button
        {...attributes}
        {...listeners}
        className="opacity-0 group-hover:opacity-100 cursor-grab text-text-muted"
        onClick={(e) => e.stopPropagation()}
      >
        <GripVertical className="h-3.5 w-3.5" />
      </button>
      <button onClick={cycleStatus} className={`${color} transition-smooth`}>
        <StatusIcon className="h-4 w-4" />
      </button>
      <span className={`flex-1 text-sm ${task.status === "done" ? "text-text-muted line-through" : "text-text-primary"}`}>
        {task.name}
      </span>
      {task.due_date && (
        <span className="text-xs text-text-muted">
          {new Date(task.due_date).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
        </span>
      )}
      {task.subtasks && task.subtasks.length > 0 && (
        <span className="text-xs text-text-muted">
          {task.subtasks.filter((s) => s.is_complete).length}/{task.subtasks.length}
        </span>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Add DndContext to task group**

Update `frontend/src/components/task-group.tsx` — wrap the task list with SortableContext:

Add imports:
```tsx
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
```

Wrap the tasks div with `<SortableContext>`:

```tsx
<SortableContext
  items={(visibleTasks).map((t) => t.id)}
  strategy={verticalListSortingStrategy}
>
  {visibleTasks.map((task) => (
    <TaskItem ... />
  ))}
</SortableContext>
```

- [ ] **Step 5: Add DndContext to campaign page**

Update `frontend/src/app/campaign/[id]/page.tsx`:

Add imports:
```tsx
import { DndContext, closestCenter } from "@dnd-kit/core";
import { useTaskDragDrop } from "@/hooks/use-drag-drop";
```

Wrap the active list content with DndContext:
```tsx
const { handleDragEnd } = useTaskDragDrop(id);

// Wrap the task groups rendering:
<DndContext collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
  {/* ... task groups ... */}
</DndContext>
```

- [ ] **Step 6: Verify build**

```bash
cd frontend
npm run build
```

Expected: Build succeeds.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/hooks/use-drag-drop.ts frontend/src/components/task-item.tsx frontend/src/components/task-group.tsx frontend/src/app/campaign/
git commit -m "feat: add drag-and-drop reordering with @dnd-kit"
```

---

## Task 16: Calendar View

**Files:**
- Create: `frontend/src/app/campaign/[id]/calendar/page.tsx`
- Create: `frontend/src/components/calendar-view.tsx`

- [ ] **Step 1: Create calendar view component**

Create `frontend/src/components/calendar-view.tsx`:

```tsx
"use client";

import { useMemo } from "react";
import type { Task } from "@/lib/types";

interface CalendarViewProps {
  tasks: Task[];
  onSelectTask: (task: Task) => void;
}

const statusColors: Record<string, string> = {
  todo: "bg-gray-500/20 text-gray-400",
  in_progress: "bg-yellow-500/20 text-yellow-400",
  done: "bg-green-500/20 text-green-400",
};

export function CalendarView({ tasks, onSelectTask }: CalendarViewProps) {
  const today = new Date();
  const currentMonth = today.getMonth();
  const currentYear = today.getFullYear();

  const daysInMonth = new Date(currentYear, currentMonth + 1, 0).getDate();
  const firstDayOfWeek = new Date(currentYear, currentMonth, 1).getDay();

  const tasksByDate = useMemo(() => {
    const map: Record<string, Task[]> = {};
    for (const task of tasks) {
      if (task.due_date) {
        const key = task.due_date;
        if (!map[key]) map[key] = [];
        map[key].push(task);
      }
    }
    return map;
  }, [tasks]);

  const days = Array.from({ length: daysInMonth }, (_, i) => i + 1);
  const blanks = Array.from({ length: firstDayOfWeek }, (_, i) => i);

  const formatDate = (day: number) => {
    const d = new Date(currentYear, currentMonth, day);
    return d.toISOString().split("T")[0];
  };

  return (
    <div>
      <h3 className="text-lg font-heading font-semibold text-text-primary mb-4">
        {new Date(currentYear, currentMonth).toLocaleDateString("en-US", { month: "long", year: "numeric" })}
      </h3>

      <div className="grid grid-cols-7 gap-1">
        {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d) => (
          <div key={d} className="text-xs font-semibold text-text-muted text-center py-2">
            {d}
          </div>
        ))}
        {blanks.map((i) => (
          <div key={`blank-${i}`} />
        ))}
        {days.map((day) => {
          const dateStr = formatDate(day);
          const dayTasks = tasksByDate[dateStr] ?? [];
          const isToday = day === today.getDate();

          return (
            <div
              key={day}
              className={`min-h-[80px] rounded-lg p-2 text-xs ${
                isToday ? "border border-accent/30 bg-accent/5" : "bg-bg-surface"
              }`}
            >
              <div className={`font-medium mb-1 ${isToday ? "text-accent" : "text-text-muted"}`}>
                {day}
              </div>
              {dayTasks.map((task) => (
                <button
                  key={task.id}
                  onClick={() => onSelectTask(task)}
                  className={`w-full text-left px-1.5 py-0.5 rounded text-[10px] truncate mb-0.5 transition-smooth hover:opacity-80 ${statusColors[task.status]}`}
                >
                  {task.name}
                </button>
              ))}
            </div>
          );
        })}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create calendar page**

Create `frontend/src/app/campaign/[id]/calendar/page.tsx`:

```tsx
"use client";

import { use, useState } from "react";
import { useRouter } from "next/navigation";
import { useCampaign } from "@/hooks/use-campaign";
import { CalendarView } from "@/components/calendar-view";
import { TaskDetail } from "@/components/task-detail";
import { ArrowLeft, LayoutGrid } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useQueryClient } from "@tanstack/react-query";
import type { Task } from "@/lib/types";

export default function CalendarPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const { data: campaign, isLoading } = useCampaign(id);
  const queryClient = useQueryClient();
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);

  if (isLoading || !campaign) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-text-muted">Loading...</div>
      </div>
    );
  }

  // Flatten all tasks from all lists/groups
  const allTasks = (campaign.task_lists ?? [])
    .flatMap((l) => l.task_groups ?? [])
    .flatMap((g) => g.tasks ?? []);

  return (
    <div className="min-h-screen p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push("/dashboard")}>
            <ArrowLeft className="h-4 w-4 text-accent" />
          </Button>
          <h1 className="text-2xl font-heading font-bold text-text-primary">{campaign.name}</h1>
        </div>
        <Button variant="ghost" size="sm" onClick={() => router.push(`/campaign/${id}`)}>
          <LayoutGrid className="h-4 w-4 mr-2" />
          Board
        </Button>
      </div>

      <CalendarView tasks={allTasks} onSelectTask={setSelectedTask} />

      {selectedTask && (
        <TaskDetail
          task={selectedTask}
          onClose={() => setSelectedTask(null)}
          onUpdate={() => queryClient.invalidateQueries({ queryKey: ["campaign", id] })}
        />
      )}
    </div>
  );
}
```

- [ ] **Step 3: Verify build**

```bash
cd frontend
npm run build
```

Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/campaign/[id]/calendar/ frontend/src/components/calendar-view.tsx
git commit -m "feat: add calendar view for tasks with due dates"
```

---

## Task 17: Dockerfile + Cloudflare Deployment Config

**Files:**
- Create: `backend/Dockerfile`
- Create: `src/index.ts`
- Create: `wrangler.toml`
- Create: `package.json` (root, for wrangler)
- Create: `tsconfig.json` (root, for worker)

- [ ] **Step 1: Create Dockerfile**

Create `backend/Dockerfile`:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server/main.go

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
RUN addgroup -S app && adduser -S app -G app
WORKDIR /home/app
COPY --from=builder /app/server .
RUN chown app:app ./server
USER app
EXPOSE 8080
CMD ["./server"]
```

- [ ] **Step 2: Create Cloudflare Worker**

Create `src/index.ts`:

```typescript
import { Container, getRandom } from "@cloudflare/containers";
import { DurableObject } from "cloudflare:workers";

interface Env {
  BACKEND: DurableObjectNamespace;
  ASSETS: Fetcher;
  DATABASE_URL: string;
  GOOGLE_CLIENT_ID: string;
  GOOGLE_CLIENT_SECRET: string;
  SESSION_SECRET: string;
  ALLOWED_EMAILS: string;
  OAUTH_REDIRECT_URL: string;
  FRONTEND_URL: string;
}

export class Backend extends Container<Env> {
  defaultPort = 8080;
  sleepAfter = "10m";
  enableInternet = true;

  constructor(ctx: DurableObject["ctx"], env: Env) {
    super(ctx, env);
    this.envVars = {
      DATABASE_URL: env.DATABASE_URL,
      GOOGLE_CLIENT_ID: env.GOOGLE_CLIENT_ID,
      GOOGLE_CLIENT_SECRET: env.GOOGLE_CLIENT_SECRET,
      SESSION_SECRET: env.SESSION_SECRET,
      ALLOWED_EMAILS: env.ALLOWED_EMAILS,
      OAUTH_REDIRECT_URL: env.OAUTH_REDIRECT_URL,
      FRONTEND_URL: env.FRONTEND_URL,
      ENV: "production",
    };
  }
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    // Route API and auth requests to Go container
    if (url.pathname.startsWith("/api") || url.pathname.startsWith("/auth")) {
      try {
        const container = await getRandom(env.BACKEND, 1);
        return await container.fetch(request);
      } catch (e: any) {
        return new Response(JSON.stringify({ error: e.message }), {
          status: 502,
          headers: { "Content-Type": "application/json" },
        });
      }
    }

    // Serve static frontend
    return env.ASSETS.fetch(request);
  },
};
```

- [ ] **Step 3: Create wrangler.toml**

Create `wrangler.toml`:

```toml
name = "music-release-planner"
main = "src/index.ts"
compatibility_date = "2025-09-01"

[observability]
enabled = true

[assets]
directory = "./frontend/out"
binding = "ASSETS"
not_found_handling = "single-page-application"
run_worker_first = ["/api/*", "/auth/*"]

[[containers]]
class_name = "Backend"
image = "./backend/Dockerfile"
max_instances = 1

[[durable_objects.bindings]]
class_name = "Backend"
name = "BACKEND"

[[migrations]]
tag = "v1"
new_sqlite_classes = ["Backend"]
```

- [ ] **Step 4: Create root package.json for wrangler**

Create `package.json` (project root):

```json
{
  "name": "music-release-planner",
  "private": true,
  "scripts": {
    "deploy": "wrangler deploy"
  },
  "devDependencies": {
    "@cloudflare/containers": "^0.1.0",
    "@cloudflare/workers-types": "^4.0.0",
    "wrangler": "^4.0.0"
  }
}
```

- [ ] **Step 5: Create root tsconfig.json**

Create `tsconfig.json` (project root):

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ES2022",
    "moduleResolution": "bundler",
    "strict": true,
    "types": ["@cloudflare/workers-types"]
  },
  "include": ["src/**/*.ts"]
}
```

- [ ] **Step 6: Update Makefile deploy target**

Update `Makefile` — replace the deploy target:

```makefile
deploy:
	cd frontend && npm run build
	npm run deploy
```

- [ ] **Step 7: Install root dependencies**

```bash
npm install
```

- [ ] **Step 8: Verify worker builds**

```bash
npx wrangler deploy --dry-run
```

Expected: Dry run completes without errors.

- [ ] **Step 9: Commit**

```bash
git add backend/Dockerfile src/ wrangler.toml package.json tsconfig.json Makefile
git commit -m "feat: add Cloudflare Workers deployment config with container backend"
```

---

## Task 18: Environment Setup + Local Dev Verification

**Files:**
- Create: `.env.example`

- [ ] **Step 1: Create .env.example**

Create `.env.example`:

```
# Database
DATABASE_URL=postgresql://user:pass@host/dbname?sslmode=require

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
OAUTH_REDIRECT_URL=http://localhost:8080/auth/google/callback

# Session
SESSION_SECRET=generate-a-random-32-char-string

# Auth whitelist (comma-separated)
ALLOWED_EMAILS=user1@example.com,user2@example.com

# Frontend URL (for OAuth redirects)
FRONTEND_URL=http://localhost:3000

# Frontend API base URL (empty in prod, Go backend in dev)
NEXT_PUBLIC_API_URL=http://localhost:8080

# Environment
ENV=development
PORT=8080
```

- [ ] **Step 2: Run the database migration**

```bash
psql $DATABASE_URL -f backend/migrations/001_initial.sql
```

Expected: Tables created successfully.

- [ ] **Step 3: Start the full dev environment**

```bash
make dev
```

Expected: Backend on :8080, frontend on :3000. Frontend proxies API requests to backend.

- [ ] **Step 4: Verify health endpoint**

```bash
curl http://localhost:8080/api/health
```

Expected: `{"status":"ok"}`

- [ ] **Step 5: Commit**

```bash
git add .env.example
git commit -m "feat: add environment example and verify local dev setup"
```
