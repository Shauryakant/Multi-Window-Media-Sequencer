package models

import "time"

type MediaItem struct {
	ID          string  `json:"id" bson:"id"`
	Type        string  `json:"type" bson:"type"`                 // "image" or "video"
	URL         string  `json:"url" bson:"url"`
	DurationSec float64 `json:"duration_sec" bson:"duration_sec"` // sec; for images > 0, for videos can be 0 or estimated
}

type Window struct {
	ID             string      `json:"id" bson:"_id"` // "A", "B", "C"
	Playlist       []MediaItem `json:"playlist" bson:"playlist"`
	CycleStartedAt time.Time   `json:"cycle_started_at" bson:"cycle_started_at"`
}

type PausedSnapshot struct {
	ItemIndex int     `json:"item_index"`
	OffsetSec float64 `json:"offset_sec"`
}

type WindowPlaybackState struct {
	CurrentItem MediaItem `json:"current_item"`
	OffsetSec   float64   `json:"offset_sec"`
}

type SyncStateResponse struct {
	SyncActive bool                           `json:"sync_active"`
	SyncItem   *MediaItem                     `json:"sync_item,omitempty"`
	SyncEndsAt *time.Time                     `json:"sync_ends_at,omitempty"`
	Windows    map[string]WindowPlaybackState `json:"windows"`
}
