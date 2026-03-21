package simulation

import (
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/types"
)

// TickRecord holds the log of a single tick's execution.
type TickRecord struct {
	// Tick is the tick number when this record was produced.
	Tick types.Tick
	// Commands lists the event commands that were executed during this tick.
	Commands []scenario.EventCommand
	// Events holds human-readable descriptions of notable occurrences.
	Events []string
}

// SimulationEngine drives the main game loop. Each call to Step executes
// one tick of the simulation, updating all subsystems in a fixed order.
type SimulationEngine struct {
	// State is the single source of truth for the running game.
	State *GameState
	// Executor applies event commands to the game state.
	Executor *CommandExecutor
	// TickLog records what happened on each tick for replay and debugging.
	TickLog []TickRecord
}
