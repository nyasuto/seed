package scenario

import "github.com/nyasuto/seed/core/types"

// WaveResult records the outcome of a single invasion wave.
type WaveResult struct {
	// WaveID is the index of the wave in the scenario's WaveSchedule.
	WaveID int
	// Result describes the outcome (e.g. "victory", "defeat", "in_progress").
	Result string
	// CompletedTick is the tick at which the wave was resolved.
	CompletedTick types.Tick
}

// ScenarioProgress tracks the mutable progress state of a running scenario.
// It is separate from Scenario (which is immutable configuration) and
// GameSnapshot (which is a per-tick read-only view).
type ScenarioProgress struct {
	// ScenarioID identifies which scenario this progress belongs to.
	ScenarioID string
	// CurrentTick is the latest tick that has been processed.
	CurrentTick types.Tick
	// FiredEventIDs lists the IDs of one-shot events that have fired.
	FiredEventIDs []string
	// WaveResults records the outcome of each completed wave.
	WaveResults []WaveResult
	// CoreHP is the current hit points of the dungeon core.
	CoreHP int
}
