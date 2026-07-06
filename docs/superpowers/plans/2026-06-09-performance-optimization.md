# Performance Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Cut campaign page load latency (currently 6 sequential Neon round trips) and eliminate the full-campaign refetch that fires after every task/subtask interaction.

**Architecture:** Two independent tracks. Backend: parallelize the 5 hierarchy queries in `GetFullCampaign` with errgroup, run the membership check concurrently in the handler, and collapse the two-round-trip `CreateTask`/`CreateSubtask` into single INSERTs. Frontend: replace blanket `invalidateQueries(["campaign", id])` after every mutation with targeted React Query cache patches using the objects the API already returns, so a status toggle or subtask check no longer re-downloads the entire campaign hierarchy.

**Tech Stack:** Go 1.25 / pgx / errgroup, React 19 / React Query (TanStack), Next.js static export on Cloudflare Workers + Containers, Neon PostgreSQL.

---

## Audit Context (2026-06-09)

What this plan is based on — verified directly in source, not just agent reports:

| Finding | Evidence | Cost today |
|---|---|---|
| `GetFullCampaign` runs 5 sequential queries (campaign → lists → groups → tasks → subtasks), each keyed only on `campaignID` | `backend/internal/store/campaign.go:157-299` | 5 × Neon round trip (5–50ms each) on every campaign page load |
| `handleGetCampaign` runs `IsCampaignMember` before `GetFullCampaign`, sequentially | `backend/internal/handler/campaign.go:60-77` | +1 round trip → 6 total per page load |
| `CreateTask`/`CreateSubtask` do `SELECT MAX(position)` then `INSERT` (2 round trips, plus a race) | `backend/internal/store/task.go:19-41,108-129` | 2× latency on every task/subtask creation |
| Every TaskDetail interaction (name blur, status, due date, assignee, description save, subtask add/toggle/delete) calls `onUpdate()` → `invalidateQueries(["campaign", id])` + `(["campaigns"])` → refetches the full hierarchy | `frontend/src/app/campaign/[id]/campaign-board.tsx:225-236`, `task-detail.tsx`, `subtask-item.tsx` | Full campaign GET (the 6-round-trip endpoint above) after *every* click |
| Board status change does an optimistic update, then immediately invalidates and refetches anyway | `campaign-board.tsx:86-87` | Optimistic update's benefit thrown away |
| `useCampaigns` overrides default staleTime with `staleTime: 0` | `frontend/src/hooks/use-campaign.ts:9` | Dashboard refetches list on every mount |
| `TaskGroup` and `TaskListTabs` not memoized; `handleStatusChange` recreated every render | `task-group.tsx:21`, `task-list-tabs.tsx:11` | Whole board re-renders on any state change (hideDone, selection) |
| `allTasks` flatten not memoized on calendar page | `calendar-page.tsx:38-40` | Re-flattens every render (minor) |
| pgx pool `MinConns = 0`; bare `http.ListenAndServe` (no timeouts) | `store.go:26`, `handler.go:23-25` | Connection handshake on first request after idle; no slow-client protection |
| No long-lived cache headers for hashed `/_next/static/*` assets | `wrangler.toml [assets]`, no `_headers` file | Browser revalidates immutable assets |

**Checked and ruled out** (do not "fix" these):
- Auth middleware does **not** hit the DB — sessions are gorilla/sessions *cookie* store (`auth.go:196-218`). No session-table round trip exists.
- `campaign_members(campaign_id)` index is **not** missing — the composite PK `(campaign_id, user_id)` covers it (`migrations/001_initial.sql:24`).
- All FK indexes exist (001 + 004 + 005 migrations). No new indexes needed.
- The April 2026 perf plan (batched template INSERTs, `retry: 1`, parallel auth/campaign fetch) is already implemented and merged.
- `key={selectedTask.id}` on TaskDetail causes a remount per task switch, but it powers the deliberate fade animation from PR #32 — **keep it**.
- Adding `Cache-Control` to API GET responses was considered and rejected: browser-level caching would serve stale data to React Query's post-mutation refetches.

**Frontend caveat:** `frontend/AGENTS.md` warns this Next.js version has breaking changes vs. training data. This plan touches no Next.js APIs (only React Query, React, and component code). If you need to touch routing/config, read `node_modules/next/dist/docs/` first.

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `backend/internal/store/campaign.go:157-300` | Modify | Parallelize `GetFullCampaign`'s 5 queries with errgroup |
| `backend/internal/handler/campaign.go:60-77` | Modify | Run membership check concurrently with `GetFullCampaign` |
| `backend/internal/store/task.go:19-41,108-129` | Modify | Single-round-trip `CreateTask` / `CreateSubtask` |
| `backend/internal/store/store.go:26` | Modify | `MinConns = 2` (warm connections) |
| `backend/internal/handler/handler.go:23-25` | Modify | `http.Server` with timeouts |
| `backend/internal/handler/routes.go:97-100` | Modify | Cache CORS preflight (dev-only benefit) |
| `backend/go.mod` | Modify | Promote `golang.org/x/sync` to direct dependency |
| `frontend/src/lib/campaign-cache.ts` | **Create** | React Query cache patch helpers for the campaign hierarchy |
| `frontend/src/hooks/use-campaign.ts:9` | Modify | Remove `staleTime: 0` override |
| `frontend/src/app/campaign/[id]/campaign-board.tsx` | Modify | `selectedTaskId` state, stable callbacks, patch instead of invalidate |
| `frontend/src/components/task-detail.tsx` | Modify | Take `campaignId`, patch cache with API responses, drop `onUpdate` |
| `frontend/src/components/subtask-item.tsx` | Modify | Take `campaignId`, patch cache, drop `onUpdate` |
| `frontend/src/components/task-group.tsx` | Modify | Patch cache on create; wrap in `memo` |
| `frontend/src/components/task-list-tabs.tsx` | Modify | Wrap in `memo` |
| `frontend/src/app/campaign/[id]/calendar/calendar-page.tsx` | Modify | Memoize flatten, `selectedTaskId`, new TaskDetail props |
| `frontend/public/_headers` | **Create** | Immutable caching for hashed static assets |
| `src/index.ts:18` | Modify (optional) | `sleepAfter` bump — see Task 8 |

---

### Task 0: Branch setup

- [ ] **Step 1: Create a feature branch from fresh main** (CLAUDE.md workflow: never push to main)

```bash
git checkout main && git pull
git checkout -b perf/parallel-queries-and-cache-patches
```

---

### Task 1: Parallelize `GetFullCampaign` (6 round trips → 1 round-trip wall time)

All five hierarchy queries filter on `campaign_id` alone — none depends on another's results (only the in-memory assembly does). Run them concurrently on the pool, then assemble. Also run the handler's membership check in parallel.

**Files:**
- Modify: `backend/internal/store/campaign.go:157-300`
- Modify: `backend/internal/handler/campaign.go:60-77`
- Modify: `backend/go.mod`
- Test: existing `backend/internal/store/campaign_test.go` (`TestGetFullCampaign`)

- [ ] **Step 1: Promote errgroup to a direct dependency**

```bash
cd backend && go get golang.org/x/sync@v0.17.0
```

(`golang.org/x/sync v0.17.0` is already in `go.mod` as indirect; this just promotes it.)

- [ ] **Step 2: Replace `GetFullCampaign` in `backend/internal/store/campaign.go` (lines 157-300)**

Add `"golang.org/x/sync/errgroup"` to the imports, then replace the whole function with:

```go
func (s *Store) GetFullCampaign(ctx context.Context, campaignID string) (*model.Campaign, error) {
	var (
		campaign model.Campaign
		lists    []model.TaskList
		groups   []model.TaskGroup
		tasks    []model.Task
		subtasks []model.Subtask
	)

	// All five queries filter on campaign_id alone, so they can run
	// concurrently; only the in-memory assembly below is order-dependent.
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := s.pool.QueryRow(gctx, `
			SELECT id, created_by, name, archived, template_type, release_date::text, schedule_weeks, created_at, updated_at
			FROM campaigns WHERE id = $1
		`, campaignID).Scan(
			&campaign.ID, &campaign.CreatedBy, &campaign.Name, &campaign.Archived,
			&campaign.TemplateType, &campaign.ReleaseDate, &campaign.ScheduleWeeks,
			&campaign.CreatedAt, &campaign.UpdatedAt,
		)
		if err == pgx.ErrNoRows {
			return fmt.Errorf("campaign not found")
		}
		return err
	})

	g.Go(func() error {
		rows, err := s.pool.Query(gctx, `
			SELECT id, campaign_id, name, color, position
			FROM task_lists WHERE campaign_id = $1 ORDER BY position
		`, campaignID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var tl model.TaskList
			if err := rows.Scan(&tl.ID, &tl.CampaignID, &tl.Name, &tl.Color, &tl.Position); err != nil {
				return err
			}
			lists = append(lists, tl)
		}
		return rows.Err()
	})

	g.Go(func() error {
		rows, err := s.pool.Query(gctx, `
			SELECT tg.id, tg.task_list_id, tg.name, tg.position, tg.collapsed
			FROM task_groups tg
			JOIN task_lists tl ON tl.id = tg.task_list_id
			WHERE tl.campaign_id = $1
			ORDER BY tg.position
		`, campaignID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var tg model.TaskGroup
			if err := rows.Scan(&tg.ID, &tg.TaskListID, &tg.Name, &tg.Position, &tg.Collapsed); err != nil {
				return err
			}
			groups = append(groups, tg)
		}
		return rows.Err()
	})

	g.Go(func() error {
		rows, err := s.pool.Query(gctx, `
			SELECT t.id, t.task_group_id, t.name, t.description, t.status, t.due_date::text, t.assigned_to, t.position, t.created_at, t.updated_at
			FROM tasks t
			JOIN task_groups tg ON tg.id = t.task_group_id
			JOIN task_lists tl ON tl.id = tg.task_list_id
			WHERE tl.campaign_id = $1
			ORDER BY t.position
		`, campaignID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var t model.Task
			if err := rows.Scan(&t.ID, &t.TaskGroupID, &t.Name, &t.Description, &t.Status, &t.DueDate, &t.AssignedTo, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
				return err
			}
			tasks = append(tasks, t)
		}
		return rows.Err()
	})

	g.Go(func() error {
		rows, err := s.pool.Query(gctx, `
			SELECT st.id, st.task_id, st.name, st.is_complete, st.position
			FROM subtasks st
			JOIN tasks t ON t.id = st.task_id
			JOIN task_groups tg ON tg.id = t.task_group_id
			JOIN task_lists tl ON tl.id = tg.task_list_id
			WHERE tl.campaign_id = $1
			ORDER BY st.position
		`, campaignID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var st model.Subtask
			if err := rows.Scan(&st.ID, &st.TaskID, &st.Name, &st.IsComplete, &st.Position); err != nil {
				return err
			}
			subtasks = append(subtasks, st)
		}
		return rows.Err()
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Assemble the hierarchy (same logic as before, from the collected slices).
	listMap := map[string]int{} // list ID -> index
	for _, tl := range lists {
		listMap[tl.ID] = len(campaign.TaskLists)
		campaign.TaskLists = append(campaign.TaskLists, tl)
	}

	groupMap := map[string]int{}     // group ID -> index in parent list
	groupList := map[string]string{} // group ID -> list ID
	for _, tg := range groups {
		li := listMap[tg.TaskListID]
		groupMap[tg.ID] = len(campaign.TaskLists[li].TaskGroups)
		groupList[tg.ID] = tg.TaskListID
		campaign.TaskLists[li].TaskGroups = append(campaign.TaskLists[li].TaskGroups, tg)
	}

	taskMap := map[string]int{}      // task ID -> index in parent group
	taskGroup := map[string]string{} // task ID -> group ID
	for _, t := range tasks {
		listID := groupList[t.TaskGroupID]
		li := listMap[listID]
		gi := groupMap[t.TaskGroupID]
		taskMap[t.ID] = len(campaign.TaskLists[li].TaskGroups[gi].Tasks)
		taskGroup[t.ID] = t.TaskGroupID
		campaign.TaskLists[li].TaskGroups[gi].Tasks = append(campaign.TaskLists[li].TaskGroups[gi].Tasks, t)
	}

	for _, st := range subtasks {
		gID := taskGroup[st.TaskID]
		listID := groupList[gID]
		li := listMap[listID]
		gi := groupMap[gID]
		ti := taskMap[st.TaskID]
		campaign.TaskLists[li].TaskGroups[gi].Tasks[ti].Subtasks = append(
			campaign.TaskLists[li].TaskGroups[gi].Tasks[ti].Subtasks, st,
		)
	}

	return &campaign, nil
}
```

- [ ] **Step 3: Run the membership check concurrently in `handleGetCampaign`**

In `backend/internal/handler/campaign.go`, replace `handleGetCampaign` (lines 60-77) with:

```go
func (s *Server) handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	userID := auth.GetUserID(r)

	// Membership check and data fetch are independent reads — run them
	// concurrently and enforce membership before writing anything out.
	var (
		wg          sync.WaitGroup
		member      bool
		memberErr   error
		campaign    *model.Campaign
		campaignErr error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		member, memberErr = s.store.IsCampaignMember(r.Context(), campaignID, userID)
	}()
	go func() {
		defer wg.Done()
		campaign, campaignErr = s.store.GetFullCampaign(r.Context(), campaignID)
	}()
	wg.Wait()

	if memberErr != nil || !member {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	if campaignErr != nil {
		log.Printf("Failed to get campaign %s: %v", campaignID, campaignErr)
		writeError(w, http.StatusNotFound, "Campaign not found")
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}
```

Add `"sync"` to the imports in `backend/internal/handler/campaign.go`.

- [ ] **Step 4: Verify build and vet**

```bash
cd backend && go build ./... && go vet ./...
```

Expected: no output, exit 0.

- [ ] **Step 5: Run store tests (integration tests run if DATABASE_URL is set, otherwise unit suite)**

```bash
cd backend && go test ./... -count=1
cd backend && go test -tags=integration -run 'TestGetFullCampaign|TestCreateCampaign' -count=1 ./internal/store/ -v
```

Expected: PASS (or SKIP for integration tests without `DATABASE_URL`). `TestGetFullCampaign` verifies the full hierarchy loads with correct nesting — it must pass unchanged, since this task must not alter the response shape.

- [ ] **Step 6: Commit**

```bash
git add backend/go.mod backend/go.sum backend/internal/store/campaign.go backend/internal/handler/campaign.go
git commit -m "perf: parallelize GetFullCampaign queries and membership check — 6 round trips to ~1 wall-time"
```

---

### Task 2: Single round-trip `CreateTask` / `CreateSubtask`

Each currently does `SELECT MAX(position)` then `INSERT` (2 round trips; the MAX error is also silently ignored, and two concurrent creates can race to the same position). Compute the position inside the INSERT.

**Files:**
- Modify: `backend/internal/store/task.go:19-41` and `:108-129`
- Test: existing `backend/internal/store/task_test.go`

- [ ] **Step 1: Replace `CreateTask` (task.go lines 19-41) with:**

```go
func (s *Store) CreateTask(ctx context.Context, groupID, name string) (*model.Task, error) {
	var task model.Task
	err := s.pool.QueryRow(ctx, `
		INSERT INTO tasks (task_group_id, name, position)
		SELECT $1::uuid, $2::text, COALESCE(MAX(position), 0) + 100
		FROM tasks WHERE task_group_id = $1
		RETURNING id, task_group_id, name, description, status, due_date::text, assigned_to, position, created_at, updated_at
	`, groupID, name).Scan(&task.ID, &task.TaskGroupID, &task.Name, &task.Description,
		&task.Status, &task.DueDate, &task.AssignedTo, &task.Position, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}
```

(`MAX` over an empty set yields one row with NULL, so the SELECT always produces exactly one row; first insert gets position 100, matching current behavior.)

- [ ] **Step 2: Replace `CreateSubtask` (task.go lines 108-129) with:**

```go
func (s *Store) CreateSubtask(ctx context.Context, taskID, name string) (*model.Subtask, error) {
	var subtask model.Subtask
	err := s.pool.QueryRow(ctx, `
		INSERT INTO subtasks (task_id, name, position)
		SELECT $1::uuid, $2::text, COALESCE(MAX(position), 0) + 100
		FROM subtasks WHERE task_id = $1
		RETURNING id, task_id, name, is_complete, position
	`, taskID, name).Scan(&subtask.ID, &subtask.TaskID, &subtask.Name, &subtask.IsComplete, &subtask.Position)
	if err != nil {
		return nil, err
	}
	return &subtask, nil
}
```

- [ ] **Step 3: Build, vet, and run task store tests**

```bash
cd backend && go build ./... && go vet ./... && go test ./internal/store/ -count=1 -v
```

Expected: PASS (integration tests skip without `DATABASE_URL`). If `task_test.go` has integration tests for create/position ordering, run them:

```bash
cd backend && go test -tags=integration -run 'TestCreateTask|TestCreateSubtask' -count=1 ./internal/store/ -v
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/store/task.go
git commit -m "perf: compute task/subtask position in the INSERT — 2 round trips to 1, removes position race"
```

---

### Task 3: Backend plumbing — warm pool connections, server timeouts, preflight cache

**Files:**
- Modify: `backend/internal/store/store.go:26`
- Modify: `backend/internal/handler/handler.go:23-25`
- Modify: `backend/internal/handler/routes.go:97-100`

- [ ] **Step 1: Keep two connections warm.** In `backend/internal/store/store.go`, change line 26:

```go
	config.MaxConns = 25
	config.MinConns = 2 // keep warm connections so the first request after idle skips the TLS+auth handshake to Neon
```

- [ ] **Step 2: Add server timeouts.** In `backend/internal/handler/handler.go`, replace `Start` (lines 23-25) with:

```go
func (s *Server) Start(port string) error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return srv.ListenAndServe()
}
```

Add `"time"` to the imports in `handler.go`.

- [ ] **Step 3: Cache CORS preflights.** In `backend/internal/handler/routes.go`, inside `corsMiddleware`, replace the OPTIONS block (lines 97-100) with:

```go
		if r.Method == "OPTIONS" {
			// Cache preflights; only matters in dev where the frontend origin differs.
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}
```

(In production the frontend and API share `release.rdelpret.com`, so no preflights occur — this removes a round trip per mutation type in local dev only.)

- [ ] **Step 4: Build, vet, test**

```bash
cd backend && go build ./... && go vet ./... && go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/store/store.go backend/internal/handler/handler.go backend/internal/handler/routes.go
git commit -m "perf: warm pool connections, http server timeouts, preflight caching"
```

---

### Task 4: Frontend — patch the React Query cache instead of refetching the whole campaign

Every task mutation API call already returns the updated object (`updateTask → Task`, `createTask → Task`, `createSubtask/updateSubtask → Subtask` — see `frontend/src/lib/api.ts`). Write those into the cache instead of invalidating `["campaign", id]`. Keep `["campaigns"]` invalidation where dashboard counts change (status, create, delete, due date) — that list query is cheap and only refetches when the dashboard mounts. Keep full invalidation only where the server recomputes data we don't get back: release-date/schedule changes and drag-drop reorders.

**Files:**
- Create: `frontend/src/lib/campaign-cache.ts`
- Modify: `frontend/src/hooks/use-campaign.ts:5-11`
- Modify: `frontend/src/app/campaign/[id]/campaign-board.tsx`
- Modify: `frontend/src/components/task-detail.tsx`
- Modify: `frontend/src/components/subtask-item.tsx`
- Modify: `frontend/src/components/task-group.tsx`
- Modify: `frontend/src/app/campaign/[id]/calendar/calendar-page.tsx`
- Test: `frontend/e2e/app.spec.ts` (existing E2E suite, runs in CI)

- [ ] **Step 1: Create `frontend/src/lib/campaign-cache.ts`:**

```typescript
import type { QueryClient } from "@tanstack/react-query";
import type { Campaign, Task, Subtask } from "./types";

// Targeted cache updates for the ["campaign", id] query, so small mutations
// don't trigger a refetch of the entire campaign hierarchy.

function mapTasks(campaign: Campaign, fn: (task: Task) => Task | null): Campaign {
  return {
    ...campaign,
    task_lists: campaign.task_lists?.map((l) => ({
      ...l,
      task_groups: l.task_groups?.map((g) => ({
        ...g,
        tasks: (g.tasks ?? []).map(fn).filter((t): t is Task => t !== null),
      })),
    })),
  };
}

export function patchTask(queryClient: QueryClient, campaignId: string, taskId: string, patch: Partial<Task>) {
  queryClient.setQueryData<Campaign>(["campaign", campaignId], (old) =>
    old
      ? mapTasks(old, (t) =>
          t.id === taskId ? { ...t, ...patch, subtasks: patch.subtasks ?? t.subtasks } : t
        )
      : old
  );
}

export function addTask(queryClient: QueryClient, campaignId: string, groupId: string, task: Task) {
  queryClient.setQueryData<Campaign>(["campaign", campaignId], (old) =>
    old
      ? {
          ...old,
          task_lists: old.task_lists?.map((l) => ({
            ...l,
            task_groups: l.task_groups?.map((g) =>
              g.id === groupId ? { ...g, tasks: [...(g.tasks ?? []), task] } : g
            ),
          })),
        }
      : old
  );
}

export function removeTask(queryClient: QueryClient, campaignId: string, taskId: string) {
  queryClient.setQueryData<Campaign>(["campaign", campaignId], (old) =>
    old ? mapTasks(old, (t) => (t.id === taskId ? null : t)) : old
  );
}

export function addSubtask(queryClient: QueryClient, campaignId: string, taskId: string, subtask: Subtask) {
  queryClient.setQueryData<Campaign>(["campaign", campaignId], (old) =>
    old
      ? mapTasks(old, (t) =>
          t.id === taskId ? { ...t, subtasks: [...(t.subtasks ?? []), subtask] } : t
        )
      : old
  );
}

export function patchSubtask(queryClient: QueryClient, campaignId: string, subtask: Subtask) {
  queryClient.setQueryData<Campaign>(["campaign", campaignId], (old) =>
    old
      ? mapTasks(old, (t) =>
          t.id === subtask.task_id
            ? { ...t, subtasks: t.subtasks?.map((st) => (st.id === subtask.id ? subtask : st)) }
            : t
        )
      : old
  );
}

export function removeSubtask(queryClient: QueryClient, campaignId: string, subtaskId: string) {
  queryClient.setQueryData<Campaign>(["campaign", campaignId], (old) =>
    old
      ? mapTasks(old, (t) =>
          t.subtasks?.some((st) => st.id === subtaskId)
            ? { ...t, subtasks: t.subtasks.filter((st) => st.id !== subtaskId) }
            : t
        )
      : old
  );
}
```

(The explicit `subtasks: patch.subtasks ?? t.subtasks` in `patchTask` matters: the backend's task JSON can carry `subtasks: null`, which a bare spread would write over the cached subtasks.)

- [ ] **Step 2: Remove the `staleTime: 0` override.** In `frontend/src/hooks/use-campaign.ts`, change `useCampaigns` (lines 5-11) to:

```typescript
export function useCampaigns() {
  return useQuery<Campaign[]>({
    queryKey: ["campaigns"],
    queryFn: api.listCampaigns,
  });
}
```

(Inherits the 2-minute default from `providers.tsx`; mutations that change counts explicitly invalidate `["campaigns"]`, so the dashboard stays correct.)

- [ ] **Step 3: Rework `campaign-board.tsx`.**

3a. Replace the `selectedTask` state (line 40) with an ID + derived lookup, and add stable callbacks. Replace lines 40-41:

```tsx
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const { handleDragEnd } = useTaskDragDrop(id);
```

3b. Immediately after the `activeList` derivation block (after line 54's auto-select `if`, before the `isLoading` early return), add:

```tsx
  // Derive the selected task from the cache so TaskDetail always renders
  // fresh data after cache patches, with no manual re-sync.
  const selectedTask = useMemo(() => {
    if (!selectedTaskId) return null;
    for (const l of campaign?.task_lists ?? [])
      for (const g of l.task_groups ?? [])
        for (const t of g.tasks ?? [])
          if (t.id === selectedTaskId) return t;
    return null;
  }, [campaign, selectedTaskId]);

  const handleSelectTask = useCallback((task: Task) => setSelectedTaskId(task.id), []);

  const handleStatusChange = useCallback(
    async (taskId: string, status: Task["status"]) => {
      const prev = queryClient.getQueryData<Campaign>(["campaign", id]);
      patchTask(queryClient, id, taskId, { status }); // optimistic
      try {
        const updated = await updateTask(taskId, { status });
        patchTask(queryClient, id, updated.id, updated); // settle with server truth
        queryClient.invalidateQueries({ queryKey: ["campaigns"] }); // dashboard done-counts
      } catch (err: any) {
        if (prev) queryClient.setQueryData(["campaign", id], prev); // rollback
        toast.error(err.message);
      }
    },
    [id, queryClient]
  );
```

3c. Delete the old `handleStatusChange` function (currently lines 64-93, defined after the early return) entirely — including its inline optimistic-update block, the `setSelectedTask` sync, and the two `invalidateQueries` calls.

3d. Update imports at the top of the file:

```tsx
import { useState, useMemo, useCallback, lazy, Suspense } from "react";
import { patchTask } from "@/lib/campaign-cache";
```

3e. Pass the stable callback to TaskGroup. In the `TaskGroup` render (lines 202-212), change `onSelectTask={setSelectedTask}` to `onSelectTask={handleSelectTask}` (the `onStatusChange={handleStatusChange}` prop stays as-is, now referencing the `useCallback` version).

3f. Replace the TaskDetail block (lines 219-239) with:

```tsx
      {selectedTask && (
        <Suspense fallback={null}>
          <TaskDetail
            key={selectedTask.id}
            task={selectedTask}
            campaignId={id}
            onClose={() => setSelectedTaskId(null)}
          />
        </Suspense>
      )}
```

(`key` stays — it drives the intentional card-switch fade from PR #32. The release-date and schedule-week handlers, lines 137-176, keep their `invalidateQueries({ queryKey: ["campaign", id] })`: the server recomputes every task's due date and we don't get them back in the response.)

- [ ] **Step 4: Rework `task-detail.tsx` to patch the cache.** Replace the props interface and every handler. Full new version of the logic section (render JSX below the handlers is unchanged except where noted):

```tsx
"use client";

import { useState } from "react";
import type { Task } from "@/lib/types";
import { SubtaskItem } from "./subtask-item";
import { X, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { updateTask, deleteTask, createSubtask } from "@/lib/api";
import { toast } from "sonner";
import { RichTextEditor } from "./rich-text-editor";
import { useUsers } from "@/hooks/use-campaign";
import { useQueryClient } from "@tanstack/react-query";
import { patchTask, removeTask, addSubtask } from "@/lib/campaign-cache";

const statusOptions = [
  { value: "todo", label: "To Do", color: "text-text-muted" },
  { value: "in_progress", label: "In Progress", color: "text-yellow-400" },
  { value: "done", label: "Done", color: "text-green-400" },
] as const;

interface TaskDetailProps {
  task: Task;
  campaignId: string;
  onClose: () => void;
}

export function TaskDetail({ task, campaignId, onClose }: TaskDetailProps) {
  const [name, setName] = useState(task.name);
  const [status, setStatus] = useState(task.status);
  const [dueDate, setDueDate] = useState(task.due_date ?? "");
  const [assignedTo, setAssignedTo] = useState(task.assigned_to ?? "");
  const [newSubtaskName, setNewSubtaskName] = useState("");
  const { data: users } = useUsers();
  const queryClient = useQueryClient();

  const handleNameBlur = async () => {
    if (name !== task.name && name.trim()) {
      try {
        const updated = await updateTask(task.id, { name: name.trim() } as any);
        patchTask(queryClient, campaignId, updated.id, updated);
      } catch (err: any) {
        toast.error(err.message);
      }
    }
  };

  const handleStatusChange = async (newStatus: string) => {
    const prev = status;
    setStatus(newStatus as Task["status"]);
    try {
      const updated = await updateTask(task.id, { status: newStatus } as any);
      patchTask(queryClient, campaignId, updated.id, updated);
      queryClient.invalidateQueries({ queryKey: ["campaigns"] });
    } catch (err: any) {
      setStatus(prev);
      toast.error(err.message);
    }
  };

  const handleDueDateChange = async (date: string) => {
    setDueDate(date);
    try {
      const updated = await updateTask(task.id, { due_date: date || undefined } as any);
      patchTask(queryClient, campaignId, updated.id, updated);
      queryClient.invalidateQueries({ queryKey: ["campaigns"] }); // overdue counts
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  const handleAssign = async (userId: string) => {
    const prev = assignedTo;
    setAssignedTo(userId);
    try {
      const updated = await updateTask(task.id, { assigned_to: userId || undefined } as any);
      patchTask(queryClient, campaignId, updated.id, updated);
    } catch (err: any) {
      setAssignedTo(prev);
      toast.error(err.message);
    }
  };

  const handleAddSubtask = async () => {
    if (!newSubtaskName.trim()) return;
    try {
      const subtask = await createSubtask(task.id, newSubtaskName.trim());
      setNewSubtaskName("");
      addSubtask(queryClient, campaignId, task.id, subtask);
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Delete this task?")) return;
    try {
      await deleteTask(task.id);
      removeTask(queryClient, campaignId, task.id);
      queryClient.invalidateQueries({ queryKey: ["campaigns"] });
      onClose();
    } catch (err: any) {
      toast.error(err.message);
    }
  };
```

In the JSX, two changes:
- the RichTextEditor `onUpdate` becomes:

```tsx
        <RichTextEditor
          content={task.description ?? null}
          onUpdate={async (content) => {
            try {
              const updated = await updateTask(task.id, { description: content } as any);
              patchTask(queryClient, campaignId, updated.id, updated);
            } catch (err: any) {
              toast.error(err.message);
            }
          }}
        />
```

- the subtask list becomes:

```tsx
          {(task.subtasks ?? []).map((subtask) => (
            <SubtaskItem key={subtask.id} subtask={subtask} campaignId={campaignId} />
          ))}
```

- [ ] **Step 5: Rework `subtask-item.tsx`.** Replace the file's logic (render JSX unchanged):

```tsx
"use client";

import { useState } from "react";
import type { Subtask } from "@/lib/types";
import { Trash2 } from "lucide-react";
import { updateSubtask, deleteSubtask } from "@/lib/api";
import { useQueryClient } from "@tanstack/react-query";
import { patchSubtask, removeSubtask } from "@/lib/campaign-cache";
import { toast } from "sonner";

interface SubtaskItemProps {
  subtask: Subtask;
  campaignId: string;
}

export function SubtaskItem({ subtask, campaignId }: SubtaskItemProps) {
  const queryClient = useQueryClient();
  const [optimisticComplete, setOptimisticComplete] = useState(subtask.is_complete);
  const [deleting, setDeleting] = useState(false);

  // Sync optimistic state when the cached prop changes underneath us
  if (optimisticComplete !== subtask.is_complete && !deleting) {
    setOptimisticComplete(subtask.is_complete);
  }

  const handleToggle = async () => {
    const newVal = !optimisticComplete;
    setOptimisticComplete(newVal); // optimistic
    try {
      const updated = await updateSubtask(subtask.id, { is_complete: newVal });
      patchSubtask(queryClient, campaignId, updated);
    } catch (err: any) {
      setOptimisticComplete(!newVal); // rollback
      toast.error(err.message);
    }
  };

  const handleDelete = async () => {
    setDeleting(true); // hide immediately
    try {
      await deleteSubtask(subtask.id);
      removeSubtask(queryClient, campaignId, subtask.id);
    } catch (err: any) {
      setDeleting(false); // rollback
      toast.error(err.message);
    }
  };

  if (deleting) return null;
  // ... existing JSX unchanged ...
}
```

- [ ] **Step 6: Rework `task-group.tsx` `handleAddTask` (lines 31-43):**

```tsx
  const handleAddTask = async () => {
    if (!newTaskName.trim()) return;
    try {
      const task = await createTask(group.id, newTaskName.trim());
      setNewTaskName("");
      setAdding(false);
      addTask(queryClient, campaignId, group.id, task);
      queryClient.invalidateQueries({ queryKey: ["campaigns"] }); // dashboard total counts
      onSelectTask(task);
    } catch (err: any) {
      toast.error(err.message);
    }
  };
```

Add the import: `import { addTask } from "@/lib/campaign-cache";`

- [ ] **Step 7: Rework `calendar-page.tsx`.** Replace the state/derivation block (lines 27 and 37-40) and the TaskDetail usage (lines 59-65):

```tsx
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);

  // Flatten all tasks from all lists/groups (memoized — runs only when data changes)
  const allTasks = useMemo(
    () =>
      (campaign?.task_lists ?? [])
        .flatMap((l) => l.task_groups ?? [])
        .flatMap((g) => g.tasks ?? []),
    [campaign?.task_lists]
  );
  const selectedTask = useMemo(
    () => allTasks.find((t) => t.id === selectedTaskId) ?? null,
    [allTasks, selectedTaskId]
  );
  const handleSelectTask = useCallback((task: Task) => setSelectedTaskId(task.id), []);
```

Both `useMemo`s and the `useCallback` must sit **above** the `if (isLoading || !campaign)` early return (hooks can't be conditional). Update the render:

```tsx
      <CalendarView tasks={allTasks} onSelectTask={handleSelectTask} />

      {selectedTask && (
        <TaskDetail
          task={selectedTask}
          campaignId={id}
          onClose={() => setSelectedTaskId(null)}
        />
      )}
```

Update imports: add `useCallback` to the react import; remove `useQueryClient` import and the `const queryClient = useQueryClient();` line (no longer used).

- [ ] **Step 8: Lint and build**

```bash
cd frontend && npm run lint && npm run build
```

Expected: clean lint, successful production build. Fix any TypeScript errors (likely spots: unused imports left behind in campaign-board/calendar-page).

- [ ] **Step 9: Commit**

```bash
git add frontend/src/lib/campaign-cache.ts frontend/src/hooks/use-campaign.ts frontend/src/app/campaign/[id]/campaign-board.tsx frontend/src/components/task-detail.tsx frontend/src/components/subtask-item.tsx frontend/src/components/task-group.tsx frontend/src/app/campaign/[id]/calendar/calendar-page.tsx
git commit -m "perf: patch campaign cache from mutation responses instead of refetching full hierarchy"
```

---

### Task 5: Memoize board components

With Task 4's stable callbacks (`handleSelectTask`, `handleStatusChange` via `useCallback`), `memo` now actually prevents re-renders: toggling hide-done or opening the detail panel no longer re-renders every group and tab. (`TaskItem` is already memoized.)

**Files:**
- Modify: `frontend/src/components/task-group.tsx:21` and `:101`
- Modify: `frontend/src/components/task-list-tabs.tsx:11` and end of component

- [ ] **Step 1: Wrap `TaskGroup` in `memo`.** In `task-group.tsx`, change the react import to `import { memo, useState } from "react";` and change the component declaration/closing:

```tsx
export const TaskGroup = memo(function TaskGroup({ group, campaignId, hideDone, users, onSelectTask, onStatusChange }: TaskGroupProps) {
```

and the final closing brace of the component from `}` to `});`

- [ ] **Step 2: Wrap `TaskListTabs` in `memo`.** In `task-list-tabs.tsx`, add `import { memo } from "react";` and change:

```tsx
export const TaskListTabs = memo(function TaskListTabs({ lists, activeId, onSelect }: TaskListTabsProps) {
```

with the closing brace becoming `});`

(Named imports `{ TaskGroup }` / `{ TaskListTabs }` at call sites keep working — only the export style changes, not the name.)

- [ ] **Step 3: Lint and build**

```bash
cd frontend && npm run lint && npm run build
```

Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/task-group.tsx frontend/src/components/task-list-tabs.tsx
git commit -m "perf: memoize TaskGroup and TaskListTabs to stop whole-board re-renders"
```

---

### Task 6: Immutable cache headers for hashed static assets

Next.js writes content-hashed files under `/_next/static/`; they can be cached forever. Cloudflare Workers static assets honor a `_headers` file in the assets directory, and `next build` copies everything in `frontend/public/` into `frontend/out/`.

**Files:**
- Create: `frontend/public/_headers`

- [ ] **Step 1: Create `frontend/public/_headers` with exactly:**

```
/_next/static/*
  Cache-Control: public, max-age=31536000, immutable
```

- [ ] **Step 2: Verify it lands in the export**

```bash
cd frontend && npm run build && cat out/_headers
```

Expected output: the two lines above.

- [ ] **Step 3: Commit**

```bash
git add frontend/public/_headers
git commit -m "perf: immutable cache headers for hashed static assets"
```

---

### Task 7 (OPTIONAL — cost tradeoff, confirm with Robbie before doing): longer container sleep timer

`src/index.ts:18` sets `sleepAfter = "2h"`. After 2 quiet hours the container sleeps and the next user eats a multi-second cold start (the frontend already shows a "waking" state for this). Bumping to `"8h"` keeps it warm across a workday. **Tradeoff:** Cloudflare Containers bill for running time, so this directly increases cost. Skip this task unless Robbie confirms.

- [ ] **Step 1 (if confirmed): change `src/index.ts` line 18:**

```typescript
  sleepAfter = "8h";
```

- [ ] **Step 2: Commit**

```bash
git add src/index.ts
git commit -m "perf: keep backend container warm for 8h to reduce cold starts"
```

---

### Task 8: Full validation and PR

- [ ] **Step 1: Run the full pre-commit-equivalent suite**

```bash
cd backend && go vet ./... && go build ./... && go test ./... -count=1
cd ../frontend && npm run lint && npm run build
```

Expected: all pass.

- [ ] **Step 2: Manual smoke test** (E2E runs in CI, but verify the hot paths locally)

```bash
make dev
```

Then in the browser: open a campaign (loads fast, full hierarchy renders) → toggle a task status (instant, **no full-campaign network request** in devtools — only the PATCH) → check a subtask (same) → add a task (appears + detail opens) → edit description, blur (only PATCH fires) → delete task (panel closes, task gone) → change release date (full refetch IS expected here) → drag a task (refetch expected) → back to dashboard (counts correct).

- [ ] **Step 3: Push and create PR**

```bash
git push -u origin perf/parallel-queries-and-cache-patches
gh pr create --title "perf: parallel campaign queries + targeted cache patches" --body "$(cat <<'EOF'
## Summary
- Parallelize GetFullCampaign's 5 hierarchy queries + membership check (6 sequential Neon round trips → ~1 round-trip wall time on campaign load)
- CreateTask/CreateSubtask: compute position in the INSERT (2 round trips → 1, removes position race)
- Frontend: patch React Query cache from mutation responses instead of refetching the full campaign after every task/subtask interaction
- Memoize TaskGroup/TaskListTabs with stable callbacks
- Warm DB connections, HTTP server timeouts, immutable cache headers for hashed assets

Audit + plan: docs/superpowers/plans/2026-06-09-performance-optimization.md

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 4: Wait for CI (E2E suite) to pass before merge** — branch protection requires it.

---

## Explicitly Not In This Plan (and why)

- **Folding the `IsCampaignMemberVia*` auth check into each mutation's SQL** (would cut 1 round trip from every PATCH/DELETE). Real but smaller win — optimistic UI already hides mutation latency — and it touches ~10 handlers + store methods + tests. Do it as a follow-up if mutations still feel slow after this lands.
- **Optimistic drag-drop reordering** — positions are computed server-side; making this optimistic needs client-side position math and conflict handling. The refetch-on-drop stays for now.
- **Cache-Control on API GET responses** — would fight React Query's invalidation (browser would serve stale JSON to refetches). Rejected.
- **New DB indexes** — audit found none missing (composite PK covers `campaign_members` lookups; all FKs indexed).
- **Removing `key={selectedTask.id}` on TaskDetail** — intentional remount for the fade animation (PR #32).
