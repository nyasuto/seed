// Command caveviz generates a hardcoded Cave and prints its ASCII representation.
// Use --chi to display a chi flow overlay with dragon veins.
// Use --beasts to display beast placements.
// Use --all to display all layers (standard + chi + beasts).
package main

import (
	"flag"
	"fmt"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

func main() {
	chiMode := flag.Bool("chi", false, "display chi flow overlay")
	beastMode := flag.Bool("beasts", false, "display beast placement overlay")
	aiMode := flag.Bool("ai", false, "display beast behavior state overlay")
	allMode := flag.Bool("all", false, "display all layers (standard + chi + beasts + ai)")
	flag.Parse()

	cave, err := buildDemoCave()
	if err != nil {
		fmt.Printf("error building demo cave: %v\n", err)
		return
	}

	showChi := *chiMode || *allMode
	showBeasts := *beastMode || *allMode
	showAI := *aiMode || *allMode

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

	if !showChi && !showBeasts && !showAI {
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
