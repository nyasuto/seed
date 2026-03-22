package view

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/asset"
)

// newEntityTestCave creates a minimal Cave with one room at the given position and size.
func newEntityTestCave(roomID int, posX, posY, w, h int) *world.Cave {
	cave, _ := world.NewCave(24, 20)
	room := &world.Room{
		ID:     roomID,
		TypeID: "fire_room",
		Pos:    types.Pos{X: posX, Y: posY},
		Width:  w,
		Height: h,
		Level:  1,
	}
	cave.Rooms = append(cave.Rooms, room)
	return cave
}

func TestRoomCenter_BasicRoom(t *testing.T) {
	// Room at (4,2) with size 3x3 → center cell is (5,3).
	cave := newEntityTestCave(1, 4, 2, 3, 3)
	mv := NewMapView(0, 0)
	er := NewEntityRenderer(mv)

	px, py, ok := er.RoomCenter(cave, 1)
	if !ok {
		t.Fatal("RoomCenter returned ok=false for existing room")
	}

	// Center cell (5,3) → screen top-left (5*32, 3*32) = (160, 96)
	// Plus half tile → (160+16, 96+16) = (176, 112)
	wantX := 5*asset.TileSize + asset.TileSize/2
	wantY := 3*asset.TileSize + asset.TileSize/2
	if px != wantX || py != wantY {
		t.Errorf("RoomCenter(1) = (%d,%d), want (%d,%d)", px, py, wantX, wantY)
	}
}

func TestRoomCenter_WithMapOffset(t *testing.T) {
	cave := newEntityTestCave(2, 6, 4, 4, 4)
	mv := NewMapView(32, 32)
	er := NewEntityRenderer(mv)

	px, py, ok := er.RoomCenter(cave, 2)
	if !ok {
		t.Fatal("RoomCenter returned ok=false")
	}

	// Center cell (8,6) → screen (8*32+32, 6*32+32) = (288, 224)
	// Plus half tile → (288+16, 224+16) = (304, 240)
	wantX := 8*asset.TileSize + 32 + asset.TileSize/2
	wantY := 6*asset.TileSize + 32 + asset.TileSize/2
	if px != wantX || py != wantY {
		t.Errorf("RoomCenter(2) = (%d,%d), want (%d,%d)", px, py, wantX, wantY)
	}
}

func TestRoomCenter_NotFound(t *testing.T) {
	cave := newEntityTestCave(1, 0, 0, 3, 3)
	mv := NewMapView(0, 0)
	er := NewEntityRenderer(mv)

	_, _, ok := er.RoomCenter(cave, 99)
	if ok {
		t.Error("RoomCenter should return ok=false for non-existent room")
	}
}

func TestRoomCenter_BeastRoomID(t *testing.T) {
	// Simulate a beast in room 1: verify its room maps to the correct screen position.
	cave := newEntityTestCave(1, 2, 3, 5, 5)
	mv := NewMapView(0, 0)
	er := NewEntityRenderer(mv)

	beastRoomID := 1
	px, py, ok := er.RoomCenter(cave, beastRoomID)
	if !ok {
		t.Fatal("RoomCenter returned ok=false")
	}

	// Room at (2,3), size 5x5 → center cell (4,5)
	// Screen: (4*32+16, 5*32+16) = (144, 176)
	wantX := 4*asset.TileSize + asset.TileSize/2
	wantY := 5*asset.TileSize + asset.TileSize/2
	if px != wantX || py != wantY {
		t.Errorf("Beast room center = (%d,%d), want (%d,%d)", px, py, wantX, wantY)
	}
}

func TestRoomCenter_InvaderRoomID(t *testing.T) {
	// Simulate an invader in room 1.
	cave := newEntityTestCave(1, 10, 8, 3, 3)
	mv := NewMapView(0, 0)
	er := NewEntityRenderer(mv)

	invaderRoomID := 1
	px, py, ok := er.RoomCenter(cave, invaderRoomID)
	if !ok {
		t.Fatal("RoomCenter returned ok=false")
	}

	// Room at (10,8), size 3x3 → center cell (11,9)
	// Screen: (11*32+16, 9*32+16) = (368, 304)
	wantX := 11*asset.TileSize + asset.TileSize/2
	wantY := 9*asset.TileSize + asset.TileSize/2
	if px != wantX || py != wantY {
		t.Errorf("Invader room center = (%d,%d), want (%d,%d)", px, py, wantX, wantY)
	}
}

func TestBuildRoomRenderMap_FireRoom(t *testing.T) {
	reg := world.NewRoomTypeRegistry()
	_ = reg.Register(world.RoomType{
		ID:      "fire_room",
		Name:    "Fire Room",
		Element: types.Fire,
	})

	cave, _ := world.NewCave(24, 20)
	cave.Rooms = append(cave.Rooms, &world.Room{
		ID:     1,
		TypeID: "fire_room",
		Pos:    types.Pos{X: 4, Y: 4},
		Width:  3,
		Height: 3,
		Level:  1,
	})

	rm := BuildRoomRenderMap(cave, reg)
	info, ok := rm[1]
	if !ok {
		t.Fatal("room 1 not found in render map")
	}
	if info.Element != types.Fire {
		t.Errorf("Element = %v, want Fire", info.Element)
	}
	if info.IsDragonHole {
		t.Error("fire_room should not be dragon hole")
	}
}

func TestBuildRoomRenderMap_DragonHole(t *testing.T) {
	reg := world.NewRoomTypeRegistry()
	_ = reg.Register(world.RoomType{
		ID:         "dragon_hole",
		Name:       "Dragon Hole",
		Element:    types.Earth,
		BaseCoreHP: 100,
	})

	cave, _ := world.NewCave(24, 20)
	cave.Rooms = append(cave.Rooms, &world.Room{
		ID:     1,
		TypeID: "dragon_hole",
		Pos:    types.Pos{X: 8, Y: 8},
		Width:  3,
		Height: 3,
		Level:  1,
	})

	rm := BuildRoomRenderMap(cave, reg)
	info, ok := rm[1]
	if !ok {
		t.Fatal("room 1 not found in render map")
	}
	if !info.IsDragonHole {
		t.Error("dragon_hole should be identified as dragon hole")
	}
}

func TestBuildRoomRenderMap_SnapshotFireColor(t *testing.T) {
	// Completion criteria: Snapshot の部屋属性が Fire の場合、
	// 対応するセルが Fire 色タイルで描画される座標計算テスト
	reg := world.NewRoomTypeRegistry()
	_ = reg.Register(world.RoomType{
		ID:      "fire_room",
		Name:    "Fire Room",
		Element: types.Fire,
	})

	cave, _ := world.NewCave(24, 20)
	room := &world.Room{
		ID:     1,
		TypeID: "fire_room",
		Pos:    types.Pos{X: 4, Y: 4},
		Width:  3,
		Height: 3,
		Level:  1,
	}
	cave.Rooms = append(cave.Rooms, room)

	rm := BuildRoomRenderMap(cave, reg)

	// Verify that for any cell with RoomID=1, the render info gives Fire element.
	info := rm[1]
	if info.Element != types.Fire {
		t.Errorf("room 1 element = %v, want Fire", info.Element)
	}

	// Verify coordinate: room cell (5,5) → screen position via MapView.
	mv := NewMapView(0, 0)
	px, py := mv.CellToScreen(5, 5)
	wantX := 5 * asset.TileSize
	wantY := 5 * asset.TileSize
	if px != wantX || py != wantY {
		t.Errorf("CellToScreen(5,5) = (%d,%d), want (%d,%d)", px, py, wantX, wantY)
	}
}
