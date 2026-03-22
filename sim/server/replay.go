package server

import (
	"fmt"
	"os"

	"github.com/nyasuto/seed/core/simulation"
)

// SaveReplay extracts a replay from the engine and writes it to the file at
// path. The engine must have recording enabled (see simulation.EnableRecording)
// and its RNG must implement CheckpointableRNG.
func SaveReplay(path string, engine *simulation.SimulationEngine) error {
	replay, err := simulation.RecordReplay(engine)
	if err != nil {
		return fmt.Errorf("record replay: %w", err)
	}
	data, err := simulation.MarshalReplay(replay)
	if err != nil {
		return fmt.Errorf("marshal replay: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write replay file: %w", err)
	}
	return nil
}

// LoadReplay reads a replay from the file at path.
func LoadReplay(path string) (*simulation.Replay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read replay file: %w", err)
	}
	replay, err := simulation.UnmarshalReplay(data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal replay: %w", err)
	}
	return replay, nil
}

// SaveReplayTo saves a replay of the current game to the file at path.
// This method is only valid while a game is running (during RunGame).
// It returns an error if no engine is active.
func (gs *GameServer) SaveReplayTo(path string) error {
	if gs.engine == nil {
		return fmt.Errorf("no active game session")
	}
	return SaveReplay(path, gs.engine)
}

// PlayReplayFrom loads a replay file and replays it using the server's
// scenario, returning the game result. The replay is deterministic: the
// same seed and actions produce the same result.
func (gs *GameServer) PlayReplayFrom(path string) (simulation.GameResult, error) {
	replay, err := LoadReplay(path)
	if err != nil {
		return simulation.GameResult{}, err
	}
	return simulation.PlayReplay(replay, gs.scenario)
}
