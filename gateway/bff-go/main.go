package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
	"github.com/KomokPol/kids_of_gpt/services/go/common/token"
)

type gateway struct {
	registryURL     string
	profileURL      string
	walletURL       string
	progressionURL  string
	sharagaURL      string
	burmaldaURL     string
	barygiURL       string
	kinoshkaURL     string
	balandaURL      string
	leaderboardURL  string
	notificationURL string
	catalogURL      string
	jwtSecret       []byte
	httpClient      *http.Client
}

func newGateway() *gateway {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-jwt-secret"
	}
	return &gateway{
		registryURL:     envOrDefault("REGISTRY_BASE_URL", "http://localhost:8101"),
		profileURL:      envOrDefault("PROFILE_BASE_URL", "http://localhost:8102"),
		walletURL:       envOrDefault("WALLET_BASE_URL", "http://localhost:8103"),
		progressionURL:  envOrDefault("PROGRESSION_BASE_URL", "http://localhost:8104"),
		sharagaURL:      envOrDefault("SHARAGA_BASE_URL", "http://localhost:8105"),
		burmaldaURL:     envOrDefault("BURMALDA_BASE_URL", "http://localhost:8112"),
		barygiURL:       envOrDefault("BARYGI_BASE_URL", "http://localhost:8113"),
		kinoshkaURL:     envOrDefault("KINOSHKA_BASE_URL", "http://localhost:8114"),
		balandaURL:      envOrDefault("BALANDA_BASE_URL", "http://localhost:8115"),
		leaderboardURL:  envOrDefault("LEADERBOARD_BASE_URL", "http://localhost:8108"),
		notificationURL: envOrDefault("NOTIFICATION_BASE_URL", "http://localhost:8109"),
		catalogURL:      envOrDefault("CATALOG_BASE_URL", "http://localhost:8110"),
		jwtSecret:       []byte(secret),
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

func (g *gateway) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", g.handleHealth)

	// Auth (no auth required)
	mux.HandleFunc("/api/v1/auth/register", g.handleRegister)
	mux.HandleFunc("/api/v1/auth/login", g.handleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", g.handleRefresh)

	// Wave A services
	mux.HandleFunc("/api/v1/profile/me", g.handleProfileMe)
	mux.HandleFunc("/api/v1/wallet/balance", g.handleWalletBalance)
	mux.HandleFunc("/api/v1/progression/snapshot", g.handleProgressionSnapshot)
	mux.HandleFunc("/api/v1/progression/progress-bar", g.handleProgressBar)
	mux.HandleFunc("/api/v1/sharaga/challenges/", g.handleRunChallenge)

	// Burmalda
	mux.HandleFunc("/api/v1/burmalda/spin", g.handleBurmaldaSpin)
	mux.HandleFunc("/api/v1/burmalda/state", g.handleBurmaldaState)

	// Barygi
	mux.HandleFunc("/api/v1/barygi/products", g.handleBarygiProducts)
	mux.HandleFunc("/api/v1/barygi/cart/add", g.handleBarygiCartAdd)
	mux.HandleFunc("/api/v1/barygi/cart", g.handleBarygiCart)
	mux.HandleFunc("/api/v1/barygi/checkout", g.handleBarygiCheckout)

	// Kinoshka
	mux.HandleFunc("/api/v1/kinoshka/films", g.handleKinoshkaFilms)
	mux.HandleFunc("/api/v1/kinoshka/watch", g.handleKinoshkaWatch)

	// Balanda
	mux.HandleFunc("/api/v1/balanda/menu", g.handleBalandaMenu)
	mux.HandleFunc("/api/v1/balanda/order", g.handleBalandaOrder)
	mux.HandleFunc("/api/v1/balanda/orders", g.handleBalandaOrders)

	// Leaderboard
	mux.HandleFunc("/api/v1/leaderboard", g.handleLeaderboard)

	// Notifications
	mux.HandleFunc("/api/v1/notifications", g.handleNotifications)

	// Catalog
	mux.HandleFunc("/api/v1/catalog", g.handleCatalog)

	return mux
}

func (g *gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "gateway-bff"})
}

// --- Auth routes (no auth) ---

func (g *gateway) handleRegister(w http.ResponseWriter, r *http.Request) {
	g.proxyWithBody(w, r, http.MethodPost, g.registryURL+"/internal/register")
}

func (g *gateway) handleLogin(w http.ResponseWriter, r *http.Request) {
	g.proxyWithBody(w, r, http.MethodPost, g.registryURL+"/internal/login")
}

func (g *gateway) handleRefresh(w http.ResponseWriter, r *http.Request) {
	g.proxyWithBody(w, r, http.MethodPost, g.registryURL+"/internal/refresh")
}

// --- Wave A routes ---

func (g *gateway) handleProfileMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/profile/%s", g.profileURL, claims.UserID))
}

func (g *gateway) handleWalletBalance(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/wallet/%s/balance", g.walletURL, claims.UserID))
}

func (g *gateway) handleProgressionSnapshot(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/progression/%s/snapshot", g.progressionURL, claims.UserID))
}

func (g *gateway) handleProgressBar(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/progression/%s/snapshot", g.progressionURL, claims.UserID))
}

type runChallengeRequest struct {
	Mode           string `json:"mode"`
	IdempotencyKey string `json:"idempotencyKey"`
}

func (g *gateway) handleRunChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/sharaga/challenges/")
	if !strings.HasSuffix(path, "/run") {
		httpx.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	challengeID := strings.TrimSuffix(path, "/run")
	challengeID = strings.TrimSuffix(challengeID, "/")
	if challengeID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "challenge id is required")
		return
	}

	var req runChallengeRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.IdempotencyKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "idempotencyKey is required")
		return
	}

	payload := map[string]any{
		"user_id":         claims.UserID,
		"challenge_id":    challengeID,
		"mode":            req.Mode,
		"idempotency_key": req.IdempotencyKey,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to encode request")
		return
	}

	out, status, err := g.doJSON(http.MethodPost, g.sharagaURL+"/internal/sharaga/run", body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "upstream service unavailable")
		return
	}
	writeRawJSON(w, status, out)
}

// --- Burmalda routes ---

func (g *gateway) handleBurmaldaSpin(w http.ResponseWriter, r *http.Request) {
	g.proxyPostWithUserID(w, r, g.burmaldaURL+"/internal/burmalda/spin")
}

func (g *gateway) handleBurmaldaState(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/burmalda/%s/state", g.burmaldaURL, claims.UserID))
}

// --- Barygi routes ---

func (g *gateway) handleBarygiProducts(w http.ResponseWriter, r *http.Request) {
	_, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, g.barygiURL+"/internal/barygi/products")
}

func (g *gateway) handleBarygiCartAdd(w http.ResponseWriter, r *http.Request) {
	g.proxyPostWithUserID(w, r, g.barygiURL+"/internal/barygi/cart/add")
}

func (g *gateway) handleBarygiCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/barygi/%s/cart", g.barygiURL, claims.UserID))
}

func (g *gateway) handleBarygiCheckout(w http.ResponseWriter, r *http.Request) {
	g.proxyPostWithUserID(w, r, g.barygiURL+"/internal/barygi/checkout")
}

// --- Kinoshka routes ---

func (g *gateway) handleKinoshkaFilms(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/kinoshka/%s/films", g.kinoshkaURL, claims.UserID))
}

func (g *gateway) handleKinoshkaWatch(w http.ResponseWriter, r *http.Request) {
	g.proxyPostWithUserID(w, r, g.kinoshkaURL+"/internal/kinoshka/watch")
}

// --- Balanda routes ---

func (g *gateway) handleBalandaMenu(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/balanda/%s/menu", g.balandaURL, claims.UserID))
}

func (g *gateway) handleBalandaOrder(w http.ResponseWriter, r *http.Request) {
	g.proxyPostWithUserID(w, r, g.balandaURL+"/internal/balanda/order")
}

func (g *gateway) handleBalandaOrders(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/balanda/%s/orders", g.balandaURL, claims.UserID))
}

// --- Leaderboard ---

func (g *gateway) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	_, ok := g.authorize(w, r)
	if !ok {
		return
	}
	query := r.URL.RawQuery
	url := g.leaderboardURL + "/internal/leaderboard/top"
	if query != "" {
		url += "?" + query
	}
	g.proxyNoBody(w, http.MethodGet, url)
}

// --- Notifications ---

func (g *gateway) handleNotifications(w http.ResponseWriter, r *http.Request) {
	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}
	g.proxyNoBody(w, http.MethodGet, fmt.Sprintf("%s/internal/notification/%s/list", g.notificationURL, claims.UserID))
}

// --- Catalog ---

func (g *gateway) handleCatalog(w http.ResponseWriter, r *http.Request) {
	_, ok := g.authorize(w, r)
	if !ok {
		return
	}
	query := r.URL.RawQuery
	url := g.catalogURL + "/internal/catalog/list"
	if query != "" {
		url += "?" + query
	}
	g.proxyNoBody(w, http.MethodGet, url)
}

// --- Helpers ---

func (g *gateway) authorize(w http.ResponseWriter, r *http.Request) (*token.Claims, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authorization header")
		return nil, false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid authorization header")
		return nil, false
	}

	claims, err := token.Parse(g.jwtSecret, parts[1])
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid token")
		return nil, false
	}
	return claims, true
}

// proxyPostWithUserID reads body, injects user_id from JWT, forwards as POST.
func (g *gateway) proxyPostWithUserID(w http.ResponseWriter, r *http.Request, url string) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := g.authorize(w, r)
	if !ok {
		return
	}

	var req map[string]any
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "failed to read request")
		return
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
	} else {
		req = make(map[string]any)
	}

	req["user_id"] = claims.UserID
	newBody, err := json.Marshal(req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to encode request")
		return
	}

	out, status, err := g.doJSON(http.MethodPost, url, newBody)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "upstream service unavailable")
		return
	}
	writeRawJSON(w, status, out)
}

func (g *gateway) proxyWithBody(w http.ResponseWriter, r *http.Request, method, url string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "failed to read request")
		return
	}
	out, status, err := g.doJSON(method, url, body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "upstream service unavailable")
		return
	}
	writeRawJSON(w, status, out)
}

func (g *gateway) proxyNoBody(w http.ResponseWriter, method, url string) {
	out, status, err := g.doJSON(method, url, nil)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "upstream service unavailable")
		return
	}
	writeRawJSON(w, status, out)
}

func (g *gateway) doJSON(method, url string, body []byte) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return out, resp.StatusCode, nil
}

func writeRawJSON(w http.ResponseWriter, status int, raw []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if len(raw) == 0 {
		_, _ = w.Write([]byte("{}"))
		return
	}
	_, _ = w.Write(raw)
}

func main() {
	addr := envOrDefault("GATEWAY_ADDR", ":8080")
	srv := &http.Server{
		Addr:    addr,
		Handler: newGateway().routes(),
	}

	log.Printf("gateway-bff listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("gateway failed: %v", err)
	}
}
