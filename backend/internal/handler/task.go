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
