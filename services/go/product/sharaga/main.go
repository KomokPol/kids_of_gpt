package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

type runRequest struct {
	UserID         string `json:"user_id"`
	ChallengeID    string `json:"challenge_id"`
	Mode           string `json:"mode"`
	IdempotencyKey string `json:"idempotency_key"`
}

type runResponse struct {
	Success       bool  `json:"success"`
	XPAwarded     int64 `json:"xp_awarded"`
	MoneyAwarded  int64 `json:"money_awarded"`
	WalletBalance int64 `json:"wallet_balance"`
	NewLevel      int32 `json:"new_level"`
}

type server struct {
	mu             sync.RWMutex
	idempotency    map[string]runResponse
	walletURL      string
	progressionURL string
	httpClient     *http.Client
}

func newServer() *server {
	return &server{
		idempotency:    make(map[string]runResponse),
		walletURL:      envOrDefault("WALLET_BASE_URL", "http://localhost:8103"),
		progressionURL: envOrDefault("PROGRESSION_BASE_URL", "http://localhost:8104"),
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/sharaga/run", s.handleRun)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "sharaga"})
}

func (s *server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req runRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.ChallengeID == "" || req.IdempotencyKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id, challenge_id and idempotency_key are required")
		return
	}

	s.mu.RLock()
	if resp, ok := s.idempotency[req.IdempotencyKey]; ok {
		s.mu.RUnlock()
		httpx.WriteJSON(w, http.StatusOK, resp)
		return
	}
	s.mu.RUnlock()

	xpAward, moneyAward := rewardForMode(req.Mode)

	grantResp, err := s.grantXP(req.UserID, xpAward, req.Mode, req.IdempotencyKey)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "failed to grant xp")
		return
	}
	walletResp, err := s.creditMoney(req.UserID, moneyAward, req.Mode, req.IdempotencyKey)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "failed to credit money")
		return
	}

	resp := runResponse{
		Success:       true,
		XPAwarded:     xpAward,
		MoneyAwarded:  moneyAward,
		WalletBalance: walletResp.Balance,
		NewLevel:      grantResp.Level,
	}

	s.mu.Lock()
	s.idempotency[req.IdempotencyKey] = resp
	s.mu.Unlock()

	httpx.WriteJSON(w, http.StatusOK, resp)
}

type progressionGrantResponse struct {
	Level int32 `json:"level"`
}

func (s *server) grantXP(userID string, xp int64, mode, idem string) (*progressionGrantResponse, error) {
	payload := map[string]any{
		"user_id":         userID,
		"xp":              xp,
		"source":          "sharaga:" + mode,
		"idempotency_key": "sharaga:xp:" + idem,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.progressionURL+"/internal/progression/grant-xp", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("progression status: %d", resp.StatusCode)
	}

	var out progressionGrantResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

type walletCreditResponse struct {
	Balance int64 `json:"balance"`
}

func (s *server) creditMoney(userID string, amount int64, mode, idem string) (*walletCreditResponse, error) {
	payload := map[string]any{
		"user_id":         userID,
		"amount":          amount,
		"source":          "sharaga:" + mode,
		"idempotency_key": "sharaga:money:" + idem,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.walletURL+"/internal/wallet/credit", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("wallet status: %d", resp.StatusCode)
	}

	var out walletCreditResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func rewardForMode(mode string) (xp int64, money int64) {
	switch mode {
	case "interrogation":
		return 80, 40
	case "scenario":
		return 50, 25
	default:
		return 20, 10
	}
}

func main() {
	addr := envOrDefault("SHARAGA_ADDR", ":8105")
	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("sharaga service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("sharaga service failed: %v", err)
	}
}
