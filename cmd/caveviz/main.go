// Command caveviz generates a hardcoded Cave and prints its ASCII representation.
package main

import (
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

func main() {
	cave, err := buildDemoCave()
	if err != nil {
		fmt.Printf("error building demo cave: %v\n", err)
		return
	}

	fmt.Print(cave.RenderASCII())
	fmt.Println()
	fmt.Println("Legend: ██=Rock  ..=Corridor  []=RoomFloor  ><= Entrance  1-9,A-Z=RoomID")
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

