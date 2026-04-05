package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

type film struct {
	FilmID string `json:"film_id"`
	Title  string `json:"title"`
	Tier   string `json:"tier"` // "pervohod", "patsan", "smotryashiy"
	Price  int64  `json:"price"`
}

type filmView struct {
	FilmID    string `json:"film_id"`
	Title     string `json:"title"`
	Tier      string `json:"tier"`
	Price     int64  `json:"price"`
	Available bool   `json:"available"`
}

type watchRequest struct {
	UserID         string `json:"user_id"`
	FilmID         string `json:"film_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

type watchResponse struct {
	Watched bool  `json:"watched"`
	Balance int64 `json:"balance"`
}

type progressionSnapshot struct {
	FilmSubscriptionTier string `json:"film_subscription_tier"`
	BannedUntilUnix      int64  `json:"banned_until_unix"`
}

type server struct {
	mu             sync.RWMutex
	films          map[string]film
	filmOrder      []string            // ordered film IDs
	watchHistory   map[string][]string // user_id -> film_ids
	idempotency    map[string]watchResponse
	walletURL      string
	progressionURL string
	httpClient     *http.Client
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func newServer() *server {
	s := &server{
		films:          make(map[string]film),
		watchHistory:   make(map[string][]string),
		idempotency:    make(map[string]watchResponse),
		walletURL:      envOrDefault("WALLET_BASE_URL", "http://localhost:8103"),
		progressionURL: envOrDefault("PROGRESSION_BASE_URL", "http://localhost:8104"),
		httpClient:     &http.Client{Timeout: 3 * time.Second},
	}
	seed := []film{
		{FilmID: "film-1", Title: "Побег из Шоушенка", Tier: "pervohod", Price: 0},
		{FilmID: "film-2", Title: "Зелёная миля", Tier: "pervohod", Price: 0},
		{FilmID: "film-3", Title: "Бригада", Tier: "patsan", Price: 100},
		{FilmID: "film-4", Title: "Крёстный отец", Tier: "patsan", Price: 150},
		{FilmID: "film-5", Title: "Джентльмены", Tier: "smotryashiy", Price: 300},
	}
	for _, f := range seed {
		s.films[f.FilmID] = f
		s.filmOrder = append(s.filmOrder, f.FilmID)
	}
	return s
}

func tierRank(t string) int {
	switch t {
	case "pervohod":
		return 1
	case "patsan":
		return 2
	case "smotryashiy":
		return 3
	default:
		return 0
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/kinoshka/watch", s.handleWatch)
	mux.HandleFunc("/internal/kinoshka/", s.handleFilms)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "kinoshka"})
}

func (s *server) handleFilms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/internal/kinoshka/")
	if !strings.HasSuffix(path, "/films") {
		httpx.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	userID := strings.TrimSuffix(path, "/films")
	userID = strings.TrimSuffix(userID, "/")
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user id is required")
		return
	}

	// Get user tier
	var snap progressionSnapshot
	snapStatus, err := httpx.GetJSON(s.httpClient,
		fmt.Sprintf("%s/internal/progression/%s/snapshot", s.progressionURL, userID), &snap)
	if err != nil || snapStatus >= 400 {
		httpx.WriteError(w, http.StatusBadGateway, "failed to get progression snapshot")
		return
	}
	if snap.BannedUntilUnix > 0 && snap.BannedUntilUnix > time.Now().UTC().Unix() {
		httpx.WriteError(w, http.StatusForbidden, "user is banned")
		return
	}

	userRank := tierRank(snap.FilmSubscriptionTier)

	s.mu.RLock()
	var result []filmView
	for _, fid := range s.filmOrder {
		f := s.films[fid]
		filmRank := tierRank(f.Tier)
		result = append(result, filmView{
			FilmID:    f.FilmID,
			Title:     f.Title,
			Tier:      f.Tier,
			Price:     f.Price,
			Available: filmRank <= userRank,
		})
	}
	s.mu.RUnlock()

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"films": result, "user_tier": snap.FilmSubscriptionTier})
}

func (s *server) handleWatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req watchRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.FilmID == "" || req.IdempotencyKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id, film_id and idempotency_key are required")
		return
	}

	s.mu.RLock()
	if resp, ok := s.idempotency[req.IdempotencyKey]; ok {
		s.mu.RUnlock()
		httpx.WriteJSON(w, http.StatusOK, resp)
		return
	}
	f, ok := s.films[req.FilmID]
	s.mu.RUnlock()
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "film not found")
		return
	}

	// Check tier access
	var snap progressionSnapshot
	snapStatus, err := httpx.GetJSON(s.httpClient,
		fmt.Sprintf("%s/internal/progression/%s/snapshot", s.progressionURL, req.UserID), &snap)
	if err != nil || snapStatus >= 400 {
		httpx.WriteError(w, http.StatusBadGateway, "failed to get progression snapshot")
		return
	}
	if tierRank(f.Tier) > tierRank(snap.FilmSubscriptionTier) {
		httpx.WriteError(w, http.StatusForbidden, "film tier not available for your subscription")
		return
	}

	var balance int64
	// If film costs money, debit
	if f.Price > 0 {
		payload := map[string]any{
			"user_id":         req.UserID,
			"amount":          f.Price,
			"source":          "kinoshka:watch",
			"idempotency_key": "kinoshka:" + req.IdempotencyKey,
		}
		out, status, err := httpx.PostJSON(s.httpClient, s.walletURL+"/internal/wallet/debit", payload)
		if err != nil || status >= 400 {
			httpx.WriteError(w, http.StatusBadGateway, "failed to debit wallet")
			return
		}
		var walletResp struct {
			Balance int64 `json:"balance"`
		}
		json.Unmarshal(out, &walletResp)
		balance = walletResp.Balance

		if balance < 0 {
			go func() {
				recalc := map[string]any{
					"user_id":        req.UserID,
					"wallet_balance": balance,
					"debt_balance":   balance,
				}
				httpx.PostJSON(s.httpClient, s.progressionURL+"/internal/progression/recalculate", recalc)
			}()
		}
	}

	resp := watchResponse{Watched: true, Balance: balance}

	s.mu.Lock()
	s.idempotency[req.IdempotencyKey] = resp
	s.watchHistory[req.UserID] = append(s.watchHistory[req.UserID], req.FilmID)
	s.mu.Unlock()

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func main() {
	addr := envOrDefault("KINOSHKA_ADDR", ":8114")
	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("kinoshka service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("kinoshka service failed: %v", err)
	}
}
