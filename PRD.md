# Music Release Planner — Product Requirements Document

## 1. Overview

A SaaS web app that helps independent artists, labels, and managers plan and execute music release campaigns. Think Trello, but purpose-built for music releases — pre-loaded with industry-standard task templates so artists don't have to figure out what to do from scratch.

**Core value prop:** You shouldn't need a label to have a label-quality release plan. Create a campaign, get a battle-tested checklist, and execute.

---

## 2. Target Users

| Persona | Description |
|---------|-------------|
| Independent Artist | Self-releasing music, needs structure and guidance |
| Manager | Managing releases for one or more artists |
| Small Label | Coordinating releases across a roster |
| Publicist / PR | Running PR campaigns for artist clients |

---

## 3. Core Concepts

### Campaign
A single music release (single, EP, album). Contains multiple **task lists**. Can be duplicated to quickly spin up new releases.

### Task List
A categorized group of tasks within a campaign. Color-coded. Each campaign comes pre-loaded with 5 default task lists.

### Task
An individual checklist item within a task list. Has a name, status, due date, and description. Can have subtasks.

---

## 4. Default Campaign Template

When a user creates a new campaign, it is pre-populated with the following 5 task lists and their default tasks. **This is the core content/IP of the product.** Each task list contains one or more collapsible **groups**.

---

### 4.1 Campaign Assets (Blue)

#### Group: "Song Assets"

| # | Task |
|---|------|
| 1 | Make New Asset Folder |
| 2 | Release Date |
| 3 | Final Master |
| 4 | Final Artwork |
| 5 | Instrumental |
| 6 | Stems |
| 7 | Lyrics |
| 8 | Press Photos |
| 9 | Music Video/ Lyric Video |
| 10 | Folder for Promo Content |

---

### 4.2 Campaign Tasklist (Green)

#### Group: "Pre-Release"

| # | Task |
|---|------|
| 1 | Campaign Setup + Timeline |
| 2 | Contracts |
| 3 | Photoshoot for release |
| 4 | Make content plan |
| 5 | Create content or work with a graphic designer |
| 6 | Create pre-save Link |
| 7 | Pre-save campaign/ contest |
| 8 | Create linktree |
| 9 | Write out hashtags |
| 10 | Update website - coming soon |
| 11 | Update all SM profiles - coming soon |
| 12 | Update streaming profiles |

---

### 4.3 Social Media Content (Purple)

#### Group: "Pre-Release Content"

| # | Task |
|---|------|
| 1 | Teaser post |
| 2 | Banner - release date - presave now |
| 3 | Single art - story sized - release date or 'pre-save now' |
| 4 | Single art edit #1 (with release date) |
| 5 | Pre-save Post |
| 6 | Song poster #1 - storytelling (inspiration behind the track) |
| 7 | Song teaser #1 |
| 8 | "Midnight" or "Tomorrow" |

#### Group: "Post-Release Content"

| # | Task |
|---|------|
| 1 | Banner- out now |
| 2 | Song Clip - out now |
| 3 | Single art edit #2 ("out now") |
| 4 | Single art - out now - story sized |
| 5 | Song poster #2 |
| 6 | Studio content #1 - storytelling (recording process) |
| 7 | Single art edit #3- storytelling (art/ visual aspects) |
| 8 | Credits page |
| 9 | Lyric poster/ Lyric Reel |

#### Group: "Extra Content"

| # | Task |
|---|------|
| 1 | Pre-save Post #2 |
| 2 | Song clip/teaser #3 |
| 3 | Song poster #3 |
| 4 | Studio content #2 |
| 5 | Tiktok/ Reels |
| 6 | Clips from Music video/ lyric video |
| 7 | Countdowns |
| 8 | Acoustic version/ demo version/ live version |

---

### 4.4 PR (Yellow)

#### Group: "Media Outreach"

| # | Task |
|---|------|
| 1 | Research new outlets & develop CRM |
| 2 | Create press release |
| 3 | Create e-mail outreach pitch |
| 4 | Traditional Media outreach |
| 5 | New Media outreach |
| 6 | Submit to third party platforms |
| 7 | Press follow ups |

#### Group: "Playlisting"

| # | Task |
|---|------|
| 1 | Research playlists |
| 2 | Write playlist pitch |
| 3 | Pitch to independent curators |
| 4 | Submit to third party platforms |

#### Group: "Community Engagement"

| # | Task |
|---|------|
| 1 | Send DMs on IG/ Twitter/FB with link |
| 2 | Post on discord chats/ reddit threads |
| 3 | Send mailing list e-mail - presave |
| 4 | Send mailing list e-mail - out now |

---

### 4.5 Distribution (Red)

#### Group: "To Do"

| # | Task | Subtasks |
|---|------|----------|
| 1 | Obtain license for cover song | — |
| 2 | Upload to distributor | Fill out credits, Upload lyrics |
| 3 | Pitch song to Spotify editors | — |
| 4 | Upload Spotify Canvas | — |
| 5 | Register song with PROs | — |
| 6 | For physical releases | — |

---

## 5. Features

### 5.1 Campaign Management
- **Create campaign** — Name it (e.g., song/EP title), auto-populated with default template tasks
- **Duplicate campaign** — Copy an existing campaign as a starting point for a new release
- **Delete / archive campaign**
- **Campaign sidebar** — List of all campaigns with quick switching

### 5.2 Task Lists
- Each campaign has multiple task lists (default 5, user can add more)
- Color-coded (user-selectable color per list)
- Collapsible sections / groups within a task list (e.g., "Song Assets" under Campaign Assets)
- Reorderable via drag-and-drop
- "Add New Task List" to create custom lists

### 5.3 Tasks
- **Status** — `To Do` | `In Progress` | `Done` (dropdown)
- **Due date** — Date picker
- **Description** — Rich text field with links to resources/guides
- **Subtasks** — Nested checklist items within a task
- **Add Task** — Inline task creation at bottom of each list
- **Task actions menu** (three-dot "...") — Edit, delete, move, duplicate
- **Hide "Done" Tasks** — Toggle to declutter the view

### 5.4 Views
- **List view** (default) — Vertical checklist layout grouped by task list
- **Calendar view** — Tasks plotted on a calendar by due date

### 5.5 Collaboration
- Invite team members to a campaign (manager, publicist, designer, etc.)
- Role-based access (owner, editor, viewer)
- Activity feed or comments per task (stretch goal)

### 5.6 Resources
- Library of guides, videos, and links to help artists execute each phase
- Contextual resource links within task descriptions

---

## 6. Information Architecture

```
App
├── Auth (Sign up / Log in)
├── My Campaigns (dashboard)
│   ├── Campaign 1
│   │   ├── Campaign Assets (task list)
│   │   │   ├── Song Assets (group)
│   │   │   │   ├── Task 1
│   │   │   │   ├── Task 2 (with subtasks)
│   │   │   │   └── ...
│   │   │   └── [Add Task]
│   │   ├── Campaign Tasklist
│   │   ├── Social Media Content
│   │   ├── PR
│   │   ├── Distribution
│   │   └── [Add New Task List]
│   ├── Campaign 2
│   └── [+ New Campaign]
└── Resources
```

---

## 7. Tech Stack (Suggested)

| Layer | Technology |
|-------|------------|
| Frontend | Next.js (App Router) + React |
| Styling | Tailwind CSS |
| Auth | NextAuth.js or Clerk |
| Database | PostgreSQL (via Supabase or Neon) |
| ORM | Prisma or Drizzle |
| Hosting | Vercel |
| Payments | Stripe (subscription billing) |

---

## 8. Data Model (Simplified)

```
User
  id, email, name, password_hash, created_at

Campaign
  id, user_id, name, created_at, archived

TaskList
  id, campaign_id, name, color, position

TaskGroup
  id, task_list_id, name, position

Task
  id, task_group_id, name, description, status, due_date, position

Subtask
  id, task_id, name, is_complete, position

Collaborator
  id, campaign_id, user_id, role
```

---

## 9. Monetization

| Plan | Price | Features |
|------|-------|----------|
| Free | $0 | 1 active campaign, default templates, no collaboration |
| Pro | ~$5/month | Unlimited campaigns, collaboration, campaign duplication, calendar view |

---

## 10. MVP Scope

**Phase 1 — Core (MVP)**
- Auth (sign up, log in, log out)
- Create / delete campaigns with pre-loaded template
- View campaign with task lists, groups, and tasks
- Update task status (To Do / In Progress / Done)
- Set due dates on tasks
- Add / edit / delete tasks
- Hide completed tasks toggle
- List view

**Phase 2 — Polish**
- Campaign duplication
- Calendar view
- Custom task lists and groups
- Drag-and-drop reordering
- Task descriptions with rich text
- Subtasks

**Phase 3 — Growth**
- Collaboration (invite team members)
- Resources / guides library
- Stripe billing integration
- Mobile-responsive design

---

## 11. Success Metrics

- Number of campaigns created
- Task completion rate per campaign
- User retention (monthly active users)
- Free-to-paid conversion rate
- Campaign duplication rate (indicates repeat value)
