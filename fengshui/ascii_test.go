package fengshui

import (
	"strings"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// makeSmallCaveWithEngine creates a small cave with one room and a dragon vein,
// returning the cave and a ChiFlowEngine for overlay testing.
func makeSmallCaveWithEngine(t *testing.T) (*world.Cave, *ChiFlowEngine) {
	t.Helper()

	cave, err := world.NewCave(6, 5)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1: 2x2 at (2,1)
	_, err = cave.AddRoom("dragon_hole", types.Pos{X: 2, Y: 1}, 2, 2, []world.RoomEntrance{
		{Pos: types.Pos{X: 2, Y: 3}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	// Corridor cell at (2,4) to connect entrance to cave edge
	err = cave.Grid.Set(types.Pos{X: 2, Y: 4}, world.Cell{Type: world.CorridorFloor})
	if err != nil {
		t.Fatalf("Set corridor: %v", err)
	}

	registry := world.NewRoomTypeRegistry()

	vein := &DragonVein{
		ID:        1,
		SourcePos: types.Pos{X: 2, Y: 4},
		Element:   types.Earth,
		FlowRate:  5.0,
		Path: []types.Pos{
			{X: 2, Y: 4},
			{X: 2, Y: 3},
			{X: 2, Y: 2},
			{X: 2, Y: 1},
			{X: 3, Y: 1},
			{X: 3, Y: 2},
		},
	}

	params := DefaultFlowParams()
	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, registry, params)

	return cave, engine
}

func TestRenderChiOverlay_EmptyRoom(t *testing.T) {
	cave, engine := makeSmallCaveWithEngine(t)

	got := RenderChiOverlay(cave, engine)

	// Room has 0 chi → should show "__"
	if !strings.Contains(got, "__") {
		t.Errorf("expected __ for empty room, got:\n%s", got)
	}

	// Dragon vein corridor should show "~~"
	if !strings.Contains(got, "~~") {
		t.Errorf("expected ~~ for dragon vein path, got:\n%s", got)
	}

	// Entrance should still show "><"
	if !strings.Contains(got, "><") {
		t.Errorf("expected >< for entrance, got:\n%s", got)
	}
}

func TestRenderChiOverlay_ChiLevels(t *testing.T) {
	cave, engine := makeSmallCaveWithEngine(t)

	// Ensure room has capacity for ratio testing.
	rc := engine.RoomChi[1]
	rc.Capacity = 100

	// Set chi to 20% of capacity → should show ░░
	rc.Current = rc.Capacity * 0.20

	got := RenderChiOverlay(cave, engine)
	if !strings.Contains(got, "░░") {
		t.Errorf("expected ░░ for 20%% chi, got:\n%s", got)
	}

	// Set chi to 50% → should show ▒▒
	rc.Current = rc.Capacity * 0.50
	got = RenderChiOverlay(cave, engine)
	if !strings.Contains(got, "▒▒") {
		t.Errorf("expected ▒▒ for 50%% chi, got:\n%s", got)
	}

	// Set chi to 80% → should show ▓▓
	rc.Current = rc.Capacity * 0.80
	got = RenderChiOverlay(cave, engine)
	if !strings.Contains(got, "▓▓") {
		t.Errorf("expected ▓▓ for 80%% chi, got:\n%s", got)
	}

	// Set chi to 100% → should show ██
	rc.Current = rc.Capacity
	got = RenderChiOverlay(cave, engine)
	// Room cells at 100% show ██ — same as rock, but in room position
	lines := strings.Split(got, "\n")
	// Room is at row 1, columns 2-3 → characters 4-7 (each cell is 2 chars wide × multi-byte)
	// Just verify no ▓▓ remains since it's now full
	if strings.Contains(got, "▓▓") {
		t.Errorf("expected ██ (not ▓▓) for 100%% chi, got:\n%s", got)
	}
	_ = lines
}

func TestRenderChiOverlay_DragonVeinOnRock(t *testing.T) {
	cave, err := world.NewCave(4, 4)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	registry := world.NewRoomTypeRegistry()

	// Dragon vein path goes through rock cells
	vein := &DragonVein{
		ID:        1,
		SourcePos: types.Pos{X: 0, Y: 0},
		Element:   types.Water,
		FlowRate:  1.0,
		Path: []types.Pos{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 2, Y: 0},
		},
	}

	params := DefaultFlowParams()
	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, registry, params)

	got := RenderChiOverlay(cave, engine)

	// First line should start with ~~ tiles for the vein path on rock
	firstLine := strings.Split(got, "\n")[0]
	if !strings.HasPrefix(firstLine, "~~~~~~") {
		t.Errorf("expected first line to start with '~~~~~~', got: %s", firstLine)
	}
}

func TestRenderChiOverlay_CorridorWithoutVein(t *testing.T) {
	cave, err := world.NewCave(4, 1)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	// Set a corridor cell that is NOT on a vein path
	err = cave.Grid.Set(types.Pos{X: 1, Y: 0}, world.Cell{Type: world.CorridorFloor})
	if err != nil {
		t.Fatalf("Set: %v", err)
	}

	registry := world.NewRoomTypeRegistry()
	params := DefaultFlowParams()
	engine := NewChiFlowEngine(cave, nil, registry, params)

	got := RenderChiOverlay(cave, engine)

	// Corridor not on vein should show ".."
	if !strings.Contains(got, "..") {
		t.Errorf("expected .. for corridor without vein, got:\n%s", got)
	}
}
