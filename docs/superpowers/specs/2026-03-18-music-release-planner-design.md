# Music Release Planner — Design Spec

Internal tool for the Subwave team to plan and execute music release campaigns.

## Context

Subwave (subwave.music) needs a structured way for 2 team members to manage release campaigns. The tool follows a Trello-like model with pre-loaded industry-standard task templates, but purpose-built for music releases.

This is an **internal-only tool** — no public signup, no billing, no multi-tenancy. Just two authenticated Subwave team members managing campaigns.

## Scope (v1)

### In Scope
- Google OAuth (2 whitelisted users, no self-registration)
- Campaigns with pre-loaded 5-list task template
- Task status (To Do / In Progress / Done), due dates, rich text descriptions, subtasks
- Drag-and-drop reordering (tasks, groups, lists)
- Hide completed tasks toggle
- Campaign CRUD (create, duplicate, delete/archive)
- Calendar view (tasks by due date)
- Subwave-branded dark theme (acid green + violet)

### Out of Scope (Later)
- Public signup / billing / Stripe
- Resources / guides library
- Custom task list / group creation
- Real-time WebSocket sync

## Architecture

**Approach:** Full monorepo — Go API backend + Next.js frontend, deployed on Cloudflare.

```
/backend    — Go API (chi router, pgx, raw SQL)
/frontend   — Next.js (App Router, TailwindCSS, shadcn/ui)
Makefile    — Build orchestration
```

### Deployment

```
Cloudflare Workers (TS)
  ├── Serves Next.js static build
  └── Proxies /api/* and /auth/* to Go container

Cloudflare Container (Docker)
  └── Go API server
      └── Connects to Neon PostgreSQL
```

- **Local dev:** Go on :8080, Next.js on :3000 with proxy
- **Migrations:** SQL files via goose or raw psql (`make migrate`)
- **Orchestration:** Makefile (`make dev`, `make build`, `make deploy`)

## Data Model

```sql
User
  id          UUID PRIMARY KEY
  email       TEXT UNIQUE NOT NULL
  name        TEXT NOT NULL
  avatar_url  TEXT
  created_at  TIMESTAMPTZ DEFAULT now()

Campaign
  id          UUID PRIMARY KEY
  created_by  UUID REFERENCES User(id)
  name        TEXT NOT NULL
  archived    BOOLEAN DEFAULT false
  created_at  TIMESTAMPTZ DEFAULT now()
  updated_at  TIMESTAMPTZ DEFAULT now()

CampaignMember
  campaign_id UUID REFERENCES Campaign(id) ON DELETE CASCADE
  user_id     UUID REFERENCES User(id) ON DELETE CASCADE
  role        TEXT CHECK (role IN ('owner', 'editor'))
  PRIMARY KEY (campaign_id, user_id)

TaskList
  id          UUID PRIMARY KEY
  campaign_id UUID REFERENCES Campaign(id) ON DELETE CASCADE
  name        TEXT NOT NULL
  color       TEXT NOT NULL
  position    INTEGER NOT NULL

TaskGroup
  id           UUID PRIMARY KEY
  task_list_id UUID REFERENCES TaskList(id) ON DELETE CASCADE
  name         TEXT NOT NULL
  position     INTEGER NOT NULL
  collapsed    BOOLEAN DEFAULT false

Task
  id             UUID PRIMARY KEY
  task_group_id  UUID REFERENCES TaskGroup(id) ON DELETE CASCADE
  name           TEXT NOT NULL
  description    JSONB           -- Tiptap/ProseMirror rich text
  status         TEXT CHECK (status IN ('todo', 'in_progress', 'done')) DEFAULT 'todo'
  due_date       DATE
  position       INTEGER NOT NULL
  created_at     TIMESTAMPTZ DEFAULT now()
  updated_at     TIMESTAMPTZ DEFAULT now()

Subtask
  id          UUID PRIMARY KEY
  task_id     UUID REFERENCES Task(id) ON DELETE CASCADE
  name        TEXT NOT NULL
  is_complete BOOLEAN DEFAULT false
  position    INTEGER NOT NULL

Session
  id         UUID PRIMARY KEY
  user_id    UUID REFERENCES User(id) ON DELETE CASCADE
  token      TEXT UNIQUE NOT NULL
  expires_at TIMESTAMPTZ NOT NULL
```

### Key Decisions
- **Rich text as JSONB** — Tiptap editor on frontend, stored as ProseMirror JSON in Postgres
- **Position integers** — gap-based ordering (100, 200, 300...) for drag-and-drop with room to insert between
- **CampaignMember** — both users see campaigns they belong to; creator is `owner`, invited user is `editor`
- **Session-based auth** — cookie with session token, not JWT

## Backend (Go API)

### Directory Structure

```
/backend
  /cmd/server/main.go          — entry point
  /internal
    /auth/                      — Google OAuth flow, session management
    /handler/                   — HTTP handlers
    /middleware/                 — auth middleware, CORS
    /model/                     — Go structs
    /store/                     — database queries (pgx, raw SQL)
    /template/                  — default campaign template data
  /migrations/                  — SQL migration files
  Dockerfile
  go.mod
```

### API Routes

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/auth/google` | Initiate OAuth |
| GET | `/auth/google/callback` | OAuth callback, create session |
| POST | `/auth/logout` | Clear session |
| GET | `/api/me` | Current user |
| GET | `/api/campaigns` | List user's campaigns |
| POST | `/api/campaigns` | Create campaign (auto-populates template) |
| POST | `/api/campaigns/:id/duplicate` | Duplicate campaign (duplicator becomes owner, no other members copied) |
| PATCH | `/api/campaigns/:id/archive` | Toggle archived status |
| DELETE | `/api/campaigns/:id` | Permanently delete campaign |
| GET | `/api/campaigns/:id` | Full campaign with all nested data |
| PATCH | `/api/tasks/:id` | Update status, due date, description |
| POST | `/api/task-groups/:id/tasks` | Add task to group |
| DELETE | `/api/tasks/:id` | Delete task |
| PATCH | `/api/tasks/:id/reorder` | Drag-and-drop reorder |
| PATCH | `/api/task-lists/:id/reorder` | Reorder lists |
| PATCH | `/api/task-groups/:id/reorder` | Reorder groups |
| POST | `/api/tasks/:id/subtasks` | Add subtask |
| PATCH | `/api/subtasks/:id` | Toggle/update subtask |
| DELETE | `/api/subtasks/:id` | Delete subtask |

### Auth Flow
1. User clicks "Sign in with Google" -> redirects to `/auth/google`
2. Go server redirects to Google OAuth consent screen
3. Google redirects back to `/auth/google/callback` with code
4. Server exchanges code for token, extracts email
5. If email is in whitelist -> create/find user, create session, set cookie
6. If email is not whitelisted -> reject with 403
7. All `/api/*` routes check session cookie via auth middleware

## Frontend (Next.js)

### Directory Structure

```
/frontend
  /src/app
    /layout.tsx                 — root layout, auth provider, Subwave theme
    /login/page.tsx             — Google OAuth login button
    /dashboard/page.tsx         — campaign list (grid of campaign cards)
    /campaign/[id]/page.tsx     — main campaign board (tabbed task lists)
    /campaign/[id]/calendar/page.tsx — calendar view
  /src/components
    /ui/                        — shadcn/ui components (restyled)
    /campaign-card.tsx          — campaign card for dashboard
    /task-list-tabs.tsx         — tab bar for switching between task lists
    /task-group.tsx             — collapsible group within a list
    /task-item.tsx              — individual task row
    /task-detail.tsx            — slide-out panel (rich text editor, subtasks)
    /subtask-item.tsx           — checkbox subtask row
    /calendar-view.tsx          — calendar with tasks by due date
    /hide-done-toggle.tsx       — toggle to hide completed tasks
  /src/lib
    /api.ts                     — fetch wrapper for Go backend
    /auth.ts                    — auth context/hooks
    /types.ts                   — TypeScript types matching Go models
  /src/hooks
    /use-campaign.ts            — campaign data fetching + mutation
    /use-drag-drop.ts           — drag-and-drop logic
```

### UI Layout

**Tabbed layout (Option C):**
- Dashboard: grid of campaign cards with "+ New Campaign" button
- Campaign view: campaign name at top with back button, Board/Calendar toggle
- Task lists as tabs across the top (color-coded: blue, green, purple, yellow, red)
- Active tab shows that list's groups and tasks vertically
- Clicking a task opens a slide-out detail panel on the right (description, subtasks, due date)

### Key Frontend Decisions
- **@dnd-kit** for drag-and-drop — supports nested sortable lists
- **Tiptap** for rich text editor — ProseMirror-based, stores as JSON
- **Slide-out detail panel** — task editing without leaving the board
- **Optimistic updates** — UI updates immediately, syncs to backend, reverts on failure
- **React Query (TanStack Query)** — cache + revalidation for data freshness

## Theming (Subwave Brand)

Tailwind config using exact tokens from `subwave.music/styles.css`:

### Colors
| Token | Value | Usage |
|-------|-------|-------|
| bg-base | `#0B0D10` | Page background |
| bg-surface | `#15121f` | Cards, panels, tabs |
| text-primary | `#E8EEF6` | Primary text |
| text-muted | `#6B7280` | Secondary text |
| accent | `#B7FF4A` | Primary accent (acid green) |
| accent-dark | `#8FCC3A` | Accent hover/active states |
| secondary | `#A78BFA` | Secondary accent (violet) |
| secondary-dark | `#8B6DF5` | Secondary hover/active |
| accent-glow | `rgba(183, 255, 74, 0.4)` | Hover glow effects |

### Task List Colors
| List | Color |
|------|-------|
| Campaign Assets | `#3B82F6` (blue) |
| Campaign Tasklist | `#22C55E` (green) |
| Social Media Content | `#A78BFA` (purple) |
| PR | `#EAB308` (yellow) |
| Distribution | `#EF4444` (red) |

### Typography
- **Headings:** Space Grotesk
- **Body:** Inter

### Effects
- Glassmorphism: `backdrop-filter: blur(30px) saturate(180%)` on nav/panels
- Accent glow: `rgba(183, 255, 74, 0.4)` box-shadow on hover states
- Transitions: `cubic-bezier(0.4, 0, 0.2, 1)` at 200ms/400ms
- No custom cursor or mouse trail (internal tool, not marketing site)
- shadcn/ui components restyled to match dark theme

## Default Campaign Template

When a user creates a new campaign, these 5 task lists are auto-populated:

### 1. Campaign Assets (Blue)
**Group: "Song Assets"**
Make New Asset Folder, Release Date, Final Master, Final Artwork, Instrumental, Stems, Lyrics, Press Photos, Music Video/Lyric Video, Folder for Promo Content

### 2. Campaign Tasklist (Green)
**Group: "Pre-Release"**
Campaign Setup + Timeline, Contracts, Photoshoot for release, Make content plan, Create content or work with a graphic designer, Create pre-save Link, Pre-save campaign/contest, Create linktree, Write out hashtags, Update website - coming soon, Update all SM profiles - coming soon, Update streaming profiles

### 3. Social Media Content (Purple)
**Group: "Pre-Release Content"**
Teaser post, Banner - release date - presave now, Single art - story sized - release date or 'pre-save now', Single art edit #1 (with release date), Pre-save Post, Song poster #1 - storytelling, Song teaser #1, "Midnight" or "Tomorrow"

**Group: "Post-Release Content"**
Banner - out now, Song Clip - out now, Single art edit #2 ("out now"), Single art - out now - story sized, Song poster #2, Studio content #1 - storytelling, Single art edit #3 - storytelling, Credits page, Lyric poster/Lyric Reel

**Group: "Extra Content"**
Pre-save Post #2, Song clip/teaser #3, Song poster #3, Studio content #2, Tiktok/Reels, Clips from Music video/lyric video, Countdowns, Acoustic version/demo version/live version

### 4. PR (Yellow)
**Group: "Media Outreach"**
Research new outlets & develop CRM, Create press release, Create e-mail outreach pitch, Traditional Media outreach, New Media outreach, Submit to third party platforms, Press follow ups

**Group: "Playlisting"**
Research playlists, Write playlist pitch, Pitch to independent curators, Submit to third party platforms

**Group: "Community Engagement"**
Send DMs on IG/Twitter/FB with link, Post on discord chats/reddit threads, Send mailing list e-mail - presave, Send mailing list e-mail - out now

### 5. Distribution (Red)
**Group: "To Do"**
Obtain license for cover song, Upload to distributor (subtasks: Fill out credits, Upload lyrics), Pitch song to Spotify editors, Upload Spotify Canvas, Register song with PROs, For physical releases

## Error Handling

- **Auth failures:** Redirect to login page with error message
- **API errors:** Toast notifications via shadcn/ui Sonner
- **Optimistic update failures:** Revert UI state, show error toast
- **Network errors:** React Query retry with exponential backoff
- **403 on non-whitelisted email:** "Access denied — this tool is for Subwave team members only"

## Testing Strategy

- **Backend:** Go table-driven tests for handlers and store layer, test against a real Neon dev database
- **Frontend:** Vitest for utility functions, Playwright for critical flows (login, create campaign, update task)
- **No unit tests for UI components** — integration/E2E tests cover more ground for an internal tool
