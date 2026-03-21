package invasion

import (
	"github.com/nyasuto/seed/core/types"
)

// WaveState represents the lifecycle state of an invasion wave.
type WaveState int

const (
	// Pending means the wave has not yet been triggered.
	Pending WaveState = iota
	// Active means the wave is currently in progress.
	Active
	// Completed means all invaders were defeated (defense success).
	Completed
	// Failed means invaders achieved their goals or escaped (defense failure).
	Failed
)

// String returns the name of the wave state.
func (s WaveState) String() string {
	switch s {
	case Pending:
		return "Pending"
	case Active:
		return "Active"
	case Completed:
		return "Completed"
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// InvasionWave represents a group of invaders that attack together.
type InvasionWave struct {
	// ID is the unique identifier for this wave.
	ID int
	// TriggerTick is the tick at which this wave becomes active.
	TriggerTick types.Tick
	// Invaders is the list of invaders in this wave.
	Invaders []*Invader
	// State is the current lifecycle state of the wave.
	State WaveState
	// Difficulty is the difficulty multiplier for this wave.
	Difficulty float64
}

// IsActive returns true if the wave is currently in progress.
func (w *InvasionWave) IsActive() bool {
	return w.State == Active
}

// IsCompleted returns true if the wave has finished (either Completed or Failed).
func (w *InvasionWave) IsCompleted() bool {
	return w.State == Completed || w.State == Failed
}

// AliveCount returns the number of invaders that are not yet defeated.
func (w *InvasionWave) AliveCount() int {
	count := 0
	for _, inv := range w.Invaders {
		if inv.State != Defeated {
			count++
		}
	}
	return count
}

// DefeatedCount returns the number of invaders that have been defeated.
func (w *InvasionWave) DefeatedCount() int {
	count := 0
	for _, inv := range w.Invaders {
		if inv.State == Defeated {
			count++
		}
	}
	return count
}
