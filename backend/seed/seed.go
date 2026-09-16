package seed

import (
	"context"
	"time"

	"media-sequencer/db"
	"media-sequencer/models"
)

// GetDefaultSeedWindows generates sample playlists for Windows A, B, and C
func GetDefaultSeedWindows(startTime time.Time) []models.Window {
	return []models.Window{
		{
			ID:             "A",
			CycleStartedAt: startTime,
			Playlist: []models.MediaItem{
				{
					ID:          "m1_a",
					Type:        "image",
					URL:         "https://picsum.photos/id/10/800/600",
					DurationSec: 8,
				},
				{
					ID:          "m2_a",
					Type:        "video",
					URL:         "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4",
					DurationSec: 10,
				},
				{
					ID:          "m3_a",
					Type:        "image",
					URL:         "https://picsum.photos/id/20/800/600",
					DurationSec: 6,
				},
			},
		},
		{
			ID:             "B",
			CycleStartedAt: startTime,
			Playlist: []models.MediaItem{
				{
					ID:          "m1_b",
					Type:        "video",
					URL:         "https://www.w3schools.com/html/mov_bbb.mp4",
					DurationSec: 10,
				},
				{
					ID:          "m2_b",
					Type:        "image",
					URL:         "https://picsum.photos/id/30/800/600",
					DurationSec: 10,
				},
				{
					ID:          "m3_b",
					Type:        "image",
					URL:         "https://picsum.photos/id/40/800/600",
					DurationSec: 7,
				},
			},
		},
		{
			ID:             "C",
			CycleStartedAt: startTime,
			Playlist: []models.MediaItem{
				{
					ID:          "m1_c",
					Type:        "image",
					URL:         "https://picsum.photos/id/50/800/600",
					DurationSec: 5,
				},
				{
					ID:          "m2_c",
					Type:        "image",
					URL:         "https://picsum.photos/id/60/800/600",
					DurationSec: 9,
				},
				{
					ID:          "m3_c",
					Type:        "video",
					URL:         "https://vjs.zencdn.net/v/oceans.mp4",
					DurationSec: 12,
				},
			},
		},
	}
}

// SeedInitialData populates store with windows A, B, and C
func SeedInitialData(ctx context.Context, store db.Store) error {
	existing, err := store.GetAllWindows(ctx)
	if err == nil && len(existing) >= 3 {
		// Overwrite seed with fresh valid media URLs
		startTime := time.Now().UTC()
		return store.SeedWindows(ctx, GetDefaultSeedWindows(startTime))
	}
	startTime := time.Now().UTC()
	return store.SeedWindows(ctx, GetDefaultSeedWindows(startTime))
}
