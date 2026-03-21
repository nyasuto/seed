package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nyasuto/seed/core/simulation"
)

// SaveCheckpoint creates a checkpoint of the engine's current state and writes
// it to the file at path. The engine's RNG must implement CheckpointableRNG.
func SaveCheckpoint(path string, engine *simulation.SimulationEngine) error {
	cp, err := simulation.CreateCheckpoint(engine)
	if err != nil {
		return fmt.Errorf("create checkpoint: %w", err)
	}
	data, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("marshal checkpoint: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write checkpoint file: %w", err)
	}
	return nil
}

// LoadCheckpoint reads a checkpoint from the file at path.
func LoadCheckpoint(path string) (*simulation.Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read checkpoint file: %w", err)
	}
	var cp simulation.Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("unmarshal checkpoint: %w", err)
	}
	return &cp, nil
}

// SaveCheckpointTo saves the current game engine state to the file at path.
// This method is only valid while a game is running (during RunGame or
// ResumeGame). It returns an error if no engine is active.
func (gs *GameServer) SaveCheckpointTo(path string) error {
	if gs.engine == nil {
		return fmt.Errorf("no active game session")
	}
	return SaveCheckpoint(path, gs.engine)
}

// LoadCheckpointFrom reads a checkpoint file and restores the engine.
// After calling this method, use ResumeGame to continue the game.
func (gs *GameServer) LoadCheckpointFrom(path string) error {
	cp, err := LoadCheckpoint(path)
	if err != nil {
		return err
	}
	engine, err := simulation.RestoreCheckpoint(cp, gs.scenario)
	if err != nil {
		return fmt.Errorf("restore checkpoint: %w", err)
	}
	gs.engine = engine
	return nil
}
