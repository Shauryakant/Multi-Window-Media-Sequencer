package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"media-sequencer/db"
	"media-sequencer/models"
)

type SyncManager struct {
	mu          sync.Mutex
	active      bool
	item        models.MediaItem
	startedAt   time.Time
	durationSec int
	paused      map[string]models.PausedSnapshot
	timer       *time.Timer
	store       db.Store
}

func NewSyncManager(store db.Store) *SyncManager {
	return &SyncManager{
		store:  store,
		paused: make(map[string]models.PausedSnapshot),
	}
}

// StartSync snapshots current window playback positions and forces all windows to play `item` for `durationSec`.
func (sm *SyncManager) StartSync(ctx context.Context, item models.MediaItem, durationSec int) error {
	windows, err := sm.store.GetAllWindows(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch windows for sync snapshot: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Stop any existing sync timer
	if sm.timer != nil {
		sm.timer.Stop()
		sm.timer = nil
	}

	now := time.Now().UTC()
	sm.paused = make(map[string]models.PausedSnapshot)

	// Snapshot exact current playback position for each window
	for _, win := range windows {
		_, snap := CalculateCurrentState(win.Playlist, win.CycleStartedAt, now)
		sm.paused[win.ID] = snap
	}

	sm.active = true
	sm.item = item
	sm.startedAt = now
	sm.durationSec = durationSec

	// Schedule automatic end sync
	duration := time.Duration(durationSec) * time.Second
	sm.timer = time.AfterFunc(duration, func() {
		sm.EndSync()
	})

	return nil
}

// EndSync restores each window's cycle_started_at so they resume smoothly from paused position.
func (sm *SyncManager) EndSync() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.active {
		return
	}

	now := time.Now().UTC()
	bgCtx := context.Background()

	// Recompute and persist new cycle_started_at for each window
	windows, err := sm.store.GetAllWindows(bgCtx)
	if err == nil {
		for _, win := range windows {
			snap, ok := sm.paused[win.ID]
			if !ok {
				continue
			}
			newStart := ComputeNewCycleStartedAt(win.Playlist, snap.ItemIndex, snap.OffsetSec, now)
			_ = sm.store.UpdateCycleStartedAt(bgCtx, win.ID, newStart)
		}
	}

	sm.active = false
	sm.paused = make(map[string]models.PausedSnapshot)
	if sm.timer != nil {
		sm.timer.Stop()
		sm.timer = nil
	}
}

// GetSyncStatus returns the active sync status, sync item, and sync expiration time.
func (sm *SyncManager) GetSyncStatus() (bool, *models.MediaItem, *time.Time) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.active {
		return false, nil, nil
	}

	endsAt := sm.startedAt.Add(time.Duration(sm.durationSec) * time.Second)
	item := sm.item
	return true, &item, &endsAt
}

// ResolveMediaItem resolves a media item by ID across all windows or validates raw item.
func (sm *SyncManager) ResolveMediaItem(ctx context.Context, id string, rawItem *models.MediaItem) (*models.MediaItem, error) {
	if id != "" {
		windows, err := sm.store.GetAllWindows(ctx)
		if err != nil {
			return nil, err
		}
		for _, win := range windows {
			for _, item := range win.Playlist {
				if item.ID == id {
					return &item, nil
				}
			}
		}
		return nil, fmt.Errorf("media item with ID %q not found", id)
	}

	if rawItem != nil && rawItem.URL != "" {
		if rawItem.Type == "" {
			rawItem.Type = "image"
		}
		if rawItem.ID == "" {
			rawItem.ID = "sync_" + time.Now().Format("150405.000")
		}
		if rawItem.Type == "image" && rawItem.DurationSec <= 0 {
			rawItem.DurationSec = 5
		}
		return rawItem, nil
	}

	return nil, fmt.Errorf("must specify either an existing media 'id' or raw '{type, url}'")
}
