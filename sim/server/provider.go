package server

import (
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
)

// ActionProvider is the interface that adapters (human, AI, batch) implement
// to supply player actions each tick and receive game lifecycle callbacks.
type ActionProvider interface {
	// ProvideActions is called each tick with the current game snapshot.
	// The provider returns the actions to execute this tick.
	ProvideActions(snapshot scenario.GameSnapshot) ([]simulation.PlayerAction, error)

	// OnTickComplete is called after each tick with the post-tick snapshot.
	OnTickComplete(snapshot scenario.GameSnapshot)

	// OnGameEnd is called once when the game reaches a terminal state.
	OnGameEnd(result simulation.RunResult)
}
