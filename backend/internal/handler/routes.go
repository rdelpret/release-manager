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

	// Only register dev-login in development (#8)
	if os.Getenv("ENV") == "development" {
		r.Get("/auth/dev-login", auth.HandleDevLogin)
	}

	// Protected API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Use(bodySizeLimit(1 << 20)) // 1 MB max request body (#6)

		r.Get("/me", auth.HandleMe)

		// Campaigns
		r.Get("/campaigns", s.handleListCampaigns)
		r.Post("/campaigns", s.handleCreateCampaign)
		r.Get("/campaigns/{id}", s.handleGetCampaign)
		r.Post("/campaigns/{id}/duplicate", s.handleDuplicateCampaign)
		r.Patch("/campaigns/{id}/archive", s.handleArchiveCampaign)
		r.Patch("/campaigns/{id}/release-date", s.handleSetReleaseDate)
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

// bodySizeLimit restricts request body size (#6)
func bodySizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigin := os.Getenv("FRONTEND_URL")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Only allow the configured origin — no hardcoded localhost fallback (#4)
		if origin == allowedOrigin {
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
