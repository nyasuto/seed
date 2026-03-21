// Command caveviz generates a hardcoded Cave and prints its ASCII representation.
// Use --chi to display a chi flow overlay with dragon veins.
// Use --beasts to display beast placements.
// Use --invasion to display invader overlay.
// Use --battle to display all layers (terrain + chi + beasts + invaders).
// Use --economy to display economy status overlay.
// Use --scenario <file> to load a scenario JSON file and display its status.
// Use --all to display all layers (standard + chi + beasts + ai).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ponpoko/chaosseed-core/economy"
	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

func main() {
	chiMode := flag.Bool("chi", false, "display chi flow overlay")
	beastMode := flag.Bool("beasts", false, "display beast placement overlay")
	aiMode := flag.Bool("ai", false, "display beast behavior state overlay")
	invasionMode := flag.Bool("invasion", false, "display invasion overlay")
	economyMode := flag.Bool("economy", false, "display economy status overlay")
	scenarioFile := flag.String("scenario", "", "load scenario JSON file and display status")
	battleMode := flag.Bool("battle", false, "display all layers (terrain + chi + beasts + invaders)")
	allMode := flag.Bool("all", false, "display all layers (standard + chi + beasts + ai + economy)")
	flag.Parse()

	cave, err := buildDemoCave()
	if err != nil {
		fmt.Printf("error building demo cave: %v\n", err)
		return
	}

	if *scenarioFile != "" {
		displayScenarioStatus(*scenarioFile)
		return
	}

	showChi := *chiMode || *allMode || *battleMode
	showBeasts := *beastMode || *allMode || *battleMode
	showAI := *aiMode || *allMode
	showInvasion := *invasionMode || *battleMode
	showEconomy := *economyMode || *allMode

	if showChi {
		engine, err := buildDemoEngine(cave)
		if err != nil {
			fmt.Printf("error building chi engine: %v\n", err)
			return
		}
		for i := 0; i < 10; i++ {
			engine.Tick()
		}
		fmt.Print(fengshui.RenderChiOverlay(cave, engine))
		fmt.Println()
		fmt.Println("Legend: ██=Rock/Full  ~~=DragonVein  ..=Corridor  ><=Entrance")
		fmt.Println("       __=Empty  ░░=Low  ▒▒=Mid  ▓▓=High  ██=Full")
	}

	if showBeasts {
		beasts := buildDemoBeasts()
		if showChi {
			fmt.Println()
		}
		fmt.Print(senju.RenderBeastOverlay(cave, beasts))
		fmt.Println()
		fmt.Println("Legend: ██=Rock  ..=Corridor  ><=Entrance  W=Wood F=Fire E=Earth M=Metal A=Water")
		fmt.Println("       WW=1beast  2F=2 fire beasts  11=RoomID(no beasts)")
	}

	if showAI {
		beasts := buildDemoBeasts()
		// Assign demo behavior states.
		beasts[0].State = senju.Patrolling
		beasts[1].State = senju.Idle // Guard
		beasts[2].State = senju.Chasing
		if showChi || showBeasts {
			fmt.Println()
		}
		fmt.Print(senju.RenderBehaviorOverlay(cave, beasts, nil))
		fmt.Println()
		fmt.Println("Legend: ██=Rock  ..=Corridor  ><=Entrance  ??=Invader")
		fmt.Println("       GG=Guard  PP=Patrol  !!=Chase  ++=Recovering  11=RoomID(no beasts)")
	}

	if showInvasion {
		waves := buildDemoWaves()
		if showChi || showBeasts || showAI {
			fmt.Println()
		}
		fmt.Print(invasion.RenderInvasionOverlay(cave, waves))
		fmt.Println()
		fmt.Println("Legend: ██=Rock  ..=Corridor  ><=Entrance  >>=Advancing  XX=Fighting")
		fmt.Println("       <<=Retreating  $$=GoalAchieved  3>>=count+state  11=RoomID(no invaders)")
	}

	if showEconomy {
		chiEngine, err := buildDemoEngine(cave)
		if err != nil {
			fmt.Printf("error building chi engine: %v\n", err)
			return
		}
		ecoEngine := buildDemoEconomyEngine()
		// Run 10 ticks of chi flow to populate RoomChi.
		for i := 0; i < 10; i++ {
			chiEngine.Tick()
		}
		// Prepare economy tick inputs.
		veins := make([]fengshui.DragonVein, len(chiEngine.Veins))
		for i, v := range chiEngine.Veins {
			veins[i] = *v
		}
		rooms := make([]world.Room, len(cave.Rooms))
		for i, r := range cave.Rooms {
			rooms[i] = *r
		}
		beasts := buildDemoBeasts()
		// Run one economy tick to get a result.
		result := ecoEngine.Tick(1, veins, chiEngine.RoomChi, 0.5, rooms, len(beasts), 0)
		if showChi || showBeasts || showAI || showInvasion {
			fmt.Println()
		}
		fmt.Println(economy.RenderEconomyStatus(ecoEngine, &result))
	}

	if !showChi && !showBeasts && !showAI && !showInvasion && !showEconomy {
		fmt.Print(cave.RenderASCII())
		fmt.Println()
		fmt.Println("Legend: ██=Rock  ..=Corridor  []=RoomFloor  ><=Entrance  1-9,A-Z=RoomID")
	}
}

// buildDemoCave creates a 24x20 cave with 4 rooms connected by corridors.
func buildDemoCave() (*world.Cave, error) {
	cave, err := world.NewCave(24, 20)
	if err != nil {
		return nil, err
	}

	// Room 1: 龍穴 (Earth) at (2,2) 4x3
	_, err = cave.AddRoom("dragon_hole", types.Pos{X: 2, Y: 2}, 4, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 4, Y: 4}, Dir: types.South},
	})
	if err != nil {
		return nil, fmt.Errorf("room 1: %w", err)
	}

	// Room 2: 蓄気室 (Water) at (10, 2) 3x3
	_, err = cave.AddRoom("chi_chamber", types.Pos{X: 10, Y: 2}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 10, Y: 4}, Dir: types.South},
	})
	if err != nil {
		return nil, fmt.Errorf("room 2: %w", err)
	}

	// Room 3: 仙獣部屋 (Wood) at (2, 10) 5x4
	_, err = cave.AddRoom("senju_room", types.Pos{X: 2, Y: 10}, 5, 4, []world.RoomEntrance{
		{Pos: types.Pos{X: 5, Y: 10}, Dir: types.North},
	})
	if err != nil {
		return nil, fmt.Errorf("room 3: %w", err)
	}

	// Room 4: 罠部屋 (Metal) at (14, 10) 4x3
	_, err = cave.AddRoom("trap_room", types.Pos{X: 14, Y: 10}, 4, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 14, Y: 11}, Dir: types.West},
	})
	if err != nil {
		return nil, fmt.Errorf("room 4: %w", err)
	}

	// Connect rooms
	if _, err = cave.ConnectRooms(1, 2); err != nil {
		return nil, fmt.Errorf("connect 1-2: %w", err)
	}
	if _, err = cave.ConnectRooms(1, 3); err != nil {
		return nil, fmt.Errorf("connect 1-3: %w", err)
	}
	if _, err = cave.ConnectRooms(3, 4); err != nil {
		return nil, fmt.Errorf("connect 3-4: %w", err)
	}

	return cave, nil
}

// buildDemoBeasts creates demo beasts placed in the cave rooms.
func buildDemoBeasts() []*senju.Beast {
	return []*senju.Beast{
		{ID: 1, SpeciesID: "suiryu", Name: "翠龍", Element: types.Wood, RoomID: 3, Level: 1},
		{ID: 2, SpeciesID: "enhou", Name: "炎鳳", Element: types.Fire, RoomID: 3, Level: 1},
		{ID: 3, SpeciesID: "kinrou", Name: "金狼", Element: types.Metal, RoomID: 4, Level: 1},
	}
}

// buildDemoEngine creates a ChiFlowEngine with two dragon veins for the demo cave.
func buildDemoEngine(cave *world.Cave) (*fengshui.ChiFlowEngine, error) {
	registry, err := world.LoadDefaultRoomTypes()
	if err != nil {
		return nil, fmt.Errorf("loading room types: %w", err)
	}

	// Build dragon veins from entrance positions.
	vein1, err := fengshui.BuildDragonVein(cave, types.Pos{X: 4, Y: 4}, types.Earth, 5.0)
	if err != nil {
		return nil, fmt.Errorf("building vein 1: %w", err)
	}
	vein1.ID = 1

	vein2, err := fengshui.BuildDragonVein(cave, types.Pos{X: 10, Y: 4}, types.Water, 3.0)
	if err != nil {
		return nil, fmt.Errorf("building vein 2: %w", err)
	}
	vein2.ID = 2

	params := fengshui.DefaultFlowParams()
	engine := fengshui.NewChiFlowEngine(cave, []*fengshui.DragonVein{vein1, vein2}, registry, params)

	return engine, nil
}

// buildDemoEconomyEngine creates an EconomyEngine with default parameters and
// a chi pool seeded with an initial balance for demo purposes.
func buildDemoEconomyEngine() *economy.EconomyEngine {
	pool := economy.NewChiPool(150.0)
	_ = pool.Deposit(50.0, economy.Supply, "initial", 0)
	return economy.NewEconomyEngine(
		pool,
		economy.DefaultSupplyParams(),
		economy.DefaultCostParams(),
		economy.DefaultDeficitParams(),
		economy.DefaultConstructionCost(),
		economy.DefaultBeastCost(),
	)
}

// buildDemoWaves creates demo invasion waves with invaders in various states.
// Room 1: 龍穴, Room 2: 蓄気室, Room 3: 仙獣部屋, Room 4: 罠部屋
func buildDemoWaves() []*invasion.InvasionWave {
	ascetic := invasion.InvaderClass{
		ID:               "wood_ascetic",
		Name:             "木行の修行者",
		Element:          types.Wood,
		BaseHP:           100,
		BaseATK:          25,
		BaseDEF:          20,
		BaseSPD:          20,
		RewardChi:        15.0,
		PreferredGoal:    invasion.DestroyCore,
		RetreatThreshold: 0.3,
	}
	thief := invasion.InvaderClass{
		ID:               "metal_thief",
		Name:             "金行の盗賊",
		Element:          types.Metal,
		BaseHP:           70,
		BaseATK:          20,
		BaseDEF:          15,
		BaseSPD:          40,
		RewardChi:        12.0,
		PreferredGoal:    invasion.StealTreasure,
		RetreatThreshold: 0.5,
	}

	// Invader 1: Advancing toward dragon hole (room 3 → room 1)
	inv1 := invasion.NewInvader(1, ascetic, 1, invasion.NewDestroyCoreGoal(), 3, 0)
	inv1.CurrentRoomID = 3
	inv1.State = invasion.Advancing

	// Invader 2: Fighting in trap room (room 4)
	inv2 := invasion.NewInvader(2, ascetic, 1, invasion.NewDestroyCoreGoal(), 4, 0)
	inv2.CurrentRoomID = 4
	inv2.State = invasion.Fighting

	// Invader 3: Retreating through senju room (room 3)
	inv3 := invasion.NewInvader(3, thief, 1, invasion.NewStealTreasureGoal(), 3, 0)
	inv3.CurrentRoomID = 3
	inv3.State = invasion.Retreating
	inv3.HP = 20

	// Invader 4: Goal achieved at chi chamber (room 2)
	inv4 := invasion.NewInvader(4, thief, 1, invasion.NewStealTreasureGoal(), 2, 0)
	inv4.CurrentRoomID = 2
	inv4.State = invasion.GoalAchieved

	wave := &invasion.InvasionWave{
		ID:          1,
		TriggerTick: 0,
		Invaders:    []*invasion.Invader{inv1, inv2, inv3, inv4},
		State:       invasion.Active,
		Difficulty:  1.0,
	}
	return []*invasion.InvasionWave{wave}
}

// displayScenarioStatus loads a scenario JSON file and prints its status line
// along with basic scenario information.
func displayScenarioStatus(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("error reading scenario file: %v\n", err)
		return
	}

	sc, err := scenario.LoadScenario(data)
	if err != nil {
		fmt.Printf("error loading scenario: %v\n", err)
		return
	}

	// Create initial progress (tick 0, no waves completed).
	prog := &scenario.ScenarioProgress{
		ScenarioID:  sc.ID,
		CurrentTick: 0,
		CoreHP:      100,
	}

	// Create a snapshot reflecting the initial state.
	snap := scenario.GameSnapshot{
		Tick:       0,
		CoreHP:     100,
		TotalWaves: len(sc.WaveSchedule),
	}

	fmt.Printf("Scenario: %s (%s)\n", sc.Name, sc.ID)
	fmt.Printf("Description: %s\n", sc.Description)
	fmt.Println(scenario.RenderScenarioStatus(sc, prog, snap))
}
