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
					URL:         "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4",
					DurationSec: 15, // 0 in json or natural length, stored as 15 for cycle math estimation
				},
				{
					ID:          "m3_a",
					Type:        "image",
					URL:         "https://picsum.photos/id/20/800/600",
					DurationSec: 6,
				},
				{
					ID:          "m4_a",
					Type:        "video",
					URL:         "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4",
					DurationSec: 15,
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
					URL:         "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4",
					DurationSec: 15,
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
				{
					ID:          "m4_b",
					Type:        "video",
					URL:         "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4",
					DurationSec: 15,
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
					URL:         "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/TearsOfSteel.mp4",
					DurationSec: 15,
				},
				{
					ID:          "m4_c",
					Type:        "image",
					URL:         "https://picsum.photos/id/70/800/600",
					DurationSec: 8,
				},
			},
		},
	}
}

// SeedInitialData populates store with windows A, B, and C if they do not exist
func SeedInitialData(ctx context.Context, store db.Store) error {
	existing, err := store.GetAllWindows(ctx)
	if err == nil && len(existing) >= 3 {
		return nil // already seeded
	}
	startTime := time.Now().UTC()
	return store.SeedWindows(ctx, GetDefaultSeedWindows(startTime))
}
