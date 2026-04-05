package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

type balance struct {
	Balance     int64 `json:"balance"`
	DebtBalance int64 `json:"debt_balance"`
}

type initWalletRequest struct {
	UserID         string `json:"user_id"`
	InitialBalance int64  `json:"initial_balance"`
	IdempotencyKey string `json:"idempotency_key"`
}

type mutateRequest struct {
	UserID         string `json:"user_id"`
	Amount         int64  `json:"amount"`
	Source         string `json:"source"`
	IdempotencyKey string `json:"idempotency_key"`
}

type mutateResponse struct {
	Balance int64 `json:"balance"`
}

type ledgerEntry struct {
	UserID    string `json:"user_id"`
	Amount    int64  `json:"amount"`
	EntryType string `json:"entry_type"`
	Source    string `json:"source"`
	AtUnix    int64  `json:"at_unix"`
}

type server struct {
	mu             sync.RWMutex
	balances       map[string]int64
	ledger         map[string][]ledgerEntry
	idempotency    map[string]int64
	progressionURL string
	httpClient     *http.Client
}

func newServer() *server {
	return &server{
		balances:       make(map[string]int64),
		ledger:         make(map[string][]ledgerEntry),
		idempotency:    make(map[string]int64),
		progressionURL: envOrDefault("PROGRESSION_BASE_URL", "http://localhost:8104"),
		httpClient:     &http.Client{Timeout: 3 * time.Second},
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
	mux.HandleFunc("/internal/wallet/init", s.handleInitWallet)
	mux.HandleFunc("/internal/wallet/credit", s.handleCredit)
	mux.HandleFunc("/internal/wallet/debit", s.handleDebit)
	mux.HandleFunc("/internal/wallet/debtors", s.handleGetDebtors)
	mux.HandleFunc("/internal/wallet/", s.handleGetBalance)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "wallet"})
}

func (s *server) handleInitWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req initWalletRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.IdempotencyKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id and idempotency_key are required")
		return
	}

	s.mu.Lock()

	if b, ok := s.idempotency[req.IdempotencyKey]; ok {
		s.mu.Unlock()
		s.syncProgression(req.UserID, b)
		httpx.WriteJSON(w, http.StatusOK, mutateResponse{Balance: b})
		return
	}

	s.balances[req.UserID] = req.InitialBalance
	s.idempotency[req.IdempotencyKey] = req.InitialBalance
	s.mu.Unlock()

	s.syncProgression(req.UserID, req.InitialBalance)
	httpx.WriteJSON(w, http.StatusOK, mutateResponse{Balance: req.InitialBalance})
}

func (s *server) handleCredit(w http.ResponseWriter, r *http.Request) {
	s.handleMutate(w, r, "credit")
}

func (s *server) handleDebit(w http.ResponseWriter, r *http.Request) {
	s.handleMutate(w, r, "debit")
}

func (s *server) handleMutate(w http.ResponseWriter, r *http.Request, op string) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req mutateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.IdempotencyKey == "" || req.Amount <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "user_id, amount>0 and idempotency_key are required")
		return
	}

	s.mu.Lock()

	idemKey := op + ":" + req.IdempotencyKey
	if b, ok := s.idempotency[idemKey]; ok {
		s.mu.Unlock()
		s.syncProgression(req.UserID, b)
		httpx.WriteJSON(w, http.StatusOK, mutateResponse{Balance: b})
		return
	}

	cur := s.balances[req.UserID]
	if op == "credit" {
		cur += req.Amount
	} else {
		cur -= req.Amount
	}
	s.balances[req.UserID] = cur
	s.idempotency[idemKey] = cur
	s.ledger[req.UserID] = append(s.ledger[req.UserID], ledgerEntry{
		UserID:    req.UserID,
		Amount:    req.Amount,
		EntryType: op,
		Source:    req.Source,
		AtUnix:    time.Now().UTC().Unix(),
	})
	s.mu.Unlock()

	s.syncProgression(req.UserID, cur)

	httpx.WriteJSON(w, http.StatusOK, mutateResponse{Balance: cur})
}

func (s *server) syncProgression(userID string, balance int64) {
	debt := int64(0)
	if balance < 0 {
		debt = balance
	}
	payload := map[string]any{
		"user_id":        userID,
		"wallet_balance": balance,
		"debt_balance":   debt,
	}
	_, status, err := httpx.PostJSON(s.httpClient, s.progressionURL+"/internal/progression/sync-wallet", payload)
	if err != nil || status >= http.StatusBadRequest {
		log.Printf("wallet: failed to sync progression for %s: status=%d err=%v", userID, status, err)
	}
}

func (s *server) handleGetDebtors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	type debtor struct {
		UserID  string `json:"user_id"`
		Balance int64  `json:"balance"`
	}
	var debtors []debtor
	for uid, bal := range s.balances {
		if bal < 0 {
			debtors = append(debtors, debtor{UserID: uid, Balance: bal})
		}
	}
	if debtors == nil {
		debtors = []debtor{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"debtors": debtors})
}

func (s *server) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/internal/wallet/")
	if !strings.HasSuffix(path, "/balance") {
		httpx.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	userID := strings.TrimSuffix(path, "/balance")
	userID = strings.TrimSuffix(userID, "/")
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user id is required")
		return
	}

	s.mu.RLock()
	b := s.balances[userID]
	s.mu.RUnlock()

	debt := int64(0)
	if b < 0 {
		debt = b
	}
	httpx.WriteJSON(w, http.StatusOK, balance{Balance: b, DebtBalance: debt})
}

func main() {
	addr := envOrDefault("WALLET_ADDR", ":8103")
	s := newServer()

	srv := &http.Server{
		Addr:    addr,
		Handler: s.routes(),
	}

	log.Printf("wallet service listening on %s", addr)
	log.Printf("wallet progression sync target: %s", s.progressionURL)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("wallet service failed: %v", err)
	}
}
