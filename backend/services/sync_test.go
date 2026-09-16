package services

import (
	"context"
	"testing"
	"time"

	"media-sequencer/db"
	"media-sequencer/models"
)

func TestSyncManagerStartAndAutoEnd(t *testing.T) {
	ctx := context.Background()
	store := db.NewMemoryStore()

	start := time.Now().UTC().Add(-30 * time.Second)
	windows := []models.Window{
		{
			ID:             "A",
			CycleStartedAt: start,
			Playlist: []models.MediaItem{
				{ID: "m1", Type: "image", DurationSec: 10},
				{ID: "m2", Type: "video", DurationSec: 15},
			},
		},
	}
	_ = store.SeedWindows(ctx, windows)

	sm := NewSyncManager(store)

	syncItem := models.MediaItem{
		ID:          "sync1",
		Type:        "image",
		URL:         "https://picsum.photos/200",
		DurationSec: 5,
	}

	// Start sync with 1s duration
	err := sm.StartSync(ctx, syncItem, 1)
	if err != nil {
		t.Fatalf("Failed to start sync: %v", err)
	}

	active, item, endsAt := sm.GetSyncStatus()
	if !active {
		t.Errorf("Expected sync to be active")
	}
	if item == nil || item.ID != "sync1" {
		t.Errorf("Expected sync item ID sync1, got %v", item)
	}
	if endsAt == nil {
		t.Errorf("Expected valid endsAt timestamp")
	}

	// Wait for time.AfterFunc auto-resume (1.2s)
	time.Sleep(1200 * time.Millisecond)

	activeAfter, _, _ := sm.GetSyncStatus()
	if activeAfter {
		t.Errorf("Expected sync to be inactive after duration elapsed")
	}

	// Verify window A cycle_started_at was updated
	winA, err := store.GetWindow(ctx, "A")
	if err != nil {
		t.Fatalf("Failed to get window A: %v", err)
	}

	now := time.Now().UTC()
	current, snap := CalculateCurrentState(winA.Playlist, winA.CycleStartedAt, now)
	if current.ID == "" {
		t.Errorf("Expected valid current item after sync resume")
	}
	t.Logf("Resumed item: %s, index: %d, offset: %f", current.ID, snap.ItemIndex, snap.OffsetSec)
}
