package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

type debtor struct {
	UserID  string `json:"user_id"`
	Balance int64  `json:"balance"`
}

type tickStatus struct {
	LastTickUnix int64 `json:"last_tick_unix"`
	DebtorsFound int   `json:"debtors_found"`
	Processed    int   `json:"processed"`
}

type server struct {
	mu              sync.RWMutex
	lastTick        tickStatus
	walletURL       string
	progressionURL  string
	notificationURL string
	httpClient      *http.Client
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
		walletURL:       envOrDefault("WALLET_BASE_URL", "http://localhost:8103"),
		progressionURL:  envOrDefault("PROGRESSION_BASE_URL", "http://localhost:8104"),
		notificationURL: envOrDefault("NOTIFICATION_BASE_URL", "http://localhost:8109"),
		httpClient:      &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/scheduler/tick", s.handleTick)
	mux.HandleFunc("/internal/scheduler/status", s.handleStatus)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "scheduler"})
}

func (s *server) handleTick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	result := s.accrueInterest()
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.mu.RLock()
	status := s.lastTick
	s.mu.RUnlock()
	httpx.WriteJSON(w, http.StatusOK, status)
}

func (s *server) accrueInterest() tickStatus {
	tickTS := time.Now().UTC().Unix()

	// Get all debtors
	var debtorsResp struct {
		Debtors []debtor `json:"debtors"`
	}
	_, err := httpx.GetJSON(s.httpClient, s.walletURL+"/internal/wallet/debtors", &debtorsResp)
	if err != nil {
		log.Printf("scheduler: failed to get debtors: %v", err)
		return tickStatus{LastTickUnix: tickTS}
	}

	processed := 0
	for _, d := range debtorsResp.Debtors {
		interest := int64(float64(-d.Balance) * 0.05)
		if interest <= 0 {
			interest = 1 // minimum 1 unit interest
		}

		// Accrue interest via wallet debit
		idemKey := fmt.Sprintf("interest:%s:%d", d.UserID, tickTS)
		payload := map[string]any{
			"user_id":         d.UserID,
			"amount":          interest,
			"source":          "scheduler:debt_interest",
			"idempotency_key": idemKey,
		}
		out, status, err := httpx.PostJSON(s.httpClient, s.walletURL+"/internal/wallet/debit", payload)
		if err != nil || status >= 400 {
			log.Printf("scheduler: failed to accrue interest for %s: %v (status=%d)", d.UserID, err, status)
			continue
		}

		var walletResp struct {
			Balance int64 `json:"balance"`
		}
		json.Unmarshal(out, &walletResp)

		// Recalculate progression if debt threshold
		if walletResp.Balance <= -5000 {
			recalc := map[string]any{
				"user_id":        d.UserID,
				"wallet_balance": walletResp.Balance,
				"debt_balance":   walletResp.Balance,
			}
			httpx.PostJSON(s.httpClient, s.progressionURL+"/internal/progression/recalculate", recalc)

			// Notify about ban
			notif := map[string]any{
				"user_id": d.UserID,
				"title":   "Этапирование!",
				"body":    fmt.Sprintf("Долг превысил 5000 (текущий: %d). Бан на 12 часов.", -walletResp.Balance),
			}
			httpx.PostJSON(s.httpClient, s.notificationURL+"/internal/notification/send", notif)
		}

		processed++
		log.Printf("scheduler: accrued %d interest for %s, new balance: %d", interest, d.UserID, walletResp.Balance)
	}

	result := tickStatus{
		LastTickUnix: tickTS,
		DebtorsFound: len(debtorsResp.Debtors),
		Processed:    processed,
	}

	s.mu.Lock()
	s.lastTick = result
	s.mu.Unlock()

	return result
}

func (s *server) runInterestTicker() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		log.Println("scheduler: running interest accrual tick")
		s.accrueInterest()
	}
}

func main() {
	addr := envOrDefault("SCHEDULER_ADDR", ":8111")
	s := newServer()

	// Start background ticker
	go s.runInterestTicker()

	srv := &http.Server{
		Addr:    addr,
		Handler: s.routes(),
	}

	log.Printf("scheduler service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("scheduler service failed: %v", err)
	}
}
