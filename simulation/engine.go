package simulation

import (
	"fmt"

	"github.com/ponpoko/chaosseed-core/economy"
	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
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

// NewSimulationEngine constructs a fully initialized SimulationEngine from the
// given scenario and RNG. It loads all default registries, creates the cave,
// applies terrain, places prebuilt rooms, builds dragon veins, places starting
// beasts, initializes the chi pool, and wires up every subsystem engine.
func NewSimulationEngine(sc *scenario.Scenario, rng types.RNG) (*SimulationEngine, error) {
	is := sc.InitialState

	// Load default registries.
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

	// Create cave.
	cave, err := world.NewCave(is.CaveWidth, is.CaveHeight)
	if err != nil {
		return nil, fmt.Errorf("creating cave: %w", err)
	}

	// Place prebuilt rooms.
	placedRooms := make([]*world.Room, 0, len(is.PrebuiltRooms))
	for i, rp := range is.PrebuiltRooms {
		rt, err := roomReg.Get(rp.TypeID)
		if err != nil {
			return nil, fmt.Errorf("prebuilt room[%d] type %q: %w", i, rp.TypeID, err)
		}
		level := rp.Level
		if level <= 0 {
			level = 1
		}
		room, err := cave.AddRoom(rp.TypeID, rp.Pos, 3, 3, []world.RoomEntrance{
			{Pos: types.Pos{X: rp.Pos.X + 1, Y: rp.Pos.Y + 2}, Dir: types.South},
		})
		if err != nil {
			return nil, fmt.Errorf("prebuilt room[%d] %q: %w", i, rp.TypeID, err)
		}
		room.Level = level
		if rt.BaseCoreHP > 0 {
			room.CoreHP = rt.CoreHPAtLevel(level)
		}
		placedRooms = append(placedRooms, room)
	}

	// Generate and apply terrain.
	terrainRNG := types.NewSeededRNG(is.TerrainSeed)
	tg := &scenario.TerrainGenerator{}
	zones := tg.GenerateTerrain(is.CaveWidth, is.CaveHeight, is.TerrainDensity, terrainRNG)
	// Filter zones that overlap existing rooms.
	var safeZones []scenario.TerrainZone
	for _, z := range zones {
		overlap := false
		for dy := 0; dy < z.Height && !overlap; dy++ {
			for dx := 0; dx < z.Width && !overlap; dx++ {
				pos := types.Pos{X: z.Pos.X + dx, Y: z.Pos.Y + dy}
				if cave.Grid.InBounds(pos) {
					cell, _ := cave.Grid.At(pos)
					if cell.RoomID != 0 {
						overlap = true
					}
				}
			}
		}
		if !overlap {
			safeZones = append(safeZones, z)
		}
	}
	if err := scenario.ApplyTerrain(cave, safeZones); err != nil {
		return nil, fmt.Errorf("applying terrain: %w", err)
	}

	// Build dragon veins.
	veins := make([]*fengshui.DragonVein, 0, len(is.DragonVeins))
	for i, dv := range is.DragonVeins {
		vein, err := fengshui.BuildDragonVein(cave, dv.SourcePos, dv.Element, dv.FlowRate)
		if err != nil {
			return nil, fmt.Errorf("dragon vein[%d]: %w", i, err)
		}
		vein.ID = i + 1
		veins = append(veins, vein)
	}

	// Create chi flow engine.
	chiFlowEngine := fengshui.NewChiFlowEngine(cave, veins, roomReg, fengshui.DefaultFlowParams())

	// Build adjacency graph.
	adjacencyGraph := cave.BuildAdjacencyGraph()

	// Create beast subsystems.
	growthEngine := senju.NewGrowthEngine(senju.DefaultGrowthParams(), speciesReg)
	behaviorEngine := senju.NewBehaviorEngine(cave, adjacencyGraph, roomReg, nil)
	defeatProcessor := senju.NewDefeatProcessor()

	// Place starting beasts.
	var beasts []*senju.Beast
	nextBeastID := 1
	for i, bp := range is.StartingBeasts {
		species, err := speciesReg.Get(bp.SpeciesID)
		if err != nil {
			return nil, fmt.Errorf("starting beast[%d] species %q: %w", i, bp.SpeciesID, err)
		}
		beast := senju.NewBeast(nextBeastID, species, 0)
		nextBeastID++
		if bp.RoomIndex >= 0 && bp.RoomIndex < len(placedRooms) {
			room := placedRooms[bp.RoomIndex]
			rt, _ := roomReg.Get(room.TypeID)
			if err := senju.PlaceBeast(beast, room, rt); err != nil {
				return nil, fmt.Errorf("starting beast[%d] placement: %w", i, err)
			}
		}
		beasts = append(beasts, beast)
	}

	// Create chi pool and deposit starting chi.
	chiPool := economy.NewChiPool(is.StartingChi * 10)
	if is.StartingChi > 0 {
		_ = chiPool.Deposit(is.StartingChi, economy.Supply, "initial chi", 0)
	}

	// Create economy engine.
	economyEngine := economy.NewEconomyEngine(
		chiPool,
		economy.DefaultSupplyParams(),
		economy.DefaultCostParams(),
		economy.DefaultDeficitParams(),
		economy.DefaultConstructionCost(),
		economy.DefaultBeastCost(),
	)

	// Create invasion engine.
	invasionEngine := invasion.NewInvasionEngine(
		cave,
		adjacencyGraph,
		invasion.DefaultCombatParams(),
		rng,
		invaderClassReg,
		nil,
	)

	// Initialize scenario progress.
	coreHP := 0
	for _, room := range placedRooms {
		if room.CoreHP > 0 {
			coreHP += room.CoreHP
		}
	}
	progress := &scenario.ScenarioProgress{
		ScenarioID: sc.ID,
		CoreHP:     coreHP,
	}

	// Create event engine.
	eventEngine := scenario.NewEventEngine()

	state := &GameState{
		Cave:                   cave,
		RoomTypeRegistry:       roomReg,
		ChiFlowEngine:          chiFlowEngine,
		Beasts:                 beasts,
		GrowthEngine:           growthEngine,
		BehaviorEngine:         behaviorEngine,
		DefeatProcessor:        defeatProcessor,
		SpeciesRegistry:        speciesReg,
		EvolutionRegistry:      evoReg,
		InvasionEngine:         invasionEngine,
		Waves:                  nil,
		InvaderClassRegistry:   invaderClassReg,
		EconomyEngine:          economyEngine,
		Scenario:               sc,
		Progress:               progress,
		EventEngine:            eventEngine,
		RNG:                    rng,
		NextBeastID:            nextBeastID,
		NextWaveID:             1,
		ScoreParams:            fengshui.DefaultScoreParams(),
		ConsecutiveDeficitTicks: 0,
	}

	return &SimulationEngine{
		State:    state,
		Executor: NewCommandExecutor(),
		TickLog:  nil,
	}, nil
}
