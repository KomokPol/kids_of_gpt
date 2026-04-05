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

type product struct {
	ItemID string `json:"item_id"`
	Title  string `json:"title"`
	Price  int64  `json:"price"`
}

type cartItem struct {
	ItemID string `json:"item_id"`
	Title  string `json:"title"`
	Price  int64  `json:"price"`
	Qty    int32  `json:"qty"`
}

type addToCartRequest struct {
	UserID string `json:"user_id"`
	ItemID string `json:"item_id"`
	Qty    int32  `json:"qty"`
}

type checkoutRequest struct {
	UserID         string `json:"user_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

type checkoutResponse struct {
	OrderID string `json:"order_id"`
	Total   int64  `json:"total"`
	Balance int64  `json:"balance"`
}

type progressionSnapshot struct {
	CartLimit       int32 `json:"cart_limit"`
	BannedUntilUnix int64 `json:"banned_until_unix"`
}

type server struct {
	mu             sync.RWMutex
	products       map[string]product
	carts          map[string][]cartItem // user_id -> cart
	orders         map[string]checkoutResponse
	idempotency    map[string]checkoutResponse
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
		products:       make(map[string]product),
		carts:          make(map[string][]cartItem),
		orders:         make(map[string]checkoutResponse),
		idempotency:    make(map[string]checkoutResponse),
		walletURL:      envOrDefault("WALLET_BASE_URL", "http://localhost:8103"),
		progressionURL: envOrDefault("PROGRESSION_BASE_URL", "http://localhost:8104"),
		httpClient:     &http.Client{Timeout: 3 * time.Second},
	}
	// Pre-seed products
	for _, p := range []product{
		{ItemID: "prod-1", Title: "Сигареты Прима", Price: 50},
		{ItemID: "prod-2", Title: "Чай грузинский", Price: 30},
		{ItemID: "prod-3", Title: "Тушёнка", Price: 80},
		{ItemID: "prod-4", Title: "Сгущёнка", Price: 40},
		{ItemID: "prod-5", Title: "Лапша Доширак", Price: 20},
	} {
		s.products[p.ItemID] = p
	}
	return s
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/barygi/products", s.handleProducts)
	mux.HandleFunc("/internal/barygi/cart/add", s.handleCartAdd)
	mux.HandleFunc("/internal/barygi/checkout", s.handleCheckout)
	mux.HandleFunc("/internal/barygi/", s.handleGetCart)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "barygi"})
}

func (s *server) handleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.mu.RLock()
	var list []product
	for _, p := range s.products {
		list = append(list, p)
	}
	s.mu.RUnlock()
	if list == nil {
		list = []product{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"products": list})
}

func (s *server) handleCartAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req addToCartRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.ItemID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id and item_id are required")
		return
	}
	if req.Qty <= 0 {
		req.Qty = 1
	}

	// Check progression snapshot for cart_limit and ban
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

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check product exists
	prod, ok := s.products[req.ItemID]
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "product not found")
		return
	}

	// Count unique items in cart
	cart := s.carts[req.UserID]
	uniqueItems := make(map[string]bool)
	for _, ci := range cart {
		uniqueItems[ci.ItemID] = true
	}
	if !uniqueItems[req.ItemID] && int32(len(uniqueItems)) >= snap.CartLimit {
		httpx.WriteError(w, http.StatusForbidden,
			fmt.Sprintf("cart limit reached (%d unique items for your level)", snap.CartLimit))
		return
	}

	// Add or update qty
	found := false
	for i, ci := range cart {
		if ci.ItemID == req.ItemID {
			cart[i].Qty += req.Qty
			found = true
			break
		}
	}
	if !found {
		cart = append(cart, cartItem{
			ItemID: prod.ItemID,
			Title:  prod.Title,
			Price:  prod.Price,
			Qty:    req.Qty,
		})
	}
	s.carts[req.UserID] = cart

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "added", "cart_size": len(cart)})
}

func (s *server) handleGetCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/internal/barygi/")
	if !strings.HasSuffix(path, "/cart") {
		httpx.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	userID := strings.TrimSuffix(path, "/cart")
	userID = strings.TrimSuffix(userID, "/")
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user id is required")
		return
	}

	s.mu.RLock()
	cart := s.carts[userID]
	s.mu.RUnlock()

	var total int64
	for _, ci := range cart {
		total += ci.Price * int64(ci.Qty)
	}
	if cart == nil {
		cart = []cartItem{}
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": cart, "total": total})
}

func (s *server) handleCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req checkoutRequest
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
	cart := s.carts[req.UserID]
	s.mu.RUnlock()

	if len(cart) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "cart is empty")
		return
	}

	var total int64
	for _, ci := range cart {
		total += ci.Price * int64(ci.Qty)
	}

	// Debit wallet
	payload := map[string]any{
		"user_id":         req.UserID,
		"amount":          total,
		"source":          "barygi:checkout",
		"idempotency_key": "barygi:" + req.IdempotencyKey,
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

	orderID := fmt.Sprintf("order-barygi-%d", time.Now().UnixNano())
	resp := checkoutResponse{
		OrderID: orderID,
		Total:   total,
		Balance: walletResp.Balance,
	}

	s.mu.Lock()
	s.idempotency[req.IdempotencyKey] = resp
	s.orders[orderID] = resp
	s.carts[req.UserID] = nil // clear cart
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

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func main() {
	addr := envOrDefault("BARYGI_ADDR", ":8113")
	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("barygi service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("barygi service failed: %v", err)
	}
}
