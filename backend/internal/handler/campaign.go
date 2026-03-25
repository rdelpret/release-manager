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
		Name        string  `json:"name"`
		ReleaseDate *string `json:"release_date,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	campaign, err := s.store.CreateCampaign(r.Context(), userID, req.Name, req.ReleaseDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create campaign")
		return
	}
	writeJSON(w, http.StatusCreated, campaign)
}

func (s *Server) handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	userID := auth.GetUserID(r)

	ok, err := s.store.IsCampaignMember(r.Context(), campaignID, userID)
	if err != nil || !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

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

	ok, err := s.store.IsCampaignMember(r.Context(), campaignID, userID)
	if err != nil || !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	campaign, err := s.store.DuplicateCampaign(r.Context(), campaignID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to duplicate campaign")
		return
	}
	writeJSON(w, http.StatusCreated, campaign)
}

func (s *Server) handleArchiveCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	userID := auth.GetUserID(r)

	ok, err := s.store.IsCampaignMember(r.Context(), campaignID, userID)
	if err != nil || !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

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

func (s *Server) handleSetReleaseDate(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	userID := auth.GetUserID(r)

	ok, err := s.store.IsCampaignMember(r.Context(), campaignID, userID)
	if err != nil || !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var req struct {
		ReleaseDate   string `json:"release_date"`
		ScheduleWeeks int    `json:"schedule_weeks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.ReleaseDate == "" {
		writeError(w, http.StatusBadRequest, "release_date is required")
		return
	}
	if req.ScheduleWeeks != 4 && req.ScheduleWeeks != 8 {
		writeError(w, http.StatusBadRequest, "schedule_weeks must be 4 or 8")
		return
	}

	if err := s.store.SetReleaseDate(r.Context(), campaignID, req.ReleaseDate, req.ScheduleWeeks); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to set release date")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeleteCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	userID := auth.GetUserID(r)

	ok, err := s.store.IsCampaignMember(r.Context(), campaignID, userID)
	if err != nil || !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err := s.store.DeleteCampaign(r.Context(), campaignID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete campaign")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
