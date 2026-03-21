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
