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

type menuItem struct {
	ItemID string `json:"item_id"`
	Title  string `json:"title"`
	Price  int64  `json:"price"`
}

type orderRequest struct {
	UserID         string   `json:"user_id"`
	Items          []string `json:"items"` // item_ids
	DeliveryMode   string   `json:"delivery_mode"`
	IdempotencyKey string   `json:"idempotency_key"`
}

type order struct {
	OrderID      string   `json:"order_id"`
	UserID       string   `json:"user_id"`
	Items        []string `json:"items"`
	Total        int64    `json:"total"`
	DeliveryMode string   `json:"delivery_mode"`
	Status       string   `json:"status"`
	Balance      int64    `json:"balance"`
	CreatedAt    int64    `json:"created_at"`
}

type progressionSnapshot struct {
	MenuChoiceCount int32    `json:"menu_choice_count"`
	DeliveryModes   []string `json:"delivery_modes"`
	BannedUntilUnix int64    `json:"banned_until_unix"`
}

type server struct {
	mu             sync.RWMutex
	menu           []menuItem
	menuMap        map[string]menuItem
	orders         map[string]order    // order_id -> order
	userOrders     map[string][]string // user_id -> order_ids
	idempotency    map[string]order
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
	menu := []menuItem{
		{ItemID: "menu-1", Title: "Баланда классическая", Price: 15},
		{ItemID: "menu-2", Title: "Каша перловая", Price: 10},
		{ItemID: "menu-3", Title: "Хлеб чёрный", Price: 5},
		{ItemID: "menu-4", Title: "Чай без сахара", Price: 3},
		{ItemID: "menu-5", Title: "Котлета рубленая", Price: 25},
		{ItemID: "menu-6", Title: "Картошка варёная", Price: 12},
		{ItemID: "menu-7", Title: "Компот", Price: 8},
		{ItemID: "menu-8", Title: "Сосиска", Price: 20},
		{ItemID: "menu-9", Title: "Макароны по-флотски", Price: 18},
		{ItemID: "menu-10", Title: "Борщ", Price: 22},
	}
	menuMap := make(map[string]menuItem)
	for _, m := range menu {
		menuMap[m.ItemID] = m
	}
	return &server{
		menu:           menu,
		menuMap:        menuMap,
		orders:         make(map[string]order),
		userOrders:     make(map[string][]string),
		idempotency:    make(map[string]order),
		walletURL:      envOrDefault("WALLET_BASE_URL", "http://localhost:8103"),
		progressionURL: envOrDefault("PROGRESSION_BASE_URL", "http://localhost:8104"),
		httpClient:     &http.Client{Timeout: 3 * time.Second},
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/balanda/order", s.handleOrder)
	mux.HandleFunc("/internal/balanda/", s.handleDynamic)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "balanda"})
}

func (s *server) handleDynamic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/internal/balanda/")
	if strings.HasSuffix(path, "/menu") {
		s.handleMenu(w, r, strings.TrimSuffix(strings.TrimSuffix(path, "/menu"), "/"))
	} else if strings.HasSuffix(path, "/orders") {
		s.handleUserOrders(w, r, strings.TrimSuffix(strings.TrimSuffix(path, "/orders"), "/"))
	} else {
		httpx.WriteError(w, http.StatusNotFound, "not found")
	}
}

func (s *server) handleMenu(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user id is required")
		return
	}

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

	limit := int(snap.MenuChoiceCount)
	if limit <= 0 {
		limit = 1
	}
	if limit > len(s.menu) {
		limit = len(s.menu)
	}

	result := s.menu[:limit]
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"menu":              result,
		"menu_choice_count": snap.MenuChoiceCount,
		"delivery_modes":    snap.DeliveryModes,
	})
}

func (s *server) handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req orderRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.IdempotencyKey == "" || len(req.Items) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "user_id, items and idempotency_key are required")
		return
	}
	if req.DeliveryMode == "" {
		req.DeliveryMode = "as_is"
	}

	s.mu.RLock()
	if o, ok := s.idempotency[req.IdempotencyKey]; ok {
		s.mu.RUnlock()
		httpx.WriteJSON(w, http.StatusOK, o)
		return
	}
	s.mu.RUnlock()

	// Check progression
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

	// Validate delivery mode
	modeAllowed := false
	for _, m := range snap.DeliveryModes {
		if m == req.DeliveryMode {
			modeAllowed = true
			break
		}
	}
	if !modeAllowed {
		httpx.WriteError(w, http.StatusForbidden,
			fmt.Sprintf("delivery mode '%s' not available for your level", req.DeliveryMode))
		return
	}

	// Validate item count
	if int32(len(req.Items)) > snap.MenuChoiceCount {
		httpx.WriteError(w, http.StatusForbidden, "too many items for your level")
		return
	}

	// Calculate total
	var total int64
	for _, itemID := range req.Items {
		if m, ok := s.menuMap[itemID]; ok {
			total += m.Price
		} else {
			httpx.WriteError(w, http.StatusBadRequest, fmt.Sprintf("unknown menu item: %s", itemID))
			return
		}
	}

	// Delivery surcharge
	switch req.DeliveryMode {
	case "standard":
		total += 20
	case "express":
		total += 50
	}

	// Debit wallet
	payload := map[string]any{
		"user_id":         req.UserID,
		"amount":          total,
		"source":          "balanda:" + req.DeliveryMode,
		"idempotency_key": "balanda:" + req.IdempotencyKey,
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

	o := order{
		OrderID:      fmt.Sprintf("order-balanda-%d", time.Now().UnixNano()),
		UserID:       req.UserID,
		Items:        req.Items,
		Total:        total,
		DeliveryMode: req.DeliveryMode,
		Status:       "confirmed",
		Balance:      walletResp.Balance,
		CreatedAt:    time.Now().UTC().Unix(),
	}

	s.mu.Lock()
	s.orders[o.OrderID] = o
	s.userOrders[req.UserID] = append(s.userOrders[req.UserID], o.OrderID)
	s.idempotency[req.IdempotencyKey] = o
	s.mu.Unlock()

	// Check debt threshold
	if walletResp.Balance < 0 {
		go func() {
			recalc := map[string]any{
				"user_id":        req.UserID,
				"wallet_balance": walletResp.Balance,
				"debt_balance":   walletResp.Balance,
			}
			httpx.PostJSON(s.httpClient, s.progressionURL+"/internal/progression/recalculate", recalc)
		}()
	}

	httpx.WriteJSON(w, http.StatusOK, o)
}

func (s *server) handleUserOrders(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user id is required")
		return
	}

	s.mu.RLock()
	orderIDs := s.userOrders[userID]
	var result []order
	for _, oid := range orderIDs {
		result = append(result, s.orders[oid])
	}
	s.mu.RUnlock()

	if result == nil {
		result = []order{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"orders": result})
}

func main() {
	addr := envOrDefault("BALANDA_ADDR", ":8115")
	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("balanda service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("balanda service failed: %v", err)
	}
}
