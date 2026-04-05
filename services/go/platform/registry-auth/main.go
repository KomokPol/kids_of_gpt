package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"strings"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
	"github.com/KomokPol/kids_of_gpt/services/go/common/token"
)

type registerRequest struct {
	Nickname       string `json:"nickname"`
	Article        string `json:"article"`
	Term           string `json:"term"`
	Cell           string `json:"cell"`
	PIN            string `json:"pin"`
	PhotoURL       string `json:"photoUrl"`
	AcceptedRules  bool   `json:"acceptedRules"`
	IdempotencyKey string `json:"idempotencyKey"`
	ExperienceMode string `json:"experienceMode"`
}

type registerResponse struct {
	UserID       string `json:"userId"`
	InmateNumber string `json:"inmateNumber"`
	BarcodeURL   string `json:"barcodeUrl"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type loginRequest struct {
	InmateNumber string `json:"inmateNumber"`
	PIN          string `json:"pin"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type user struct {
	UserID       string
	InmateNumber string
	PINHash      []byte
	Nickname     string
	Article      string
	Term         string
	Cell         string
	PhotoURL     string
}

type server struct {
	mu               sync.RWMutex
	usersByInmate    map[string]user
	usersByID        map[string]user
	refreshToUser    map[string]string
	registerIdem     map[string]registerResponse
	jwtSecret        []byte
	profileURL       string
	walletURL        string
	progressionURL   string
	httpClient       *http.Client
	inmateIDSequence int64
}

func newServer() *server {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-jwt-secret"
	}
	return &server{
		usersByInmate:  make(map[string]user),
		usersByID:      make(map[string]user),
		refreshToUser:  make(map[string]string),
		registerIdem:   make(map[string]registerResponse),
		jwtSecret:      []byte(secret),
		profileURL:     envOrDefault("PROFILE_BASE_URL", "http://localhost:8102"),
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
	mux.HandleFunc("/internal/register", s.handleRegister)
	mux.HandleFunc("/internal/login", s.handleLogin)
	mux.HandleFunc("/internal/refresh", s.handleRefresh)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "registry-auth"})
}

func (s *server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req registerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Nickname == "" || req.PIN == "" || req.IdempotencyKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "nickname, pin and idempotencyKey are required")
		return
	}
	if len(req.PIN) < 4 {
		httpx.WriteError(w, http.StatusBadRequest, "pin is too short")
		return
	}
	if !req.AcceptedRules {
		httpx.WriteError(w, http.StatusBadRequest, "acceptedRules must be true")
		return
	}

	s.mu.RLock()
	if resp, ok := s.registerIdem[req.IdempotencyKey]; ok {
		s.mu.RUnlock()
		httpx.WriteJSON(w, http.StatusOK, resp)
		return
	}
	s.mu.RUnlock()

	inmateNum := fmt.Sprintf("INM-%06d", atomic.AddInt64(&s.inmateIDSequence, 1))
	userID := fmt.Sprintf("u-%d", time.Now().UTC().UnixNano())
	u := user{
		UserID:       userID,
		InmateNumber: inmateNum,
		PINHash:      s.hashPIN(req.PIN),
		Nickname:     req.Nickname,
		Article:      req.Article,
		Term:         req.Term,
		Cell:         req.Cell,
		PhotoURL:     req.PhotoURL,
	}

	s.mu.Lock()
	if _, exists := s.usersByInmate[inmateNum]; exists {
		s.mu.Unlock()
		httpx.WriteError(w, http.StatusConflict, "inmate number collision")
		return
	}
	s.usersByInmate[inmateNum] = u
	s.usersByID[userID] = u
	s.mu.Unlock()

	if err := s.bootstrapProfile(u, req.IdempotencyKey); err != nil {
		s.deleteUser(u)
		httpx.WriteError(w, http.StatusBadGateway, "profile bootstrap failed")
		return
	}
	if err := s.bootstrapWallet(u, req.IdempotencyKey); err != nil {
		s.deleteUser(u)
		httpx.WriteError(w, http.StatusBadGateway, "wallet bootstrap failed")
		return
	}
	if err := s.bootstrapProgression(u, req.IdempotencyKey, req.ExperienceMode); err != nil {
		s.deleteUser(u)
		httpx.WriteError(w, http.StatusBadGateway, "progression bootstrap failed")
		return
	}

	access, _, err := token.Generate(s.jwtSecret, userID, inmateNum, 15*time.Minute)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to sign access token")
		return
	}
	refresh, _, err := token.Generate(s.jwtSecret, userID, inmateNum, 7*24*time.Hour)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to sign refresh token")
		return
	}

	resp := registerResponse{
		UserID:       userID,
		InmateNumber: inmateNum,
		BarcodeURL:   fmt.Sprintf("/assets/barcodes/%s.png", inmateNum),
		AccessToken:  access,
		RefreshToken: refresh,
	}

	s.mu.Lock()
	s.refreshToUser[refresh] = userID
	s.registerIdem[req.IdempotencyKey] = resp
	s.mu.Unlock()

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	s.mu.RLock()
	u, ok := s.usersByInmate[req.InmateNumber]
	s.mu.RUnlock()
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if !s.comparePIN(u.PINHash, req.PIN) {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	access, _, err := token.Generate(s.jwtSecret, u.UserID, u.InmateNumber, 15*time.Minute)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to sign access token")
		return
	}
	refresh, _, err := token.Generate(s.jwtSecret, u.UserID, u.InmateNumber, 7*24*time.Hour)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to sign refresh token")
		return
	}

	s.mu.Lock()
	s.refreshToUser[refresh] = u.UserID
	s.mu.Unlock()

	httpx.WriteJSON(w, http.StatusOK, registerResponse{
		UserID:       u.UserID,
		InmateNumber: u.InmateNumber,
		BarcodeURL:   fmt.Sprintf("/assets/barcodes/%s.png", u.InmateNumber),
		AccessToken:  access,
		RefreshToken: refresh,
	})
}

func (s *server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req refreshRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	s.mu.RLock()
	userID, ok := s.refreshToUser[req.RefreshToken]
	if !ok {
		s.mu.RUnlock()
		httpx.WriteError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	u := s.usersByID[userID]
	s.mu.RUnlock()

	access, _, err := token.Generate(s.jwtSecret, u.UserID, u.InmateNumber, 15*time.Minute)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to sign access token")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, registerResponse{
		UserID:       u.UserID,
		InmateNumber: u.InmateNumber,
		BarcodeURL:   fmt.Sprintf("/assets/barcodes/%s.png", u.InmateNumber),
		AccessToken:  access,
		RefreshToken: req.RefreshToken,
	})
}

func (s *server) deleteUser(u user) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.usersByInmate, u.InmateNumber)
	delete(s.usersByID, u.UserID)
}

func (s *server) hashPIN(pin string) []byte {
	h := sha256.New()
	_, _ = h.Write(s.jwtSecret)
	_, _ = h.Write([]byte(":"))
	_, _ = h.Write([]byte(pin))
	return h.Sum(nil)
}

func (s *server) comparePIN(stored []byte, pin string) bool {
	computed := s.hashPIN(pin)
	return hmac.Equal(stored, computed)
}

func (s *server) bootstrapProfile(u user, idem string) error {
	payload := map[string]any{
		"user_id":         u.UserID,
		"nickname":        u.Nickname,
		"article":         u.Article,
		"term":            u.Term,
		"cell":            u.Cell,
		"photo_url":       u.PhotoURL,
		"idempotency_key": "profile:" + idem,
	}
	return s.postJSON(s.profileURL+"/internal/profile/init", payload)
}

func (s *server) bootstrapWallet(u user, idem string) error {
	payload := map[string]any{
		"user_id":         u.UserID,
		"initial_balance": 0,
		"idempotency_key": "wallet:" + idem,
	}
	return s.postJSON(s.walletURL+"/internal/wallet/init", payload)
}

func normalizeExperienceMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "punish" {
		return "punish"
	}
	return "repair"
}

func (s *server) bootstrapProgression(u user, idem, mode string) error {
	payload := map[string]any{
		"user_id":         u.UserID,
		"level":           0,
		"xp":              0,
		"idempotency_key": "progression:" + idem,
		"mode":            normalizeExperienceMode(mode),
	}
	return s.postJSON(s.progressionURL+"/internal/progression/init", payload)
}

func (s *server) postJSON(url string, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func main() {
	addr := envOrDefault("REGISTRY_ADDR", ":8101")
	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("registry-auth service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("registry-auth service failed: %v", err)
	}
}
