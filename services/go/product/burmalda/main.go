package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

type spinRequest struct {
	UserID         string `json:"user_id"`
	Stake          int64  `json:"stake"`
	IdempotencyKey string `json:"idempotency_key"`
}

type spinResponse struct {
	SpinID  string `json:"spin_id"`
	Delta   int64  `json:"delta"`
	Outcome string `json:"outcome"`
	Balance int64  `json:"balance"`
}

type spinRecord struct {
	SpinID  string `json:"spin_id"`
	Stake   int64  `json:"stake"`
	Delta   int64  `json:"delta"`
	Outcome string `json:"outcome"`
	AtUnix  int64  `json:"at_unix"`
}

type progressionSnapshot struct {
	BannedUntilUnix    int64 `json:"banned_until_unix"`
	DailySpinAvailable bool  `json:"daily_spin_available"`
	DebtBalance        int64 `json:"debt_balance"`
}

type server struct {
	mu              sync.RWMutex
	spins           map[string][]spinRecord // user_id -> history
	idempotency     map[string]spinResponse
	dailySpinUsed   map[string]string // user_id -> date "2006-01-02"
	walletURL       string
	progressionURL  string
	leaderboardURL  string
	notificationURL string
	httpClient      *http.Client
	rng             *rand.Rand
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func newServer() *server {
	return &server{
		spins:           make(map[string][]spinRecord),
		idempotency:     make(map[string]spinResponse),
		dailySpinUsed:   make(map[string]string),
		walletURL:       envOrDefault("WALLET_BASE_URL", "http://localhost:8103"),
		progressionURL:  envOrDefault("PROGRESSION_BASE_URL", "http://localhost:8104"),
		leaderboardURL:  envOrDefault("LEADERBOARD_BASE_URL", "http://localhost:8108"),
		notificationURL: envOrDefault("NOTIFICATION_BASE_URL", "http://localhost:8109"),
		httpClient:      &http.Client{Timeout: 3 * time.Second},
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/burmalda/spin", s.handleSpin)
	mux.HandleFunc("/internal/burmalda/", s.handleState)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "burmalda"})
}

func (s *server) handleSpin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req spinRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.IdempotencyKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id and idempotency_key are required")
		return
	}

	s.mu.RLock()
	if resp, ok := s.idempotency[req.IdempotencyKey]; ok {
		s.mu.RUnlock()
		httpx.WriteJSON(w, http.StatusOK, resp)
		return
	}
	s.mu.RUnlock()

	// Check ban via progression snapshot
	var snap progressionSnapshot
	snapStatus, err := httpx.GetJSON(s.httpClient,
		fmt.Sprintf("%s/internal/progression/%s/snapshot", s.progressionURL, req.UserID), &snap)
	if err != nil || snapStatus >= 400 {
		httpx.WriteError(w, http.StatusBadGateway, "failed to get progression snapshot")
		return
	}
	if snap.BannedUntilUnix > 0 && snap.BannedUntilUnix > time.Now().UTC().Unix() {
		httpx.WriteError(w, http.StatusForbidden, "user is banned")
		return
	}

	// Free daily spin
	stake := req.Stake
	if stake == 0 {
		today := time.Now().UTC().Format("2006-01-02")
		s.mu.RLock()
		usedDate := s.dailySpinUsed[req.UserID]
		s.mu.RUnlock()
		if usedDate == today {
			httpx.WriteError(w, http.StatusConflict, "daily spin already used today")
			return
		}
		stake = 10 // free spin worth 10
		s.mu.Lock()
		s.dailySpinUsed[req.UserID] = today
		s.mu.Unlock()
	} else if stake < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "stake must be >= 0")
		return
	}

	// RNG outcome
	outcome, delta := s.rollOutcome(stake)

	// Execute wallet mutation
	var walletBalance int64
	if delta > 0 {
		walletBalance, err = s.walletMutate("credit", req.UserID, delta, "burmalda:spin", req.IdempotencyKey)
	} else if delta < 0 {
		walletBalance, err = s.walletMutate("debit", req.UserID, -delta, "burmalda:spin", req.IdempotencyKey)
	}
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "failed to update wallet")
		return
	}

	// Check debt threshold
	if walletBalance < 0 {
		go s.recalculateProgression(req.UserID, walletBalance)
		go s.updateLeaderboard(req.UserID, walletBalance)
	}

	spinID := fmt.Sprintf("spin-%d", time.Now().UnixNano())
	resp := spinResponse{
		SpinID:  spinID,
		Delta:   delta,
		Outcome: outcome,
		Balance: walletBalance,
	}

	s.mu.Lock()
	s.idempotency[req.IdempotencyKey] = resp
	s.spins[req.UserID] = append(s.spins[req.UserID], spinRecord{
		SpinID:  spinID,
		Stake:   stake,
		Delta:   delta,
		Outcome: outcome,
		AtUnix:  time.Now().UTC().Unix(),
	})
	s.mu.Unlock()

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (s *server) rollOutcome(stake int64) (string, int64) {
	roll := s.rng.Intn(100)
	switch {
	case roll < 55: // 0-54: lose (55%)
		return "lose", -stake
	case roll < 70: // 55-69: near_miss (15%)
		return "near_miss", -stake
	case roll < 88: // 70-87: win_2x (18%)
		return "win_2x", stake
	case roll < 96: // 88-95: win_5x (8%)
		return "win_5x", 4 * stake
	default: // 96-99: jackpot (4%)
		return "jackpot", 9 * stake
	}
}

func (s *server) walletMutate(op, userID string, amount int64, source, idem string) (int64, error) {
	payload := map[string]any{
		"user_id":         userID,
		"amount":          amount,
		"source":          source,
		"idempotency_key": "burmalda:" + op + ":" + idem,
	}
	out, status, err := httpx.PostJSON(s.httpClient, s.walletURL+"/internal/wallet/"+op, payload)
	if err != nil {
		return 0, err
	}
	if status >= 400 {
		return 0, fmt.Errorf("wallet %s status: %d", op, status)
	}
	var resp struct {
		Balance int64 `json:"balance"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return 0, err
	}
	return resp.Balance, nil
}

func (s *server) recalculateProgression(userID string, balance int64) {
	payload := map[string]any{
		"user_id":        userID,
		"wallet_balance": balance,
		"debt_balance":   balance,
	}
	httpx.PostJSON(s.httpClient, s.progressionURL+"/internal/progression/recalculate", payload)
}

func (s *server) updateLeaderboard(userID string, debtBalance int64) {
	payload := map[string]any{
		"user_id": userID,
		"scope":   "debt",
		"score":   debtBalance,
	}
	httpx.PostJSON(s.httpClient, s.leaderboardURL+"/internal/leaderboard/update", payload)
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/internal/burmalda/")
	if !strings.HasSuffix(path, "/state") {
		httpx.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	userID := strings.TrimSuffix(path, "/state")
	userID = strings.TrimSuffix(userID, "/")
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user id is required")
		return
	}

	today := time.Now().UTC().Format("2006-01-02")
	s.mu.RLock()
	history := s.spins[userID]
	dailyAvailable := s.dailySpinUsed[userID] != today
	s.mu.RUnlock()

	if history == nil {
		history = []spinRecord{}
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"user_id":              userID,
		"spins":                history,
		"daily_spin_available": dailyAvailable,
	})
}

func main() {
	addr := envOrDefault("BURMALDA_ADDR", ":8112")
	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("burmalda service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("burmalda service failed: %v", err)
	}
}
