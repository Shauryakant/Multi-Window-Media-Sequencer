package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"media-sequencer/db"
	"media-sequencer/models"
	"media-sequencer/services"
)

type SyncProvider interface {
	GetSyncStatus() (bool, *models.MediaItem, *time.Time)
}

type Handler struct {
	store db.Store
	sync  SyncProvider
}

func NewHandler(store db.Store, sync SyncProvider) *Handler {
	return &Handler{
		store: store,
		sync:  sync,
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) GetState(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	windows, err := h.store.GetAllWindows(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	windowStates := make(map[string]models.WindowPlaybackState)

	for _, win := range windows {
		item, snap := services.CalculateCurrentState(win.Playlist, win.CycleStartedAt, now)
		windowStates[win.ID] = models.WindowPlaybackState{
			CurrentItem: item,
			OffsetSec:   snap.OffsetSec,
		}
	}

	syncActive, syncItem, syncEndsAt := false, (*models.MediaItem)(nil), (*time.Time)(nil)
	if h.sync != nil {
		syncActive, syncItem, syncEndsAt = h.sync.GetSyncStatus()
	}

	resp := models.SyncStateResponse{
		SyncActive: syncActive,
		SyncItem:   syncItem,
		SyncEndsAt: syncEndsAt,
		Windows:    windowStates,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	windowID := r.PathValue("id")
	if windowID == "" {
		windowID = r.URL.Query().Get("id")
	}

	win, err := h.store.GetWindow(r.Context(), windowID)
	if err != nil {
		http.Error(w, "Window not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(win)
}

type AddMediaRequest struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	URL         string  `json:"url"`
	DurationSec float64 `json:"duration_sec"`
}

func (h *Handler) AddMedia(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	windowID := r.PathValue("id")
	if windowID == "" {
		windowID = r.URL.Query().Get("id")
	}

	var req AddMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		req.Type = "image"
	}
	if req.ID == "" {
		req.ID = "m_" + time.Now().Format("150405.000")
	}
	if req.Type == "image" && req.DurationSec <= 0 {
		req.DurationSec = 5.0
	}

	item := models.MediaItem{
		ID:          req.ID,
		Type:        req.Type,
		URL:         req.URL,
		DurationSec: req.DurationSec,
	}

	if err := h.store.AppendMedia(r.Context(), windowID, item); err != nil {
		http.Error(w, "Failed to append media: "+err.Error(), http.StatusInternalServerError)
		return
	}

	updatedWin, err := h.store.GetWindow(r.Context(), windowID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(item)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(updatedWin)
}
