# Performance Fix: Batch DB Operations & Parallel Frontend Fetching

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate the 12+ second page loads and campaign creation hangs caused by excessive sequential DB round trips to Neon PostgreSQL and waterfalled frontend fetches.

**Architecture:** Three independent fixes targeting the three root causes: (1) replace 178+ individual INSERTs in `populateTemplate` with 4 batched multi-row INSERTs, (2) fetch auth and campaigns in parallel on the frontend, (3) disable React Query's default 3x retry on auth to fail fast on cold start. Each fix is independently deployable and testable.

**Tech Stack:** Go/pgx (batch INSERT with `unnest`), React/React Query, Next.js

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `backend/internal/store/campaign.go` | Modify (lines 582-647) | Replace `populateTemplate` loop with batched INSERTs |
| `backend/internal/store/campaign_test.go` | Modify | Add benchmark test for `CreateCampaign` |
| `frontend/src/app/providers.tsx` | Modify | Add `retry: 1` to React Query defaults |
| `frontend/src/lib/auth.tsx` | Modify | Expose campaigns prefetch alongside auth |
| `frontend/src/app/dashboard/page.tsx` | Modify | Fetch campaigns in parallel with auth, not after |
| `frontend/src/hooks/use-campaign.ts` | Modify | Remove `enabled` gate that blocks on auth |

---

### Task 1: Batch `populateTemplate` INSERTs

The biggest win. Currently 178 sequential INSERTs for the "single" template (208 for LP/EP). Replace with 4 multi-row INSERTs using PostgreSQL `unnest` arrays — one per entity type.

**Files:**
- Modify: `backend/internal/store/campaign.go:582-647`
- Test: `backend/internal/store/campaign_test.go`

- [ ] **Step 1: Write a benchmark test for CreateCampaign**

Add to `backend/internal/store/campaign_test.go`:

```go
func BenchmarkCreateCampaign(b *testing.B) {
	if os.Getenv("DATABASE_URL") == "" {
		b.Skip("DATABASE_URL not set, skipping integration benchmark")
	}
	s, err := New()
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	user, err := s.UpsertUser(context.Background(), "bench@example.com", "Bench User", nil)
	if err != nil {
		b.Fatalf("failed to create user: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		campaign, err := s.CreateCampaign(context.Background(), user.ID, fmt.Sprintf("Bench %d", i), nil, "single")
		if err != nil {
			b.Fatalf("failed to create campaign: %v", err)
		}
		s.DeleteCampaign(context.Background(), campaign.ID)
	}
}
```

Note: this file needs `"fmt"` added to its imports.

- [ ] **Step 2: Run benchmark to capture baseline**

```bash
cd backend && go test -tags=integration -bench=BenchmarkCreateCampaign -benchtime=3x -count=1 ./internal/store/ -v
```

Record the ns/op. With Neon this should be 1-5 seconds per op.

- [ ] **Step 3: Rewrite `populateTemplate` to collect all entities, then batch INSERT**

Replace the `populateTemplate` function in `backend/internal/store/campaign.go` (lines 582-647) with:

```go
func (s *Store) populateTemplate(ctx context.Context, tx pgx.Tx, campaignID, templateType string, releaseDate *string) error {
	tmpl := template.GetTemplate(templateType)

	// Parse release date if provided
	var relDate *time.Time
	if releaseDate != nil && *releaseDate != "" {
		t, err := time.Parse("2006-01-02", *releaseDate)
		if err == nil {
			relDate = &t
		}
	}

	// --- Step 1: Batch insert all task lists ---
	var listNames []string
	var listColors []string
	var listPositions []int
	for listPos, list := range tmpl {
		listNames = append(listNames, list.Name)
		listColors = append(listColors, list.Color)
		listPositions = append(listPositions, (listPos+1)*100)
	}

	listRows, err := tx.Query(ctx, `
		INSERT INTO task_lists (campaign_id, name, color, position)
		SELECT $1, unnest($2::text[]), unnest($3::text[]), unnest($4::int[])
		RETURNING id, name
	`, campaignID, listNames, listColors, listPositions)
	if err != nil {
		return fmt.Errorf("batch inserting task lists: %w", err)
	}
	listIDByName := map[string]string{}
	for listRows.Next() {
		var id, name string
		if err := listRows.Scan(&id, &name); err != nil {
			listRows.Close()
			return fmt.Errorf("scanning task list: %w", err)
		}
		listIDByName[name] = id
	}
	listRows.Close()
	if err := listRows.Err(); err != nil {
		return fmt.Errorf("task list rows: %w", err)
	}

	// --- Step 2: Batch insert all task groups ---
	var groupListIDs []string
	var groupNames []string
	var groupPositions []int
	for _, list := range tmpl {
		listID := listIDByName[list.Name]
		for groupPos, group := range list.Groups {
			groupListIDs = append(groupListIDs, listID)
			groupNames = append(groupNames, group.Name)
			groupPositions = append(groupPositions, (groupPos+1)*100)
		}
	}

	groupRows, err := tx.Query(ctx, `
		INSERT INTO task_groups (task_list_id, name, position)
		SELECT unnest($1::uuid[]), unnest($2::text[]), unnest($3::int[])
		RETURNING id, task_list_id, name
	`, groupListIDs, groupNames, groupPositions)
	if err != nil {
		return fmt.Errorf("batch inserting task groups: %w", err)
	}
	// Key by "listID|groupName" since group names are unique within a list
	groupIDByKey := map[string]string{}
	for groupRows.Next() {
		var id, listID, name string
		if err := groupRows.Scan(&id, &listID, &name); err != nil {
			groupRows.Close()
			return fmt.Errorf("scanning task group: %w", err)
		}
		groupIDByKey[listID+"|"+name] = id
	}
	groupRows.Close()
	if err := groupRows.Err(); err != nil {
		return fmt.Errorf("task group rows: %w", err)
	}

	// --- Step 3: Batch insert all tasks ---
	var taskGroupIDs []string
	var taskNames []string
	var taskDueDates []*string
	var taskPositions []int
	type taskRef struct {
		groupID  string
		name     string
		subtasks []string
	}
	var tasksWithSubtasks []taskRef

	for _, list := range tmpl {
		listID := listIDByName[list.Name]
		for _, group := range list.Groups {
			groupID := groupIDByKey[listID+"|"+group.Name]
			for taskPos, task := range group.Tasks {
				taskGroupIDs = append(taskGroupIDs, groupID)
				taskNames = append(taskNames, task.Name)
				taskPositions = append(taskPositions, (taskPos+1)*100)

				var dueDate *string
				if relDate != nil && task.DaysOffset != nil {
					d := relDate.AddDate(0, 0, *task.DaysOffset).Format("2006-01-02")
					dueDate = &d
				}
				taskDueDates = append(taskDueDates, dueDate)

				if len(task.Subtasks) > 0 {
					tasksWithSubtasks = append(tasksWithSubtasks, taskRef{
						groupID:  groupID,
						name:     task.Name,
						subtasks: task.Subtasks,
					})
				}
			}
		}
	}

	// pgx needs a concrete []string for due dates, not []*string.
	// Use a sentinel for NULL and cast in SQL.
	dueDateStrs := make([]string, len(taskDueDates))
	for i, d := range taskDueDates {
		if d != nil {
			dueDateStrs[i] = *d
		} else {
			dueDateStrs[i] = ""
		}
	}

	taskRows, err := tx.Query(ctx, `
		INSERT INTO tasks (task_group_id, name, due_date, position)
		SELECT unnest($1::uuid[]), unnest($2::text[]),
			NULLIF(unnest($3::text[]), '')::date,
			unnest($4::int[])
		RETURNING id, task_group_id, name
	`, taskGroupIDs, taskNames, dueDateStrs, taskPositions)
	if err != nil {
		return fmt.Errorf("batch inserting tasks: %w", err)
	}
	taskIDByKey := map[string]string{}
	for taskRows.Next() {
		var id, groupID, name string
		if err := taskRows.Scan(&id, &groupID, &name); err != nil {
			taskRows.Close()
			return fmt.Errorf("scanning task: %w", err)
		}
		taskIDByKey[groupID+"|"+name] = id
	}
	taskRows.Close()
	if err := taskRows.Err(); err != nil {
		return fmt.Errorf("task rows: %w", err)
	}

	// --- Step 4: Batch insert all subtasks ---
	if len(tasksWithSubtasks) > 0 {
		var subTaskIDs []string
		var subNames []string
		var subPositions []int
		for _, ref := range tasksWithSubtasks {
			taskID := taskIDByKey[ref.groupID+"|"+ref.name]
			for subPos, subtask := range ref.subtasks {
				subTaskIDs = append(subTaskIDs, taskID)
				subNames = append(subNames, subtask)
				subPositions = append(subPositions, (subPos+1)*100)
			}
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO subtasks (task_id, name, position)
			SELECT unnest($1::uuid[]), unnest($2::text[]), unnest($3::int[])
		`, subTaskIDs, subNames, subPositions)
		if err != nil {
			return fmt.Errorf("batch inserting subtasks: %w", err)
		}
	}

	return nil
}
```

**Why this works:** Instead of 178 round trips, this does 4 (one per entity type). Each uses `unnest` to expand Go slices into multi-row INSERTs. The `RETURNING` clause gives us the generated UUIDs we need for the next level's foreign keys. Entity names are unique within their parent scope, so `parentID|name` is a safe lookup key.

- [ ] **Step 4: Verify build compiles**

```bash
cd backend && go build ./... && go vet ./...
```

- [ ] **Step 5: Run the existing integration test to verify correctness**

```bash
cd backend && go test -tags=integration -run=TestCreateCampaign -count=1 ./internal/store/ -v
```

The existing `TestCreateCampaign` creates a campaign with template "single" and verifies it returns. `TestGetFullCampaign` verifies the full hierarchy loads. Both must pass.

- [ ] **Step 6: Run benchmark to measure improvement**

```bash
cd backend && go test -tags=integration -bench=BenchmarkCreateCampaign -benchtime=3x -count=1 ./internal/store/ -v
```

Compare ns/op against baseline from Step 2. Expect 5-10x improvement.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/store/campaign.go backend/internal/store/campaign_test.go
git commit -m "perf: batch template INSERTs — 178 round trips to 4"
```

---

### Task 2: Parallel auth + campaigns fetch on dashboard

Currently `useCampaigns` won't fire until `useAuth` resolves (because the dashboard gates on `authLoading`). These should fire simultaneously — the backend's `RequireAuth` middleware handles the auth check, so the campaigns request doesn't need to wait for the frontend auth state.

**Files:**
- Modify: `frontend/src/app/dashboard/page.tsx`
- Modify: `frontend/src/hooks/use-campaign.ts`

- [ ] **Step 1: Remove the auth-loading gate from useCampaigns**

In `frontend/src/hooks/use-campaign.ts`, the `useCampaigns` hook has no `enabled` gate — it fires immediately. Good, no change needed here.

The bottleneck is in `frontend/src/lib/auth.tsx` — the `AuthProvider` calls `getMe()` on mount, and the dashboard doesn't render campaign content until `authLoading` is false. But `useCampaigns()` is actually called unconditionally in the dashboard. The real issue is React Query's default retry behavior: if `getMe()` is slow, the 3 retries add latency before `authLoading` becomes false.

Verify: `useCampaigns` in `dashboard/page.tsx` (line 31) is called unconditionally — it doesn't depend on `authLoading`. So campaigns ARE fetched in parallel with auth. The perceived delay is from React Query retries.

Skip to Task 3 — the retry fix is the real bottleneck here.

- [ ] **Step 2: Commit (no-op — document finding)**

No code change needed. The campaigns query already fires in parallel with auth. The 12s delay is explained by React Query retry behavior (Task 3) and container cold start.

---

### Task 3: Fix React Query retry behavior for fast failure

React Query defaults to 3 retries with exponential backoff. On a cold container start, the first `getMe()` request may take 10+ seconds. If it fails once (timeout), React Query retries 3 more times — each attempt potentially waiting for the container to boot. This turns a single slow request into 12+ seconds.

**Files:**
- Modify: `frontend/src/app/providers.tsx`

- [ ] **Step 1: Set `retry: 1` in React Query defaults**

In `frontend/src/app/providers.tsx`, update the QueryClient config:

```typescript
new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 2 * 60 * 1000, // 2 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes
      refetchOnWindowFocus: false,
      refetchOnReconnect: false,
      retry: 1,
    },
  },
})
```

This gives one retry (for transient network blips) but won't compound the cold-start delay with 3 additional attempts.

- [ ] **Step 2: Verify frontend builds**

```bash
cd frontend && npm run build
```

- [ ] **Step 3: Run E2E tests**

```bash
make test-e2e
```

All existing tests must pass — the retry change should be transparent to tests since the test backend starts instantly.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/providers.tsx
git commit -m "perf: reduce React Query retries from 3 to 1 for faster cold-start UX"
```

---

### Task 4: Run full pre-commit validation

- [ ] **Step 1: Run the pre-commit checks**

```bash
cd backend && go vet ./... && go build ./... && go test ./... -short -count=1
cd ../frontend && npm run lint && npm run build
```

- [ ] **Step 2: Run E2E tests**

```bash
make test-e2e
```

All campaign creation, navigation, and board tests must pass.

- [ ] **Step 3: Create PR**

```bash
git push -u origin feat/performance-batch-inserts
gh pr create --title "perf: batch template INSERTs and reduce retry delays" --body "..."
```
