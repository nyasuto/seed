package simulation

import (
	"github.com/nyasuto/seed/core/economy"
	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// GameStatus represents the current state of a game session.
type GameStatus int

const (
	// Running indicates the game is still in progress.
	Running GameStatus = iota
	// Won indicates the player has achieved victory.
	Won
	// Lost indicates the player has been defeated.
	Lost
)

// String returns the string representation of a GameStatus.
func (s GameStatus) String() string {
	switch s {
	case Running:
		return "Running"
	case Won:
		return "Won"
	case Lost:
		return "Lost"
	default:
		return "Unknown"
	}
}

// GameResult captures the outcome of a completed (or still-running) game.
// When Status is Running, FinalTick and Reason are meaningless.
type GameResult struct {
	// Status is the current game status.
	Status GameStatus
	// FinalTick is the tick at which the game ended.
	FinalTick types.Tick
	// Reason describes why the game ended (e.g. "core HP reached 0").
	Reason string
}

// GameState holds every subsystem engine, the active scenario, progress
// tracker, economy state, and the deterministic RNG. It is the single
// source of truth for a running game session.
type GameState struct {
	// Cave is the dungeon map containing rooms, corridors, and the grid.
	Cave *world.Cave

	// RoomTypeRegistry holds all available room type definitions.
	RoomTypeRegistry *world.RoomTypeRegistry

	// ChiFlowEngine simulates chi flow through dragon veins and rooms.
	ChiFlowEngine *fengshui.ChiFlowEngine

	// Beasts is the list of all beasts currently in the dungeon.
	Beasts []*senju.Beast

	// GrowthEngine handles beast leveling from chi absorption.
	GrowthEngine *senju.GrowthEngine

	// BehaviorEngine drives beast AI decisions each tick.
	BehaviorEngine *senju.BehaviorEngine

	// DefeatProcessor handles beast defeat, stunning, and revival.
	DefeatProcessor *senju.DefeatProcessor

	// SpeciesRegistry holds all available beast species definitions.
	SpeciesRegistry *senju.SpeciesRegistry

	// EvolutionRegistry holds all beast evolution paths.
	EvolutionRegistry *senju.EvolutionRegistry

	// InvasionEngine orchestrates invader pathfinding, combat, and retreat.
	InvasionEngine *invasion.InvasionEngine

	// Waves is the list of all invasion waves (pending, active, and completed).
	Waves []*invasion.InvasionWave

	// InvaderClassRegistry holds all invader class definitions.
	InvaderClassRegistry *invasion.InvaderClassRegistry

	// EconomyEngine manages chi resource supply, maintenance, and costs.
	EconomyEngine *economy.EconomyEngine

	// Scenario is the immutable scenario configuration for this game.
	Scenario *scenario.Scenario

	// Progress tracks mutable scenario progress (tick, fired events, wave results, core HP).
	Progress *scenario.ScenarioProgress

	// EventEngine evaluates scripted event conditions each tick.
	EventEngine *scenario.EventEngine

	// RNG is the deterministic random number generator for this game session.
	RNG types.RNG

	// NextBeastID is the auto-incrementing ID counter for new beasts.
	NextBeastID int

	// NextWaveID is the auto-incrementing ID counter for new invasion waves.
	NextWaveID int

	// ScoreParams holds feng shui scoring parameters for cave evaluation.
	ScoreParams *fengshui.ScoreParams

	// ConsecutiveDeficitTicks tracks how many consecutive ticks the economy
	// has been in deficit. Reset to 0 when the tick has no deficit.
	ConsecutiveDeficitTicks int

	// DefeatResults stores pending beast defeat results keyed by beast ID.
	// Used to track revival tick and HP for stunned beasts.
	DefeatResults map[int]senju.DefeatResult

	// --- Accumulated statistics (updated each tick by Step) ---

	// PeakChi is the highest chi pool balance observed during the simulation.
	PeakChi float64

	// TotalDamageDealt is the cumulative damage dealt to invaders.
	TotalDamageDealt int

	// TotalDamageReceived is the cumulative damage dealt to beasts and the core.
	TotalDamageReceived int

	// EvolutionCount is the number of beast evolutions that occurred.
	EvolutionCount int

	// TotalDeficitTicks is the total number of ticks where the economy was in deficit.
	TotalDeficitTicks int

	// ScheduledWaves is the total number of waves scheduled in the scenario
	// (counted from spawn_wave event commands). Used by BuildSnapshot so that
	// defeat_all_waves does not trigger before all waves have been spawned.
	ScheduledWaves int
}
