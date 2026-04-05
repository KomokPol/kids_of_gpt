package main

import (
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

type updateRequest struct {
	UserID string `json:"user_id"`
	Scope  string `json:"scope"`
	Score  int64  `json:"score"`
}

type entry struct {
	UserID string `json:"user_id"`
	Score  int64  `json:"score"`
}

type server struct {
	mu     sync.RWMutex
	boards map[string]map[string]int64 // scope -> user_id -> score
}

func newServer() *server {
	return &server{
		boards: make(map[string]map[string]int64),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/leaderboard/update", s.handleUpdate)
	mux.HandleFunc("/internal/leaderboard/top", s.handleTop)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "leaderboard"})
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req updateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.Scope == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id and scope are required")
		return
	}

	s.mu.Lock()
	if s.boards[req.Scope] == nil {
		s.boards[req.Scope] = make(map[string]int64)
	}
	s.boards[req.Scope][req.UserID] = req.Score
	s.mu.Unlock()

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *server) handleTop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "xp"
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	s.mu.RLock()
	board := s.boards[scope]
	entries := make([]entry, 0, len(board))
	for uid, score := range board {
		entries = append(entries, entry{UserID: uid, Score: score})
	}
	s.mu.RUnlock()

	if scope == "debt" {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Score < entries[j].Score // most negative first
		})
	} else {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Score > entries[j].Score // highest first
		})
	}

	if len(entries) > limit {
		entries = entries[:limit]
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"scope": scope, "entries": entries})
}

func main() {
	addr := os.Getenv("LEADERBOARD_ADDR")
	if addr == "" {
		addr = ":8108"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("leaderboard service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("leaderboard service failed: %v", err)
	}
}
