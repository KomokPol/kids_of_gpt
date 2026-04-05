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

type snapshot struct {
	Level                 int32    `json:"level"`
	XP                    int64    `json:"xp"`
	WalletBalance         int64    `json:"wallet_balance"`
	DebtBalance           int64    `json:"debt_balance"`
	SearchEnabled         bool     `json:"search_enabled"`
	FiltersEnabled        bool     `json:"filters_enabled"`
	CartLimit             int32    `json:"cart_limit"`
	DeliveryModes         []string `json:"delivery_modes"`
	PreciseETAEnabled     bool     `json:"precise_eta_enabled"`
	MenuChoiceCount       int32    `json:"menu_choice_count"`
	DailySpinAvailable    bool     `json:"daily_spin_available"`
	BannedUntilUnix       int64    `json:"banned_until_unix"`
	CanEarlyUnbanViaTasks bool     `json:"can_early_unban_via_tasks"`
	FilmSubscriptionTier  string   `json:"film_subscription_tier"`
	ProgressPercent       int32    `json:"progress_percent"`
	XPToNextLevel         int64    `json:"xp_to_next_level"`
	CurrentStreakDays     int32    `json:"current_streak_days"`
	LastActivityUnix      int64    `json:"last_activity_unix"`
	ProgressMode          string   `json:"progress_mode"`
	UIBurdenScore         int32    `json:"ui_burden_score"`
	ForcedActionDelayMS   int32    `json:"forced_action_delay_ms"`
	CaptchaEnabled        bool     `json:"captcha_enabled"`
	AdsIntensity          int32    `json:"ads_intensity"`
	MemeBadge             string   `json:"meme_badge"`
	FunnyMessage          string   `json:"funny_message"`
}

type initRequest struct {
	UserID         string `json:"user_id"`
	Level          int32  `json:"level"`
	XP             int64  `json:"xp"`
	IdempotencyKey string `json:"idempotency_key"`
	Mode           string `json:"mode"`
}

type grantXPRequest struct {
	UserID         string `json:"user_id"`
	XP             int64  `json:"xp"`
	Source         string `json:"source"`
	IdempotencyKey string `json:"idempotency_key"`
}

type recalculateRequest struct {
	UserID        string `json:"user_id"`
	WalletBalance int64  `json:"wallet_balance"`
	DebtBalance   int64  `json:"debt_balance"`
}

type syncWalletRequest struct {
	UserID        string `json:"user_id"`
	WalletBalance int64  `json:"wallet_balance"`
	DebtBalance   int64  `json:"debt_balance"`
}

type playerState struct {
	Snapshot        snapshot
	LastActivityDay string
	LastPenaltyDay  string
}

type server struct {
	mu          sync.RWMutex
	state       map[string]playerState
	idempotency map[string]snapshot
}

func newServer() *server {
	return &server{
		state:       make(map[string]playerState),
		idempotency: make(map[string]snapshot),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/internal/progression/init", s.handleInit)
	mux.HandleFunc("/internal/progression/grant-xp", s.handleGrantXP)
	mux.HandleFunc("/internal/progression/recalculate", s.handleRecalculate)
	mux.HandleFunc("/internal/progression/sync-wallet", s.handleSyncWallet)
	mux.HandleFunc("/internal/progression/", s.handleGetSnapshot)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "progression"})
}

func (s *server) handleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req initRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.IdempotencyKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id and idempotency_key are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if snap, ok := s.idempotency["init:"+req.IdempotencyKey]; ok {
		httpx.WriteJSON(w, http.StatusOK, snap)
		return
	}

	now := time.Now().UTC()
	mode := normalizeMode(req.Mode)
	level := req.Level
	if level == 0 {
		level = levelForXP(req.XP)
	}
	snap := makeSnapshot(level, req.XP, 0, 0, mode, 1, now.Unix(), 0)
	st := playerState{
		Snapshot:        snap,
		LastActivityDay: now.Format("2006-01-02"),
		LastPenaltyDay:  "",
	}
	s.state[req.UserID] = st
	s.idempotency["init:"+req.IdempotencyKey] = snap
	httpx.WriteJSON(w, http.StatusOK, snap)
}

func (s *server) handleGrantXP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req grantXPRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" || req.XP <= 0 || req.IdempotencyKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id, xp>0 and idempotency_key are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if snap, ok := s.idempotency["xp:"+req.IdempotencyKey]; ok {
		httpx.WriteJSON(w, http.StatusOK, snap)
		return
	}

	now := time.Now().UTC()
	st, ok := s.state[req.UserID]
	if !ok {
		st = playerState{Snapshot: makeSnapshot(0, 0, 0, 0, "repair", 0, 0, 0)}
	}

	st.Snapshot.XP += req.XP
	st.Snapshot.Level = levelForXP(st.Snapshot.XP)
	st.markActivity(now)

	inactiveDays := daysSince(st.Snapshot.LastActivityUnix, now)
	if inactiveDays < 0 {
		inactiveDays = 0
	}
	updated := makeSnapshot(
		st.Snapshot.Level,
		st.Snapshot.XP,
		st.Snapshot.WalletBalance,
		st.Snapshot.DebtBalance,
		st.Snapshot.ProgressMode,
		st.Snapshot.CurrentStreakDays,
		st.Snapshot.LastActivityUnix,
		inactiveDays,
	)
	if st.Snapshot.BannedUntilUnix != 0 {
		updated.BannedUntilUnix = st.Snapshot.BannedUntilUnix
	}
	st.Snapshot = updated
	s.state[req.UserID] = st
	s.idempotency["xp:"+req.IdempotencyKey] = updated

	httpx.WriteJSON(w, http.StatusOK, updated)
}

func (s *server) handleRecalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req recalculateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	st, ok := s.state[req.UserID]
	if !ok {
		st = playerState{Snapshot: makeSnapshot(0, 0, 0, 0, "repair", 0, 0, 0)}
	}
	inactiveDays := daysSince(st.Snapshot.LastActivityUnix, now)
	if inactiveDays < 0 {
		inactiveDays = 0
	}
	updated := makeSnapshot(
		st.Snapshot.Level,
		st.Snapshot.XP,
		req.WalletBalance,
		req.DebtBalance,
		st.Snapshot.ProgressMode,
		st.Snapshot.CurrentStreakDays,
		st.Snapshot.LastActivityUnix,
		inactiveDays,
	)
	if req.DebtBalance <= -5000 {
		if st.Snapshot.BannedUntilUnix > now.Unix() {
			updated.BannedUntilUnix = st.Snapshot.BannedUntilUnix
		} else {
			updated.BannedUntilUnix = now.Add(12 * time.Hour).Unix()
		}
		updated.CanEarlyUnbanViaTasks = true
	} else if st.Snapshot.BannedUntilUnix > now.Unix() {
		updated.BannedUntilUnix = st.Snapshot.BannedUntilUnix
	}

	st.Snapshot = updated
	s.state[req.UserID] = st
	httpx.WriteJSON(w, http.StatusOK, updated)
}

func (s *server) handleSyncWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req syncWalletRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	st, ok := s.state[req.UserID]
	if !ok {
		st = playerState{Snapshot: makeSnapshot(0, 0, 0, 0, "repair", 0, 0, 0)}
	}
	inactiveDays := daysSince(st.Snapshot.LastActivityUnix, now)
	if inactiveDays < 0 {
		inactiveDays = 0
	}
	updated := makeSnapshot(
		st.Snapshot.Level,
		st.Snapshot.XP,
		req.WalletBalance,
		req.DebtBalance,
		st.Snapshot.ProgressMode,
		st.Snapshot.CurrentStreakDays,
		st.Snapshot.LastActivityUnix,
		inactiveDays,
	)
	if req.DebtBalance <= -5000 {
		if st.Snapshot.BannedUntilUnix > now.Unix() {
			updated.BannedUntilUnix = st.Snapshot.BannedUntilUnix
		} else {
			updated.BannedUntilUnix = now.Add(12 * time.Hour).Unix()
		}
		updated.CanEarlyUnbanViaTasks = true
	} else if st.Snapshot.BannedUntilUnix > now.Unix() {
		updated.BannedUntilUnix = st.Snapshot.BannedUntilUnix
	}

	st.Snapshot = updated
	s.state[req.UserID] = st
	httpx.WriteJSON(w, http.StatusOK, updated)
}

func (s *server) handleGetSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/internal/progression/")
	if !strings.HasSuffix(path, "/snapshot") {
		httpx.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	userID := strings.TrimSuffix(path, "/snapshot")
	userID = strings.TrimSuffix(userID, "/")
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user id is required")
		return
	}

	s.mu.Lock()
	st, ok := s.state[userID]
	if !ok {
		s.mu.Unlock()
		httpx.WriteError(w, http.StatusNotFound, "snapshot not found")
		return
	}
	now := time.Now().UTC()
	st.applyInactivityPenalty(now)
	inactiveDays := daysSince(st.Snapshot.LastActivityUnix, now)
	if inactiveDays < 0 {
		inactiveDays = 0
	}
	st.Snapshot = makeSnapshot(
		st.Snapshot.Level,
		st.Snapshot.XP,
		st.Snapshot.WalletBalance,
		st.Snapshot.DebtBalance,
		st.Snapshot.ProgressMode,
		st.Snapshot.CurrentStreakDays,
		st.Snapshot.LastActivityUnix,
		inactiveDays,
	)
	snap := st.Snapshot
	if snap.BannedUntilUnix != 0 && snap.BannedUntilUnix < time.Now().UTC().Unix() {
		snap.BannedUntilUnix = 0
		snap.CanEarlyUnbanViaTasks = false
		st.Snapshot = snap
	}
	s.state[userID] = st
	s.mu.Unlock()

	httpx.WriteJSON(w, http.StatusOK, snap)
}

func makeSnapshot(level int32, xp, walletBalance, debtBalance int64, mode string, streak int32, lastActivityUnix int64, inactiveDays int32) snapshot {
	mode = normalizeMode(mode)
	cartLimit := int32(1)
	if level >= 3 {
		cartLimit = 3
	} else if level >= 1 {
		cartLimit = 2
	}

	deliveryModes := []string{"as_is"}
	if level >= 2 {
		deliveryModes = append(deliveryModes, "standard")
	}
	if level >= 4 {
		deliveryModes = append(deliveryModes, "express")
	}

	tier := "pervohod"
	if level >= 6 {
		tier = "smotryashiy"
	} else if level >= 3 {
		tier = "patsan"
	}

	menuChoiceCount := int32(1 + 2*level)
	if menuChoiceCount > 10 {
		menuChoiceCount = 10
	}

	xpInLevel := xp % 100
	if xpInLevel < 0 {
		xpInLevel = 0
	}
	progressPercent := int32(xpInLevel)
	xpToNext := int64(100 - xpInLevel)
	if xpToNext <= 0 {
		xpToNext = 100
	}

	uiBurden := int32(40)
	if mode == "repair" {
		uiBurden = clamp(95-level*12+inactiveDays*8, 5, 100)
	} else {
		uiBurden = clamp(25+inactiveDays*15-level*2, 5, 100)
	}
	adsIntensity := clamp(uiBurden/20, 0, 5)

	interfaceState := "patched"
	if uiBurden >= 80 {
		interfaceState = "painful"
	} else if uiBurden >= 55 {
		interfaceState = "legacy"
	} else if uiBurden >= 25 {
		interfaceState = "normal"
	} else {
		interfaceState = "premium"
	}

	return snapshot{
		Level:                 level,
		XP:                    xp,
		WalletBalance:         walletBalance,
		DebtBalance:           debtBalance,
		SearchEnabled:         level >= 2,
		FiltersEnabled:        level >= 3,
		CartLimit:             cartLimit,
		DeliveryModes:         deliveryModes,
		PreciseETAEnabled:     level >= 4,
		MenuChoiceCount:       menuChoiceCount,
		DailySpinAvailable:    true,
		CanEarlyUnbanViaTasks: debtBalance <= -5000,
		FilmSubscriptionTier:  tier,
		ProgressPercent:       progressPercent,
		XPToNextLevel:         xpToNext,
		CurrentStreakDays:     streak,
		LastActivityUnix:      lastActivityUnix,
		ProgressMode:          mode,
		UIBurdenScore:         uiBurden,
		ForcedActionDelayMS:   uiBurden * 35,
		CaptchaEnabled:        uiBurden >= 60,
		AdsIntensity:          adsIntensity,
		MemeBadge:             badgeByLevel(level),
		FunnyMessage:          funnyMessage(interfaceState, mode, streak),
	}
}

func clamp(v, min, max int32) int32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func badgeByLevel(level int32) string {
	switch {
	case level >= 10:
		return "Легенда барака"
	case level >= 6:
		return "Смотрящий по UX"
	case level >= 3:
		return "Пацан интерфейса"
	default:
		return "Новичок-каторжник"
	}
}

func funnyMessage(interfaceState, mode string, streak int32) string {
	if interfaceState == "painful" {
		return "Интерфейс скрипит как двери в СИЗО. Качай активность, и будет легче."
	}
	if mode == "punish" && streak == 0 {
		return "Стрик потерян, сервис обиделся и включил режим придирок."
	}
	if streak >= 5 {
		return "Стрик жарит. Сервис уважает и убирает лишнюю боль."
	}
	return "Система наблюдает. Еще пара действий и станет приятнее."
}

func levelForXP(xp int64) int32 {
	if xp <= 0 {
		return 0
	}
	return int32(xp / 100)
}

func normalizeMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "punish" {
		return mode
	}
	return "repair"
}

func daysSince(lastActivityUnix int64, now time.Time) int32 {
	if lastActivityUnix == 0 {
		return 0
	}
	last := time.Unix(lastActivityUnix, 0).UTC()
	if now.Before(last) {
		return 0
	}
	days := int32(now.Sub(last).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

func (s *playerState) markActivity(now time.Time) {
	today := now.UTC().Format("2006-01-02")
	if s.LastActivityDay == "" {
		s.Snapshot.CurrentStreakDays = 1
	} else {
		last, err1 := time.Parse("2006-01-02", s.LastActivityDay)
		todayDate, err2 := time.Parse("2006-01-02", today)
		if err1 != nil || err2 != nil {
			s.Snapshot.CurrentStreakDays = 1
		} else {
			deltaDays := int(todayDate.Sub(last).Hours() / 24)
			if deltaDays == 0 {
				// no changes
			} else if deltaDays == 1 {
				s.Snapshot.CurrentStreakDays++
			} else {
				s.Snapshot.CurrentStreakDays = 1
			}
		}
	}
	s.LastActivityDay = today
	s.LastPenaltyDay = ""
	s.Snapshot.LastActivityUnix = now.Unix()
}

func (s *playerState) applyInactivityPenalty(now time.Time) {
	if s.Snapshot.LastActivityUnix == 0 {
		return
	}
	inactiveDays := daysSince(s.Snapshot.LastActivityUnix, now)
	if inactiveDays <= 0 {
		return
	}
	today := now.UTC().Format("2006-01-02")
	if s.LastPenaltyDay == today {
		return
	}

	var penalty int64
	mode := normalizeMode(s.Snapshot.ProgressMode)
	if mode == "punish" {
		penalty = 20 * int64(inactiveDays)
	} else if inactiveDays > 1 {
		penalty = 8 * int64(inactiveDays-1)
	}

	if penalty > 0 {
		s.Snapshot.XP -= penalty
		if s.Snapshot.XP < 0 {
			s.Snapshot.XP = 0
		}
		s.Snapshot.Level = levelForXP(s.Snapshot.XP)
		s.Snapshot.CurrentStreakDays = 0
	}
	s.LastPenaltyDay = today
}

func main() {
	addr := os.Getenv("PROGRESSION_ADDR")
	if addr == "" {
		addr = ":8104"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: newServer().routes(),
	}

	log.Printf("progression service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("progression service failed: %v", err)
	}
}
