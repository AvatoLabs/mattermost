package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-read-index-service/internal/index"
)

type Server struct {
	indexService *index.Service
	router       *mux.Router
	httpServer   *http.Server
	port         string
}

type ReadersResponse struct {
	Count     int      `json:"count"`
	Readers   []string `json:"readers"`
	Truncated bool     `json:"truncated"`
}

type ReadCountsRequest struct {
	ChannelID string  `json:"channel_id"`
	Seqs      []int64 `json:"seqs"`
}

func NewServer(indexService *index.Service, port string) *Server {
	s := &Server{
		indexService: indexService,
		port:         port,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()

	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Get readers for a specific message
	s.router.HandleFunc("/channels/{channel_id}/posts/{seq}/readers", s.handleGetReaders).Methods("GET")

	// Batch get read counts
	s.router.HandleFunc("/read-counts", s.handleGetReadCounts).Methods("POST")

	// Service stats
	s.router.HandleFunc("/stats", s.handleStats).Methods("GET")
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleGetReaders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID := vars["channel_id"]
	seqStr := vars["seq"]

	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid sequence number", http.StatusBadRequest)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	readers, count, err := s.indexService.GetReadersForSeq(channelID, seq, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ReadersResponse{
		Count:     count,
		Readers:   readers,
		Truncated: count > len(readers),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetReadCounts(w http.ResponseWriter, r *http.Request) {
	var req ReadCountsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	counts := s.indexService.GetReadCounts(req.ChannelID, req.Seqs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.indexService.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
