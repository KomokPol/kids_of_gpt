package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

type profile struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Article  string `json:"article"`
	Term     string `json:"term"`
	Cell     string `json:"cell"`
	PhotoURL string `json:"photo_url"`
}

type initProfileRequest struct {
	UserID         string `json:"user_id"`
	Nickname       string `json:"nickname"`
	Article        string `json:"article"`
	Term           string `json:"term"`
	Cell           string `json:"cell"`
	PhotoURL       string `json:"photo_url"`
	IdempotencyKey string `json:"idempotency_key"`
}

type initProfileResponse struct {
	ProfileID string `json:"profile_id"`
}

type server struct {
	mu          sync.RWMutex
	profiles    map[string]profile
	idempotency map[string]string
}

func newServer() *server {
	return &server{
		profiles:    make(map[string]profile),
		idempotency: make(map[string]string),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/profile/init", s.handleInitProfile)
	mux.HandleFunc("/internal/profile/", s.handleGetProfile)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "profile"})
}

func (s *server) handleInitProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req initProfileRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.Nickname == "" || req.IdempotencyKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id, nickname and idempotency_key are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if profileID, ok := s.idempotency[req.IdempotencyKey]; ok {
		httpx.WriteJSON(w, http.StatusOK, initProfileResponse{ProfileID: profileID})
		return
	}

	profileID := "profile-" + req.UserID
	s.profiles[req.UserID] = profile{
		UserID:   req.UserID,
		Nickname: req.Nickname,
		Article:  req.Article,
		Term:     req.Term,
		Cell:     req.Cell,
		PhotoURL: req.PhotoURL,
	}
	s.idempotency[req.IdempotencyKey] = profileID

	httpx.WriteJSON(w, http.StatusOK, initProfileResponse{ProfileID: profileID})
}

func (s *server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID := strings.TrimPrefix(r.URL.Path, "/internal/profile/")
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user id is required")
		return
	}

	s.mu.RLock()
	p, ok := s.profiles[userID]
	s.mu.RUnlock()
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "profile not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, p)
}

func main() {
	addr := os.Getenv("PROFILE_ADDR")
	if addr == "" {
		addr = ":8102"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("profile service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("profile service failed: %v", err)
	}
}
