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

// Run executes the simulation loop up to maxTicks ticks. Each tick, the
// actionProvider is called with a snapshot of the current game state to obtain
// player actions. If actionProvider is nil, no actions are submitted each tick.
// The loop ends early if a win or loss condition is met. If maxTicks is reached
// without a terminal condition, the game is considered lost due to timeout.
func (e *SimulationEngine) Run(maxTicks int, actionProvider func(scenario.GameSnapshot) []PlayerAction) (GameResult, error) {
	for range maxTicks {
		var actions []PlayerAction
		if actionProvider != nil {
			snapshot := BuildSnapshot(e.State)
			actions = actionProvider(snapshot)
		}
		if actions == nil {
			actions = []PlayerAction{NoAction{}}
		}

		result, err := e.Step(actions)
		if err != nil {
			return GameResult{}, err
		}
		if result.Status != Running {
			return result, nil
		}
	}

	return GameResult{
		Status:    Lost,
		FinalTick: e.State.Progress.CurrentTick,
		Reason:    "max ticks reached",
	}, nil
}

// Step executes one tick of the simulation. It processes player actions, then
// runs all subsystem engines in a fixed order, evaluates win/loss conditions,
// and records the tick log. The returned GameResult indicates whether the game
// is still running, won, or lost.
func (e *SimulationEngine) Step(actions []PlayerAction) (GameResult, error) {
	s := e.State
	tick := s.Progress.CurrentTick
	var tickEvents []string

	// Record actions for replay.
	if e.RecordedActions != nil {
		e.RecordedActions[tick] = actions
	}

	// 1. Validate and execute player actions.
	actionEvents, err := e.processActions(actions, tick)
	if err != nil {
		return GameResult{}, err
	}
	tickEvents = append(tickEvents, actionEvents...)

	// 2. Chi flow: supply, propagation, decay.
	s.ChiFlowEngine.Tick()

	// 3-5. Beast subsystems: growth, revival, evolution.
	tickEvents = append(tickEvents, e.advanceBeastSystems(tick)...)

	// 6. Beast behavior AI.
	invaderPositions := buildInvaderPositions(s.Waves)
	s.BehaviorEngine.Tick(s.Beasts, invaderPositions, s.ChiFlowEngine.RoomChi)

	// 7-8. Invasion tick and economy processing.
	tickEvents = append(tickEvents, e.advanceInvasion(tick)...)

	// 9. Economy tick: supply, maintenance, deficit processing.
	tickEvents = append(tickEvents, e.advanceEconomy(tick)...)

	// 10. Event engine tick → command executor.
	eventEvents, cmds, err := e.advanceEvents(tick)
	if err != nil {
		return GameResult{}, err
	}
	tickEvents = append(tickEvents, eventEvents...)

	// 11. Record tick log.
	e.TickLog = append(e.TickLog, TickRecord{
		Tick:     tick,
		Commands: cmds,
		Events:   tickEvents,
	})

	// Advance tick counter before condition evaluation so that
	// survive_until(N) succeeds after processing the Nth tick.
	s.Progress.CurrentTick++

	// 12. Victory/defeat condition evaluation.
	result := evaluateEndConditions(s)
	if result.Status != Running {
		result.FinalTick = tick
		return result, nil
	}

	return GameResult{Status: Running}, nil
}

// processActions validates and executes player actions, returning event descriptions.
func (e *SimulationEngine) processActions(actions []PlayerAction, tick types.Tick) ([]string, error) {
	var events []string
	for _, action := range actions {
		result, err := ApplyAction(action, e.State)
		if err != nil {
			return nil, fmt.Errorf("tick %d action %s: %w", tick, action.ActionType(), err)
		}
		if result.Success && result.Description != "" {
			events = append(events, result.Description)
		}
	}
	return events, nil
}

// advanceBeastSystems runs beast growth, stunned revival, and evolution checks.
func (e *SimulationEngine) advanceBeastSystems(tick types.Tick) []string {
	s := e.State
	var events []string

	// Growth from chi absorption.
	roomMap := buildRoomMap(s.Cave.Rooms)
	growthEvents := s.GrowthEngine.Tick(s.Beasts, s.ChiFlowEngine.RoomChi, roomMap)
	for _, ge := range growthEvents {
		if ge.Type == senju.LevelUp {
			events = append(events, fmt.Sprintf("beast %d leveled up: %d→%d", ge.BeastID, ge.OldLevel, ge.NewLevel))
		}
	}

	// Stunned beast revival check.
	if s.DefeatResults == nil {
		s.DefeatResults = make(map[int]senju.DefeatResult)
	}
	for _, beast := range s.Beasts {
		if beast.State != senju.Stunned {
			continue
		}
		dr, ok := s.DefeatResults[beast.ID]
		if !ok {
			continue
		}
		if tick >= dr.RevivalTick {
			beast.State = senju.Idle
			beast.HP = max(dr.RevivalHP, 1)
			delete(s.DefeatResults, beast.ID)
			// Re-assign Guard behavior after revival.
			if s.BehaviorEngine != nil {
				s.BehaviorEngine.AssignBehavior(beast, senju.Guard)
			}
			events = append(events, fmt.Sprintf("beast %d revived", beast.ID))
		}
	}

	// Evolution condition check and execution.
	if s.EvolutionRegistry != nil {
		chiBalance := s.EconomyEngine.ChiPool.Balance()
		for _, beast := range s.Beasts {
			if beast.State == senju.Stunned || beast.RoomID == 0 {
				continue
			}
			room := s.Cave.RoomByID(beast.RoomID)
			if room == nil {
				continue
			}
			rt, err := s.RoomTypeRegistry.Get(room.TypeID)
			if err != nil {
				continue
			}
			rc := s.ChiFlowEngine.RoomChi[beast.RoomID]
			var roomChiRatio float64
			if rc != nil && rc.Capacity > 0 {
				roomChiRatio = rc.Current / rc.Capacity
			}
			path := s.EvolutionRegistry.CheckEvolution(beast, rt.Element, roomChiRatio, chiBalance)
			if path != nil {
				oldSpecies := beast.SpeciesID
				if err := senju.Evolve(beast, path, s.SpeciesRegistry); err == nil {
					s.EvolutionCount++
					events = append(events, fmt.Sprintf("beast %d evolved: %s→%s", beast.ID, oldSpecies, beast.SpeciesID))
				}
			}
		}
	}

	return events
}

// advanceInvasion runs the invasion engine tick and processes invasion events.
func (e *SimulationEngine) advanceInvasion(tick types.Tick) []string {
	s := e.State
	var events []string

	invasionEvents := s.InvasionEngine.Tick(
		tick,
		s.Waves,
		s.Beasts,
		s.Cave.Rooms,
		s.RoomTypeRegistry,
		s.ChiFlowEngine.RoomChi,
	)

	// Process invasion events: damage tracking, core damage, beast defeat.
	for _, ie := range invasionEvents {
		if ie.Type == invasion.CombatOccurred && ie.Damage > 0 {
			s.TotalDamageDealt += ie.Damage
		}
		if ie.Type == invasion.InvaderDefeated && ie.Damage > 0 {
			s.TotalDamageDealt += ie.Damage
		}
		if ie.Type == invasion.BeastDefeated && ie.Damage > 0 {
			s.TotalDamageReceived += ie.Damage
		}
		if ie.Type == invasion.GoalAchievedEvent {
			coreDamage := findInvaderATK(s.Waves, ie.InvaderID)
			if coreDamage > 0 {
				s.TotalDamageReceived += coreDamage
				s.Progress.CoreHP -= coreDamage
				if s.Progress.CoreHP < 0 {
					s.Progress.CoreHP = 0
				}
				events = append(events, fmt.Sprintf("core damaged by invader %d: -%d HP", ie.InvaderID, coreDamage))
			}
		}
		if ie.Type == invasion.BeastDefeated {
			for _, beast := range s.Beasts {
				if beast.ID == ie.BeastID && beast.HP <= 0 {
					dr := s.DefeatProcessor.ProcessDefeat(beast, tick)
					s.DefeatResults[dr.BeastID] = dr
					// Clear behavior for stunned beast to prevent stale map entries.
					if s.BehaviorEngine != nil {
						s.BehaviorEngine.RemoveBehavior(beast.ID)
					}
					events = append(events, fmt.Sprintf("beast %d stunned", beast.ID))
				}
			}
		}
	}

	// Invasion economy: rewards and losses.
	invasionEconProc := economy.NewInvasionEconomyProcessor()
	econSummary := invasionEconProc.ProcessInvasionEvents(invasionEvents, s.EconomyEngine.ChiPool, tick)
	if econSummary.NetChi != 0 {
		events = append(events, fmt.Sprintf("invasion economy: net chi %.1f", econSummary.NetChi))
	}

	return events
}

// advanceEconomy runs the economy engine tick and tracks deficit/peak chi.
func (e *SimulationEngine) advanceEconomy(tick types.Tick) []string {
	s := e.State
	var events []string

	veins := derefVeins(s.ChiFlowEngine.Veins)
	roomValues := derefRooms(s.Cave.Rooms)
	caveScore := 0.0
	if s.ScoreParams != nil {
		ev := fengshui.NewEvaluator(s.Cave, s.RoomTypeRegistry, s.ScoreParams)
		caveScore = ev.CaveTotal(s.ChiFlowEngine)
	}
	econResult := s.EconomyEngine.Tick(
		tick,
		veins,
		s.ChiFlowEngine.RoomChi,
		caveScore,
		roomValues,
		len(s.Beasts),
		0, // trapCount — not yet tracked
	)
	if econResult.DeficitResult.Shortage > 0 {
		s.ConsecutiveDeficitTicks++
		s.TotalDeficitTicks++
		events = append(events, fmt.Sprintf("deficit: shortage %.1f (consecutive: %d)", econResult.DeficitResult.Shortage, s.ConsecutiveDeficitTicks))
	} else {
		s.ConsecutiveDeficitTicks = 0
	}

	// Track peak chi pool balance.
	if chiBalance := s.EconomyEngine.ChiPool.Balance(); chiBalance > s.PeakChi {
		s.PeakChi = chiBalance
	}

	return events
}

// advanceEvents runs the event engine and applies resulting commands.
func (e *SimulationEngine) advanceEvents(tick types.Tick) ([]string, []scenario.EventCommand, error) {
	s := e.State
	snapshot := BuildSnapshot(s)
	cmds, err := s.EventEngine.Tick(snapshot, s.Scenario.Events)
	if err != nil {
		return nil, nil, fmt.Errorf("tick %d event engine: %w", tick, err)
	}
	var events []string
	if len(cmds) > 0 {
		if err := e.Executor.Apply(s, cmds); err != nil {
			return nil, nil, fmt.Errorf("tick %d command executor: %w", tick, err)
		}
		events = append(events, e.Executor.Messages...)
	}
	return events, cmds, nil
}

// buildRoomMap creates a map from room ID to room pointer for GrowthEngine.
func buildRoomMap(rooms []*world.Room) map[int]*world.Room {
	m := make(map[int]*world.Room, len(rooms))
	for _, r := range rooms {
		m[r.ID] = r
	}
	return m
}

// buildInvaderPositions creates a map from room ID to invader IDs for BehaviorEngine.
func buildInvaderPositions(waves []*invasion.InvasionWave) map[int][]int {
	m := make(map[int][]int)
	for _, w := range waves {
		if w.State != invasion.Active {
			continue
		}
		for _, inv := range w.Invaders {
			if inv.State == invasion.Defeated {
				continue
			}
			m[inv.CurrentRoomID] = append(m[inv.CurrentRoomID], inv.ID)
		}
	}
	return m
}

// findInvaderATK looks up an invader's ATK value from active waves.
func findInvaderATK(waves []*invasion.InvasionWave, invaderID int) int {
	for _, w := range waves {
		for _, inv := range w.Invaders {
			if inv.ID == invaderID {
				return inv.ATK
			}
		}
	}
	return 0
}

// derefVeins converts a slice of DragonVein pointers to values for EconomyEngine.
func derefVeins(veins []*fengshui.DragonVein) []fengshui.DragonVein {
	result := make([]fengshui.DragonVein, len(veins))
	for i, v := range veins {
		result[i] = *v
	}
	return result
}

// derefRooms converts a slice of Room pointers to values for EconomyEngine.
func derefRooms(rooms []*world.Room) []world.Room {
	result := make([]world.Room, len(rooms))
	for i, r := range rooms {
		result[i] = *r
	}
	return result
}

// evaluateEndConditions checks win and lose conditions against the current state.
func evaluateEndConditions(s *GameState) GameResult {
	snapshot := BuildSnapshot(s)

	// Check lose conditions first (they take priority).
	for _, condDef := range s.Scenario.LoseConditions {
		cond, err := scenario.NewCondition(condDef)
		if err != nil {
			continue
		}
		if cond.Evaluate(snapshot) {
			return GameResult{
				Status: Lost,
				Reason: fmt.Sprintf("lose condition met: %s", condDef.Type),
			}
		}
	}

	// Check win conditions.
	for _, condDef := range s.Scenario.WinConditions {
		cond, err := scenario.NewCondition(condDef)
		if err != nil {
			continue
		}
		if cond.Evaluate(snapshot) {
			return GameResult{
				Status: Won,
				Reason: fmt.Sprintf("win condition met: %s", condDef.Type),
			}
		}
	}

	return GameResult{Status: Running}
}

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
	// RecordedActions stores the player actions submitted on each tick for replay.
	RecordedActions map[types.Tick][]PlayerAction
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
	supplyParams, err := economy.DefaultSupplyParams()
	if err != nil {
		return nil, fmt.Errorf("loading supply params: %w", err)
	}
	costParams, err := economy.DefaultCostParams()
	if err != nil {
		return nil, fmt.Errorf("loading cost params: %w", err)
	}
	deficitParams, err := economy.DefaultDeficitParams()
	if err != nil {
		return nil, fmt.Errorf("loading deficit params: %w", err)
	}
	constructionCost, err := economy.DefaultConstructionCost()
	if err != nil {
		return nil, fmt.Errorf("loading construction cost: %w", err)
	}
	beastCost, err := economy.DefaultBeastCost()
	if err != nil {
		return nil, fmt.Errorf("loading beast cost: %w", err)
	}
	economyEngine := economy.NewEconomyEngine(
		chiPool,
		supplyParams,
		costParams,
		deficitParams,
		constructionCost,
		beastCost,
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

	// Count scheduled spawn_wave events so defeat_all_waves knows
	// how many waves to expect before declaring victory.
	scheduledWaves := 0
	for _, ev := range sc.Events {
		for _, cmd := range ev.Commands {
			if cmd.Type == "spawn_wave" {
				scheduledWaves++
			}
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
		Waves:                   nil,
		InvaderClassRegistry:    invaderClassReg,
		EconomyEngine:           economyEngine,
		Scenario:                sc,
		Progress:                progress,
		EventEngine:             eventEngine,
		RNG:                     rng,
		NextBeastID:             nextBeastID,
		NextWaveID:              1,
		ScoreParams:             fengshui.DefaultScoreParams(),
		ConsecutiveDeficitTicks: 0,
		ScheduledWaves:          scheduledWaves,
	}

	return &SimulationEngine{
		State:    state,
		Executor: NewCommandExecutor(),
		TickLog:  nil,
	}, nil
}
