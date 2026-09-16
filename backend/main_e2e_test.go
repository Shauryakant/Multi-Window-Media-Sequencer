package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"media-sequencer/db"
	"media-sequencer/handlers"
	"media-sequencer/models"
	"media-sequencer/seed"
	"media-sequencer/services"
)

func TestFullE2EFlow(t *testing.T) {
	ctx := context.Background()
	store := db.NewMemoryStore()

	// 1. Seed initial data
	if err := seed.SeedInitialData(ctx, store); err != nil {
		t.Fatalf("Failed to seed data: %v", err)
	}

	syncManager := services.NewSyncManager(store)
	h := handlers.NewHandler(store, syncManager, 2)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /api/state", h.GetState)
	mux.HandleFunc("GET /api/windows/{id}/playlist", h.GetPlaylist)
	mux.HandleFunc("POST /api/windows/{id}/media", h.AddMedia)
	mux.HandleFunc("POST /api/sync", h.TriggerSync)

	server := httptest.Server{
		Config: &http.Server{Handler: mux},
	}

	// Test 1: GET /health
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("GET", "/health", nil)
	mux.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("Expected health 200, got %d", w1.Code)
	}

	// Test 2: GET /api/state
	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("GET", "/api/state", nil)
	mux.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("Expected state 200, got %d", w2.Code)
	}
	var stateResp models.SyncStateResponse
	_ = json.NewDecoder(w2.Body).Decode(&stateResp)
	if stateResp.SyncActive {
		t.Errorf("Expected sync_active false initially")
	}
	if len(stateResp.Windows) != 3 {
		t.Errorf("Expected 3 windows, got %d", len(stateResp.Windows))
	}

	// Test 3: POST /api/windows/A/media
	addBody, _ := json.Marshal(map[string]interface{}{
		"type":         "image",
		"url":          "https://picsum.photos/800/600",
		"duration_sec": 7,
	})
	w3 := httptest.NewRecorder()
	r3, _ := http.NewRequest("POST", "/api/windows/A/media", bytes.NewBuffer(addBody))
	r3.SetPathValue("id", "A")
	mux.ServeHTTP(w3, r3)
	if w3.Code != http.StatusCreated {
		t.Fatalf("Expected add media 201, got %d", w3.Code)
	}

	// Test 4: POST /api/sync
	syncBody, _ := json.Marshal(map[string]interface{}{
		"id":           "m1_a",
		"duration_sec": 1,
	})
	w4 := httptest.NewRecorder()
	r4, _ := http.NewRequest("POST", "/api/sync", bytes.NewBuffer(syncBody))
	mux.ServeHTTP(w4, r4)
	if w4.Code != http.StatusOK {
		t.Fatalf("Expected sync trigger 200, got %d: %s", w4.Code, w4.Body.String())
	}

	// Verify sync is active
	w5 := httptest.NewRecorder()
	r5, _ := http.NewRequest("GET", "/api/state", nil)
	mux.ServeHTTP(w5, r5)
	var stateSync models.SyncStateResponse
	_ = json.NewDecoder(w5.Body).Decode(&stateSync)
	if !stateSync.SyncActive {
		t.Errorf("Expected sync_active to be true after POST /api/sync")
	}

	// Wait for sync to auto-end (1.2s)
	time.Sleep(1200 * time.Millisecond)

	w6 := httptest.NewRecorder()
	r6, _ := http.NewRequest("GET", "/api/state", nil)
	mux.ServeHTTP(w6, r6)
	var stateResumed models.SyncStateResponse
	_ = json.NewDecoder(w6.Body).Decode(&stateResumed)
	if stateResumed.SyncActive {
		t.Errorf("Expected sync_active to be false after duration elapsed")
	}

	_ = server
}
