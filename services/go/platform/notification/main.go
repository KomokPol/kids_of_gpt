package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

type notification struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Read      bool   `json:"read"`
	CreatedAt int64  `json:"created_at"`
}

type sendRequest struct {
	UserID string `json:"user_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type server struct {
	mu    sync.RWMutex
	store map[string][]notification // user_id -> notifications
}

func newServer() *server {
	return &server{
		store: make(map[string][]notification),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/notification/send", s.handleSend)
	mux.HandleFunc("/internal/notification/", s.handleList)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "notification"})
}

func (s *server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req sendRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.Title == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id and title are required")
		return
	}

	n := notification{
		ID:        fmt.Sprintf("notif-%d", time.Now().UnixNano()),
		UserID:    req.UserID,
		Title:     req.Title,
		Body:      req.Body,
		Read:      false,
		CreatedAt: time.Now().UTC().Unix(),
	}

	s.mu.Lock()
	s.store[req.UserID] = append(s.store[req.UserID], n)
	s.mu.Unlock()

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"accepted": true, "id": n.ID})
}

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/internal/notification/")
	if !strings.HasSuffix(path, "/list") {
		httpx.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	userID := strings.TrimSuffix(path, "/list")
	userID = strings.TrimSuffix(userID, "/")
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user id is required")
		return
	}

	s.mu.RLock()
	items := s.store[userID]
	s.mu.RUnlock()

	// Return newest first
	result := make([]notification, len(items))
	for i, n := range items {
		result[len(items)-1-i] = n
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"notifications": result})
}

func main() {
	addr := os.Getenv("NOTIFICATION_ADDR")
	if addr == "" {
		addr = ":8109"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("notification service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("notification service failed: %v", err)
	}
}
