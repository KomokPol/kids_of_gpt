package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

type catalogEntry struct {
	ItemID      string `json:"item_id"`
	Title       string `json:"title"`
	Kind        string `json:"kind"` // "product", "film", "menu_item"
	Price       int64  `json:"price"`
	Description string `json:"description"`
}

type server struct {
	mu    sync.RWMutex
	items map[string]catalogEntry
}

func newServer() *server {
	s := &server{
		items: make(map[string]catalogEntry),
	}
	// Pre-seed catalog
	seed := []catalogEntry{
		{ItemID: "prod-1", Title: "Сигареты Прима", Kind: "product", Price: 50, Description: "Классика"},
		{ItemID: "prod-2", Title: "Чай грузинский", Kind: "product", Price: 30, Description: "Крепкий"},
		{ItemID: "prod-3", Title: "Тушёнка", Kind: "product", Price: 80, Description: "Говядина"},
		{ItemID: "prod-4", Title: "Сгущёнка", Kind: "product", Price: 40, Description: "Молочная"},
		{ItemID: "prod-5", Title: "Лапша Доширак", Kind: "product", Price: 20, Description: "Быстрая"},
		{ItemID: "film-1", Title: "Побег из Шоушенка", Kind: "film", Price: 0, Description: "pervohod"},
		{ItemID: "film-2", Title: "Зелёная миля", Kind: "film", Price: 0, Description: "pervohod"},
		{ItemID: "film-3", Title: "Бригада", Kind: "film", Price: 100, Description: "patsan"},
		{ItemID: "film-4", Title: "Крёстный отец", Kind: "film", Price: 150, Description: "patsan"},
		{ItemID: "film-5", Title: "Джентльмены", Kind: "film", Price: 300, Description: "smotryashiy"},
		{ItemID: "menu-1", Title: "Баланда классическая", Kind: "menu_item", Price: 15, Description: "Суп дня"},
		{ItemID: "menu-2", Title: "Каша перловая", Kind: "menu_item", Price: 10, Description: "На воде"},
		{ItemID: "menu-3", Title: "Хлеб чёрный", Kind: "menu_item", Price: 5, Description: "Ржаной"},
		{ItemID: "menu-4", Title: "Чай без сахара", Kind: "menu_item", Price: 3, Description: "Горячий"},
		{ItemID: "menu-5", Title: "Котлета рубленая", Kind: "menu_item", Price: 25, Description: "Мясная"},
		{ItemID: "menu-6", Title: "Картошка варёная", Kind: "menu_item", Price: 12, Description: "Гарнир"},
		{ItemID: "menu-7", Title: "Компот", Kind: "menu_item", Price: 8, Description: "Из сухофруктов"},
		{ItemID: "menu-8", Title: "Сосиска", Kind: "menu_item", Price: 20, Description: "Варёная"},
		{ItemID: "menu-9", Title: "Макароны по-флотски", Kind: "menu_item", Price: 18, Description: "С фаршем"},
		{ItemID: "menu-10", Title: "Борщ", Kind: "menu_item", Price: 22, Description: "Красный"},
	}
	for _, e := range seed {
		s.items[e.ItemID] = e
	}
	return s
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/catalog/list", s.handleList)
	mux.HandleFunc("/internal/catalog/seed", s.handleSeed)
	mux.HandleFunc("/internal/catalog/", s.handleGetItem)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "catalog"})
}

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	kind := r.URL.Query().Get("kind")

	s.mu.RLock()
	var result []catalogEntry
	for _, e := range s.items {
		if kind == "" || e.Kind == kind {
			result = append(result, e)
		}
	}
	s.mu.RUnlock()

	if result == nil {
		result = []catalogEntry{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"entries": result})
}

func (s *server) handleSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var entries []catalogEntry
	if err := httpx.DecodeJSON(r, &entries); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	s.mu.Lock()
	for _, e := range entries {
		if e.ItemID != "" {
			s.items[e.ItemID] = e
		}
	}
	s.mu.Unlock()

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"seeded": len(entries)})
}

func (s *server) handleGetItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	itemID := strings.TrimPrefix(r.URL.Path, "/internal/catalog/")
	itemID = strings.TrimSuffix(itemID, "/")
	if itemID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "item_id is required")
		return
	}

	s.mu.RLock()
	entry, ok := s.items[itemID]
	s.mu.RUnlock()
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, entry)
}

func main() {
	addr := os.Getenv("CATALOG_ADDR")
	if addr == "" {
		addr = ":8110"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("catalog service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("catalog service failed: %v", err)
	}
}
