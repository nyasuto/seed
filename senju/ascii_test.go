package senju

import (
	"strings"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// makeSmallCaveWithBeasts creates a 6x5 cave with two rooms for overlay testing.
func makeSmallCaveWithBeasts(t *testing.T) (*world.Cave, []*Beast) {
	t.Helper()

	cave, err := world.NewCave(6, 5)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1: 2x2 at (1,1) — senju_room (Wood)
	_, err = cave.AddRoom("senju_room", types.Pos{X: 1, Y: 1}, 2, 2, []world.RoomEntrance{
		{Pos: types.Pos{X: 1, Y: 3}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	// Room 2: 2x2 at (4,1) — trap_room (Metal)
	_, err = cave.AddRoom("trap_room", types.Pos{X: 4, Y: 1}, 2, 1, []world.RoomEntrance{
		{Pos: types.Pos{X: 4, Y: 2}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	beasts := []*Beast{
		{ID: 1, SpeciesID: "suiryu", Name: "翠龍", Element: types.Wood, RoomID: 1},
	}

	return cave, beasts
}

func TestRenderBeastOverlay_SingleBeast(t *testing.T) {
	cave, beasts := makeSmallCaveWithBeasts(t)

	got := RenderBeastOverlay(cave, beasts)

	// Room 1 has one Wood beast → cells should show "WW"
	if !strings.Contains(got, "WW") {
		t.Errorf("expected WW for single Wood beast, got:\n%s", got)
	}

	// Room 2 has no beasts → should show room ID "22"
	if !strings.Contains(got, "22") {
		t.Errorf("expected 22 for room without beasts, got:\n%s", got)
	}
}

func TestRenderBeastOverlay_MultipleBeastsInRoom(t *testing.T) {
	cave, beasts := makeSmallCaveWithBeasts(t)

	// Add a second beast to room 1
	beasts = append(beasts, &Beast{
		ID: 2, SpeciesID: "suiryu", Name: "翠龍2", Element: types.Wood, RoomID: 1,
	})

	got := RenderBeastOverlay(cave, beasts)

	// Room 1 has 2 Wood beasts → cells should show "2W"
	if !strings.Contains(got, "2W") {
		t.Errorf("expected 2W for two Wood beasts, got:\n%s", got)
	}
}

func TestRenderBeastOverlay_DifferentElements(t *testing.T) {
	tests := []struct {
		element types.Element
		want    string
	}{
		{types.Wood, "WW"},
		{types.Fire, "FF"},
		{types.Earth, "EE"},
		{types.Metal, "MM"},
		{types.Water, "AA"},
	}

	for _, tt := range tests {
		t.Run(tt.element.String(), func(t *testing.T) {
			cave, err := world.NewCave(4, 3)
			if err != nil {
				t.Fatalf("NewCave: %v", err)
			}

			_, err = cave.AddRoom("senju_room", types.Pos{X: 1, Y: 1}, 2, 1, nil)
			if err != nil {
				t.Fatalf("AddRoom: %v", err)
			}

			beasts := []*Beast{
				{ID: 1, Element: tt.element, RoomID: 1},
			}

			got := RenderBeastOverlay(cave, beasts)
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected %s for %s beast, got:\n%s", tt.want, tt.element, got)
			}
		})
	}
}

func TestRenderBeastOverlay_NoBeastsShowsRoomID(t *testing.T) {
	cave, err := world.NewCave(4, 3)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	_, err = cave.AddRoom("senju_room", types.Pos{X: 1, Y: 1}, 2, 1, nil)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	got := RenderBeastOverlay(cave, nil)

	// Room 1 with no beasts → should show "11"
	if !strings.Contains(got, "11") {
		t.Errorf("expected 11 for room without beasts, got:\n%s", got)
	}
}

func TestRenderBeastOverlay_UnassignedBeastIgnored(t *testing.T) {
	cave, err := world.NewCave(4, 3)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	_, err = cave.AddRoom("senju_room", types.Pos{X: 1, Y: 1}, 2, 1, nil)
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	// Beast with RoomID 0 (unassigned) should not appear
	beasts := []*Beast{
		{ID: 1, Element: types.Fire, RoomID: 0},
	}

	got := RenderBeastOverlay(cave, beasts)

	// Room should show room ID, not beast
	if !strings.Contains(got, "11") {
		t.Errorf("expected 11 for room (unassigned beast), got:\n%s", got)
	}
	if strings.Contains(got, "FF") {
		t.Errorf("unassigned beast should not appear in overlay, got:\n%s", got)
	}
}

func TestRenderBeastOverlay_OtherCellTypes(t *testing.T) {
	cave, beasts := makeSmallCaveWithBeasts(t)

	got := RenderBeastOverlay(cave, beasts)

	// Should contain rock
	if !strings.Contains(got, "██") {
		t.Errorf("expected ██ for rock, got:\n%s", got)
	}

	// Should contain entrance
	if !strings.Contains(got, "><") {
		t.Errorf("expected >< for entrance, got:\n%s", got)
	}
}

// --- Behavior overlay tests ---

func TestRenderBehaviorOverlay_GuardState(t *testing.T) {
	cave, beasts := makeSmallCaveWithBeasts(t)
	beasts[0].State = Idle // Guard beasts are Idle

	got := RenderBehaviorOverlay(cave, beasts, nil)

	if !strings.Contains(got, "GG") {
		t.Errorf("expected GG for Guard/Idle beast, got:\n%s", got)
	}
}

func TestRenderBehaviorOverlay_PatrolState(t *testing.T) {
	cave, beasts := makeSmallCaveWithBeasts(t)
	beasts[0].State = Patrolling

	got := RenderBehaviorOverlay(cave, beasts, nil)

	if !strings.Contains(got, "PP") {
		t.Errorf("expected PP for Patrolling beast, got:\n%s", got)
	}
}

func TestRenderBehaviorOverlay_ChaseState(t *testing.T) {
	cave, beasts := makeSmallCaveWithBeasts(t)
	beasts[0].State = Chasing

	got := RenderBehaviorOverlay(cave, beasts, nil)

	if !strings.Contains(got, "!!") {
		t.Errorf("expected !! for Chasing beast, got:\n%s", got)
	}
}

func TestRenderBehaviorOverlay_FleeRecoveringState(t *testing.T) {
	cave, beasts := makeSmallCaveWithBeasts(t)
	beasts[0].State = Recovering

	got := RenderBehaviorOverlay(cave, beasts, nil)

	if !strings.Contains(got, "++") {
		t.Errorf("expected ++ for Recovering beast, got:\n%s", got)
	}
}

func TestRenderBehaviorOverlay_InvaderPlaceholder(t *testing.T) {
	cave, beasts := makeSmallCaveWithBeasts(t)
	// Room 2 has invaders
	invaders := map[int][]int{2: {100}}

	got := RenderBehaviorOverlay(cave, beasts, invaders)

	// Room 2 should show "??" for invader placeholder
	if !strings.Contains(got, "??") {
		t.Errorf("expected ?? for invader placeholder, got:\n%s", got)
	}
}

func TestRenderBehaviorOverlay_MultipleBeastsShowCount(t *testing.T) {
	cave, beasts := makeSmallCaveWithBeasts(t)
	beasts[0].State = Patrolling
	beasts = append(beasts, &Beast{
		ID: 2, Element: types.Fire, RoomID: 1, State: Patrolling,
	})

	got := RenderBehaviorOverlay(cave, beasts, nil)

	// 2 beasts in room 1, first is Patrolling → "2P"
	if !strings.Contains(got, "2P") {
		t.Errorf("expected 2P for two Patrolling beasts, got:\n%s", got)
	}
}

func TestRenderBehaviorOverlay_NoBeastsShowsRoomID(t *testing.T) {
	cave, _ := makeSmallCaveWithBeasts(t)

	got := RenderBehaviorOverlay(cave, nil, nil)

	if !strings.Contains(got, "11") {
		t.Errorf("expected 11 for room without beasts, got:\n%s", got)
	}
}

func TestStateTag(t *testing.T) {
	tests := []struct {
		state BeastState
		want  string
	}{
		{Idle, "[G]"},
		{Patrolling, "[P]"},
		{Chasing, "[!]"},
		{Fighting, "[!]"},
		{Recovering, "[+]"},
	}
	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			got := stateTag(tt.state)
			if got != tt.want {
				t.Errorf("stateTag(%v) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}
