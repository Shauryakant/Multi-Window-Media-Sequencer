package services

import (
	"math"
	"time"

	"media-sequencer/models"
)

const DefaultVideoDurationSec = 15.0

// GetEffectiveDuration returns the item duration in seconds.
// For videos with duration_sec <= 0, it uses DefaultVideoDurationSec for cycle math.
func GetEffectiveDuration(item models.MediaItem) float64 {
	if item.DurationSec > 0 {
		return item.DurationSec
	}
	if item.Type == "video" {
		return DefaultVideoDurationSec
	}
	// Default fallback for any zero duration item
	return 5.0
}

// CalculateTotalPlaylistDuration returns the total duration in seconds of a window's playlist.
func CalculateTotalPlaylistDuration(playlist []models.MediaItem) float64 {
	var total float64
	for _, item := range playlist {
		total += GetEffectiveDuration(item)
	}
	return total
}

// CalculateCurrentState computes which item is playing and the offset into that item based on cycle_started_at.
func CalculateCurrentState(playlist []models.MediaItem, cycleStartedAt time.Time, now time.Time) (models.MediaItem, models.PausedSnapshot) {
	if len(playlist) == 0 {
		return models.MediaItem{}, models.PausedSnapshot{ItemIndex: 0, OffsetSec: 0}
	}

	totalDuration := CalculateTotalPlaylistDuration(playlist)
	if totalDuration <= 0 {
		return playlist[0], models.PausedSnapshot{ItemIndex: 0, OffsetSec: 0}
	}

	elapsedSec := now.Sub(cycleStartedAt).Seconds()
	if elapsedSec < 0 {
		elapsedSec = 0
	}

	cycleOffset := math.Mod(elapsedSec, totalDuration)
	if cycleOffset < 0 {
		cycleOffset += totalDuration
	}

	accumulated := 0.0
	for i, item := range playlist {
		dur := GetEffectiveDuration(item)
		if cycleOffset < accumulated+dur || i == len(playlist)-1 {
			offsetInItem := cycleOffset - accumulated
			if offsetInItem < 0 {
				offsetInItem = 0
			}
			return item, models.PausedSnapshot{
				ItemIndex: i,
				OffsetSec:  offsetInItem,
			}
		}
		accumulated += dur
	}

	return playlist[0], models.PausedSnapshot{ItemIndex: 0, OffsetSec: 0}
}

// ComputeNewCycleStartedAt calculates a new cycle_started_at time given a target item index and offset,
// such that resuming normal cycle math at time `now` lands exactly on {targetIndex, targetOffsetSec}.
func ComputeNewCycleStartedAt(playlist []models.MediaItem, targetIndex int, targetOffsetSec float64, now time.Time) time.Time {
	if len(playlist) == 0 {
		return now
	}

	totalDuration := CalculateTotalPlaylistDuration(playlist)
	if totalDuration <= 0 {
		return now
	}

	if targetIndex < 0 || targetIndex >= len(playlist) {
		targetIndex = 0
	}

	// Calculate cumulative duration up to targetIndex
	cumOffset := 0.0
	for i := 0; i < targetIndex; i++ {
		cumOffset += GetEffectiveDuration(playlist[i])
	}

	targetItemDur := GetEffectiveDuration(playlist[targetIndex])
	if targetOffsetSec < 0 {
		targetOffsetSec = 0
	}
	if targetOffsetSec > targetItemDur {
		targetOffsetSec = targetItemDur
	}

	offsetIntoLoop := cumOffset + targetOffsetSec
	// Wrap around if offset exceeds total duration
	offsetIntoLoop = math.Mod(offsetIntoLoop, totalDuration)

	// new_cycle_started_at = now - offsetIntoLoop
	return now.Add(-time.Duration(offsetIntoLoop * float64(time.Second)))
}
