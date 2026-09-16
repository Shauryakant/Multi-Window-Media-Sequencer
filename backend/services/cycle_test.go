package services

import (
	"testing"
	"time"

	"media-sequencer/models"
)

func TestCalculateCurrentState(t *testing.T) {
	playlist := []models.MediaItem{
		{ID: "m1", Type: "image", DurationSec: 10},
		{ID: "m2", Type: "video", DurationSec: 0}, // default 15s
		{ID: "m3", Type: "image", DurationSec: 5},
	}
	// Total duration = 10 + 15 + 5 = 30 seconds

	start := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)

	// Test t = start (elapsed 0s) -> m1, offset 0s
	item, snap := CalculateCurrentState(playlist, start, start)
	if item.ID != "m1" || snap.ItemIndex != 0 || snap.OffsetSec != 0 {
		t.Errorf("Expected m1 index 0 offset 0, got %s index %d offset %f", item.ID, snap.ItemIndex, snap.OffsetSec)
	}

	// Test t = start + 5s -> m1, offset 5s
	t5 := start.Add(5 * time.Second)
	item, snap = CalculateCurrentState(playlist, start, t5)
	if item.ID != "m1" || snap.ItemIndex != 0 || snap.OffsetSec != 5 {
		t.Errorf("Expected m1 index 0 offset 5, got %s index %d offset %f", item.ID, snap.ItemIndex, snap.OffsetSec)
	}

	// Test t = start + 12s -> m2 (starts at 10s), offset 2s
	t12 := start.Add(12 * time.Second)
	item, snap = CalculateCurrentState(playlist, start, t12)
	if item.ID != "m2" || snap.ItemIndex != 1 || snap.OffsetSec != 2 {
		t.Errorf("Expected m2 index 1 offset 2, got %s index %d offset %f", item.ID, snap.ItemIndex, snap.OffsetSec)
	}

	// Test t = start + 27s -> m3 (starts at 25s), offset 2s
	t27 := start.Add(27 * time.Second)
	item, snap = CalculateCurrentState(playlist, start, t27)
	if item.ID != "m3" || snap.ItemIndex != 2 || snap.OffsetSec != 2 {
		t.Errorf("Expected m3 index 2 offset 2, got %s index %d offset %f", item.ID, snap.ItemIndex, snap.OffsetSec)
	}

	// Test t = start + 35s -> wrap around (elapsed 35%30 = 5s) -> m1, offset 5s
	t35 := start.Add(35 * time.Second)
	item, snap = CalculateCurrentState(playlist, start, t35)
	if item.ID != "m1" || snap.ItemIndex != 0 || snap.OffsetSec != 5 {
		t.Errorf("Expected m1 index 0 offset 5 after wrap, got %s index %d offset %f", item.ID, snap.ItemIndex, snap.OffsetSec)
	}
}

func TestComputeNewCycleStartedAtRoundtrip(t *testing.T) {
	playlist := []models.MediaItem{
		{ID: "m1", Type: "image", DurationSec: 10},
		{ID: "m2", Type: "video", DurationSec: 15},
		{ID: "m3", Type: "image", DurationSec: 5},
	}
	// Total duration = 30s

	now := time.Now().UTC()
	targetIndex := 1
	targetOffsetSec := 7.5

	newStart := ComputeNewCycleStartedAt(playlist, targetIndex, targetOffsetSec, now)

	// Verifying: CalculateCurrentState with newStart at `now` must yield targetIndex and targetOffsetSec
	item, snap := CalculateCurrentState(playlist, newStart, now)
	if item.ID != "m2" || snap.ItemIndex != targetIndex || snap.OffsetSec != targetOffsetSec {
		t.Errorf("Roundtrip failed: expected m2 index %d offset %f, got %s index %d offset %f",
			targetIndex, targetOffsetSec, item.ID, snap.ItemIndex, snap.OffsetSec)
	}
}
