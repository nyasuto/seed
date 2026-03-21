package simulation

import (
	"encoding/json"
	"fmt"

	"github.com/ponpoko/chaosseed-core/economy"
	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// Checkpoint captures a point-in-time snapshot of the entire simulation state.
// It holds serialized subsystem data and scalar fields, allowing the simulation
// to be restored to an earlier state for replay or undo functionality.
type Checkpoint struct {
	// Serialized subsystem states.
	CaveData     []byte
	ChiFlowData  []byte
	BeastData    []byte
	EconomyData  []byte
	InvasionData []byte
	ProgressData []byte

	// RNG state for deterministic continuation.
	RNGState types.RNGState

	// Scalar game state fields.
	NextBeastID             int
	NextWaveID              int
	ConsecutiveDeficitTicks int

	// DefeatResults stores pending beast defeat results keyed by beast ID.
	DefeatResults map[int]senju.DefeatResult

	// FiredEvents preserves the EventEngine's one-shot event tracking.
	FiredEvents map[string]bool
}

// CreateCheckpoint serializes the current simulation state into a Checkpoint.
// The engine's RNG must implement types.CheckpointableRNG; otherwise an error
// is returned.
func CreateCheckpoint(engine *SimulationEngine) (*Checkpoint, error) {
	s := engine.State

	crng, ok := s.RNG.(types.CheckpointableRNG)
	if !ok {
		return nil, fmt.Errorf("RNG does not support checkpointing; use types.NewCheckpointableRNG")
	}

	caveData, err := json.Marshal(s.Cave)
	if err != nil {
		return nil, fmt.Errorf("marshalling cave: %w", err)
	}

	chiFlowData, err := json.Marshal(s.ChiFlowEngine)
	if err != nil {
		return nil, fmt.Errorf("marshalling chi flow engine: %w", err)
	}

	beastData, err := senju.MarshalBeasts(s.Beasts)
	if err != nil {
		return nil, fmt.Errorf("marshalling beasts: %w", err)
	}

	economyData, err := economy.MarshalEconomyState(s.EconomyEngine)
	if err != nil {
		return nil, fmt.Errorf("marshalling economy: %w", err)
	}

	invasionData, err := invasion.MarshalInvasionState(s.Waves)
	if err != nil {
		return nil, fmt.Errorf("marshalling invasion state: %w", err)
	}

	progressData, err := scenario.MarshalProgress(s.Progress)
	if err != nil {
		return nil, fmt.Errorf("marshalling progress: %w", err)
	}

	// Deep copy DefeatResults.
	var defeatResults map[int]senju.DefeatResult
	if len(s.DefeatResults) > 0 {
		defeatResults = make(map[int]senju.DefeatResult, len(s.DefeatResults))
		for k, v := range s.DefeatResults {
			defeatResults[k] = v
		}
	}

	// Deep copy FiredEvents from EventEngine.
	var firedEvents map[string]bool
	if len(s.EventEngine.FiredEvents) > 0 {
		firedEvents = make(map[string]bool, len(s.EventEngine.FiredEvents))
		for k, v := range s.EventEngine.FiredEvents {
			firedEvents[k] = v
		}
	}

	return &Checkpoint{
		CaveData:                caveData,
		ChiFlowData:            chiFlowData,
		BeastData:               beastData,
		EconomyData:             economyData,
		InvasionData:            invasionData,
		ProgressData:            progressData,
		RNGState:                crng.RNGState(),
		NextBeastID:             s.NextBeastID,
		NextWaveID:              s.NextWaveID,
		ConsecutiveDeficitTicks: s.ConsecutiveDeficitTicks,
		DefeatResults:           defeatResults,
		FiredEvents:             firedEvents,
	}, nil
}

// RestoreCheckpoint reconstructs a SimulationEngine from a previously created
// Checkpoint. The scenario is required because it is immutable and not
// serialized into the checkpoint. Registries are reloaded from embedded data.
func RestoreCheckpoint(cp *Checkpoint, sc *scenario.Scenario) (*SimulationEngine, error) {
	// Load registries (immutable, loaded from embedded JSON).
	roomReg, err := world.LoadDefaultRoomTypes()
	if err != nil {
		return nil, fmt.Errorf("loading room types: %w", err)
	}
	speciesReg, err := senju.LoadDefaultSpecies()
	if err != nil {
		return nil, fmt.Errorf("loading species: %w", err)
	}
	evoReg, err := senju.LoadDefaultEvolution()
	if err != nil {
		return nil, fmt.Errorf("loading evolution: %w", err)
	}
	invaderClassReg, err := invasion.LoadDefaultInvaderClasses()
	if err != nil {
		return nil, fmt.Errorf("loading invader classes: %w", err)
	}

	// Restore cave.
	cave, err := world.UnmarshalCave(cp.CaveData)
	if err != nil {
		return nil, fmt.Errorf("restoring cave: %w", err)
	}

	// Restore chi flow engine.
	chiFlowEngine, err := fengshui.UnmarshalChiFlowEngine(
		cp.ChiFlowData, cave, roomReg, fengshui.DefaultFlowParams(),
	)
	if err != nil {
		return nil, fmt.Errorf("restoring chi flow engine: %w", err)
	}

	// Restore beasts.
	beasts, err := senju.UnmarshalBeasts(cp.BeastData, speciesReg)
	if err != nil {
		return nil, fmt.Errorf("restoring beasts: %w", err)
	}

	// Restore economy engine.
	economyEngine, err := economy.UnmarshalEconomyState(
		cp.EconomyData,
		economy.DefaultSupplyParams(),
		economy.DefaultCostParams(),
		economy.DefaultDeficitParams(),
		economy.DefaultConstructionCost(),
		economy.DefaultBeastCost(),
	)
	if err != nil {
		return nil, fmt.Errorf("restoring economy: %w", err)
	}

	// Restore invasion waves.
	waves, err := invasion.UnmarshalInvasionState(cp.InvasionData, invaderClassReg)
	if err != nil {
		return nil, fmt.Errorf("restoring invasion state: %w", err)
	}

	// Restore scenario progress.
	progress, err := scenario.UnmarshalProgress(cp.ProgressData)
	if err != nil {
		return nil, fmt.Errorf("restoring progress: %w", err)
	}

	// Restore RNG.
	rng := types.RestoreRNG(cp.RNGState)

	// Rebuild adjacency graph from restored cave.
	adjacencyGraph := cave.BuildAdjacencyGraph()

	// Reconstruct stateless engines.
	growthEngine := senju.NewGrowthEngine(senju.DefaultGrowthParams(), speciesReg)
	behaviorEngine := senju.NewBehaviorEngine(cave, adjacencyGraph, roomReg, nil)
	defeatProcessor := senju.NewDefeatProcessor()

	// Reconstruct invasion engine with restored RNG.
	invasionEngine := invasion.NewInvasionEngine(
		cave,
		adjacencyGraph,
		invasion.DefaultCombatParams(),
		rng,
		invaderClassReg,
		nil,
	)

	// Restore EventEngine with fired events.
	eventEngine := scenario.NewEventEngine()
	if cp.FiredEvents != nil {
		for k, v := range cp.FiredEvents {
			eventEngine.FiredEvents[k] = v
		}
	}

	// Restore DefeatResults.
	var defeatResults map[int]senju.DefeatResult
	if cp.DefeatResults != nil {
		defeatResults = make(map[int]senju.DefeatResult, len(cp.DefeatResults))
		for k, v := range cp.DefeatResults {
			defeatResults[k] = v
		}
	}

	state := &GameState{
		Cave:                    cave,
		RoomTypeRegistry:        roomReg,
		ChiFlowEngine:           chiFlowEngine,
		Beasts:                  beasts,
		GrowthEngine:            growthEngine,
		BehaviorEngine:          behaviorEngine,
		DefeatProcessor:         defeatProcessor,
		SpeciesRegistry:         speciesReg,
		EvolutionRegistry:       evoReg,
		InvasionEngine:          invasionEngine,
		Waves:                   waves,
		InvaderClassRegistry:    invaderClassReg,
		EconomyEngine:           economyEngine,
		Scenario:                sc,
		Progress:                progress,
		EventEngine:             eventEngine,
		RNG:                     rng,
		NextBeastID:             cp.NextBeastID,
		NextWaveID:              cp.NextWaveID,
		ScoreParams:             fengshui.DefaultScoreParams(),
		ConsecutiveDeficitTicks: cp.ConsecutiveDeficitTicks,
		DefeatResults:           defeatResults,
	}

	return &SimulationEngine{
		State:    state,
		Executor: NewCommandExecutor(),
		TickLog:  nil,
	}, nil
}
