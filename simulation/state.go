package simulation

import (
	"github.com/ponpoko/chaosseed-core/economy"
	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
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
}
