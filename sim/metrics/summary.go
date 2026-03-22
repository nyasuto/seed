package metrics

import "github.com/nyasuto/seed/core/simulation"

// GameSummary holds aggregate statistics for a completed game.
type GameSummary struct {
	// Result is the game outcome (Won/Lost).
	Result simulation.GameStatus
	// Reason describes why the game ended.
	Reason string
	// TotalTicks is the number of ticks the game ran.
	TotalTicks int
	// RoomsBuilt is the peak number of rooms observed during the game.
	RoomsBuilt int
	// FinalCoreHP is the core HP at game end.
	FinalCoreHP int
	// PeakChi is the highest chi pool balance observed.
	PeakChi float64
	// FinalFengShui is the feng shui score at game end.
	FinalFengShui float64
	// WavesDefeated is the number of invasion waves defeated.
	WavesDefeated int
	// TotalWaves is the total number of invasion waves in the scenario.
	TotalWaves int
	// PeakBeasts is the highest beast count observed during the game.
	PeakBeasts int
	// TotalDamageDealt is cumulative damage dealt to invaders.
	TotalDamageDealt int
	// TotalDamageReceived is cumulative damage received by beasts and core.
	TotalDamageReceived int
	// DeficitTicks is the total number of ticks with economy deficit.
	DeficitTicks int
	// Evolutions is the total number of beast evolutions.
	Evolutions int
}
