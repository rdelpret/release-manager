# Music Release Planner

Internal release campaign management tool for the Subwave team (subwave.music).

## Project Structure

```
/backend    — Go API (chi router, pgx, Neon PostgreSQL)
/frontend   — Next.js (App Router, TailwindCSS, shadcn/ui)
Makefile    — Build orchestration (make dev, make build, make deploy)
```

## Tech Stack

- **Backend:** Go 1.24+, chi/v5 router, pgx (raw SQL, no ORM), Google OAuth
- **Frontend:** Next.js 16+, React 19, TailwindCSS 4, shadcn/ui, TypeScript
- **Database:** Neon PostgreSQL, migrations via goose or raw psql
- **Deployment:** Cloudflare Workers (TS) + Cloudflare Containers (Go Docker)
- **Key libraries:** @dnd-kit (drag-and-drop), Tiptap (rich text), React Query (data fetching)

## Design Spec

Full spec at `docs/superpowers/specs/2026-03-18-music-release-planner-design.md`

## Conventions

- Follow existing patterns from Robbie's other projects (setlist-planner, neon-todo)
- Go: table-driven tests, raw SQL with pgx, chi router, standard project layout
- Frontend: App Router, components in /src/components, hooks in /src/hooks
- Styling: Subwave brand — dark theme (#0B0D10), acid green (#B7FF4A), violet (#A78BFA)
- Fonts: Space Grotesk (headings), Inter (body)
- Use shadcn/ui components restyled to match the Subwave dark theme
- Session-based auth with cookies, not JWT
- Optimistic updates on the frontend with React Query

## Commands

```bash
make dev        # Run both backend and frontend in development
make build      # Build both for production
make deploy     # Deploy to Cloudflare
make migrate    # Run database migrations
```

## Auth

- Google OAuth only, whitelisted email addresses
- Non-whitelisted emails get 403 at callback
- Session cookie checked by auth middleware on all /api/* routes

## Workflow

- **Never push directly to main.** Always create a feature branch and PR.
- Branch protection requires all CI checks to pass before merge.
- At the start of each session or when a task is done: `git checkout main && git pull` then create a new branch.
- Pre-commit hook runs: go vet, go build, go test, npm lint, production build.
- E2E tests run in CI against a Docker Postgres (not prod DB).
