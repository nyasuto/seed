package scenario

import (
	"encoding/json"

	"github.com/nyasuto/seed/core/types"
)

// GameSnapshot is a read-only snapshot of the current game state,
// passed to ConditionEvaluator.Evaluate each tick to determine
// whether a win or lose condition has been met.
type GameSnapshot struct {
	Tick                   types.Tick
	CoreHP                 int
	ChiPoolBalance         float64
	BeastCount             int
	AliveBeasts            int
	DefeatedWaves          int
	TotalWaves             int
	CaveFengShuiScore      float64
	ConsecutiveDeficitTicks int
}

// ConditionEvaluator evaluates a game condition against a snapshot.
// Each concrete condition type implements this interface.
type ConditionEvaluator interface {
	Evaluate(snapshot GameSnapshot) bool
}

// ConditionDef defines a win or lose condition in data-driven form.
// Type identifies the kind of condition (e.g. "survive_until",
// "defeat_all_waves", "core_destroyed"), and Params holds type-specific
// parameters as key-value pairs loaded from JSON scenario data.
type ConditionDef struct {
	// Type is the condition identifier used by the factory to instantiate
	// the corresponding ConditionEvaluator.
	Type string
	// Params holds condition-specific parameters as raw JSON. For example a
	// "survive_until" condition might contain {"ticks": 3000}.
	Params json.RawMessage
}
