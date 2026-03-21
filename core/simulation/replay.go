package simulation

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/types"
)

// Errors returned by replay functions.
var (
	// ErrRNGNotCheckpointable is returned when the engine's RNG does not
	// support state extraction for replay recording.
	ErrRNGNotCheckpointable = errors.New("rng does not support checkpointing")
)

// Replay captures the minimal information needed to deterministically
// reproduce a simulation run: the RNG seed, scenario ID, and every player
// action submitted on each tick.
type Replay struct {
	// Seed is the initial RNG seed used to create the simulation.
	Seed int64
	// ScenarioID identifies the scenario that was played.
	ScenarioID string
	// Actions maps each tick to the player actions submitted on that tick.
	Actions map[types.Tick][]PlayerAction
}

// RecordReplay extracts a Replay from a SimulationEngine that has been run
// with action recording enabled. The engine's RNG must implement
// CheckpointableRNG so the original seed can be recovered.
// Call EnableRecording on the engine before running the simulation.
func RecordReplay(engine *SimulationEngine) (*Replay, error) {
	crng, ok := engine.State.RNG.(types.CheckpointableRNG)
	if !ok {
		return nil, ErrRNGNotCheckpointable
	}

	actions := engine.RecordedActions
	if actions == nil {
		actions = make(map[types.Tick][]PlayerAction)
	}

	return &Replay{
		Seed:       crng.RNGState().Seed,
		ScenarioID: engine.State.Scenario.ID,
		Actions:    actions,
	}, nil
}

// EnableRecording initializes action recording on the engine. This must be
// called before Run or Step to capture player actions for replay.
func EnableRecording(engine *SimulationEngine) {
	if engine.RecordedActions == nil {
		engine.RecordedActions = make(map[types.Tick][]PlayerAction)
	}
}

// actionEnvelope wraps a PlayerAction for JSON serialization, preserving the
// concrete type via a "type" discriminator field.
type actionEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// replayJSON is the JSON-friendly representation of Replay. The Actions map
// uses string keys because JSON object keys must be strings.
type replayJSON struct {
	Seed       int64                         `json:"seed"`
	ScenarioID string                        `json:"scenario_id"`
	Actions    map[string][]actionEnvelope   `json:"actions"`
}

// MarshalReplay serializes a Replay to JSON bytes.
func MarshalReplay(r *Replay) ([]byte, error) {
	rj := replayJSON{
		Seed:       r.Seed,
		ScenarioID: r.ScenarioID,
		Actions:    make(map[string][]actionEnvelope, len(r.Actions)),
	}

	for tick, actions := range r.Actions {
		key := strconv.FormatUint(uint64(tick), 10)
		envelopes := make([]actionEnvelope, len(actions))
		for i, a := range actions {
			data, err := json.Marshal(a)
			if err != nil {
				return nil, fmt.Errorf("marshal action at tick %d: %w", tick, err)
			}
			envelopes[i] = actionEnvelope{
				Type: a.ActionType(),
				Data: data,
			}
		}
		rj.Actions[key] = envelopes
	}

	return json.Marshal(rj)
}

// UnmarshalReplay deserializes JSON bytes into a Replay.
func UnmarshalReplay(data []byte) (*Replay, error) {
	var rj replayJSON
	if err := json.Unmarshal(data, &rj); err != nil {
		return nil, fmt.Errorf("unmarshal replay: %w", err)
	}

	actions := make(map[types.Tick][]PlayerAction, len(rj.Actions))
	for key, envelopes := range rj.Actions {
		tick, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid tick key %q: %w", key, err)
		}

		pas := make([]PlayerAction, len(envelopes))
		for i, env := range envelopes {
			pa, err := unmarshalAction(env)
			if err != nil {
				return nil, fmt.Errorf("unmarshal action at tick %d: %w", tick, err)
			}
			pas[i] = pa
		}
		actions[types.Tick(tick)] = pas
	}

	return &Replay{
		Seed:       rj.Seed,
		ScenarioID: rj.ScenarioID,
		Actions:    actions,
	}, nil
}

func unmarshalAction(env actionEnvelope) (PlayerAction, error) {
	switch env.Type {
	case "dig_room":
		var a DigRoomAction
		if err := json.Unmarshal(env.Data, &a); err != nil {
			return nil, err
		}
		return a, nil
	case "dig_corridor":
		var a DigCorridorAction
		if err := json.Unmarshal(env.Data, &a); err != nil {
			return nil, err
		}
		return a, nil
	case "place_beast":
		var a PlaceBeastAction
		if err := json.Unmarshal(env.Data, &a); err != nil {
			return nil, err
		}
		return a, nil
	case "upgrade_room":
		var a UpgradeRoomAction
		if err := json.Unmarshal(env.Data, &a); err != nil {
			return nil, err
		}
		return a, nil
	case "summon_beast":
		var a SummonBeastAction
		if err := json.Unmarshal(env.Data, &a); err != nil {
			return nil, err
		}
		return a, nil
	case "evolve_beast":
		var a EvolveBeastAction
		if err := json.Unmarshal(env.Data, &a); err != nil {
			return nil, err
		}
		return a, nil
	case "no_action":
		return NoAction{}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownAction, env.Type)
	}
}

// PlayReplay recreates a simulation from a Replay and the corresponding
// scenario. It constructs a new engine with the same RNG seed and replays
// the recorded actions tick by tick. The result should be identical to the
// original run due to deterministic design.
func PlayReplay(replay *Replay, sc *scenario.Scenario) (GameResult, error) {
	rng := types.NewCheckpointableRNG(replay.Seed)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		return GameResult{}, err
	}

	// Determine the maximum tick to replay.
	var maxTick types.Tick
	for t := range replay.Actions {
		if t > maxTick {
			maxTick = t
		}
	}

	// Run the simulation providing recorded actions on each tick.
	result, err := engine.Run(int(maxTick)+1, func(snapshot scenario.GameSnapshot) []PlayerAction {
		if actions, ok := replay.Actions[snapshot.Tick]; ok {
			return actions
		}
		return []PlayerAction{NoAction{}}
	})
	if err != nil {
		return GameResult{}, err
	}

	return result, nil
}
