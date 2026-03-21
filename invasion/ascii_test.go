package invasion

import (
	"strings"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// makeSmallCaveForInvasion creates a 6x5 cave with two rooms for overlay testing.
// Room 1: 2x2 at (1,1), Room 2: 2x1 at (4,1).
func makeSmallCaveForInvasion(t *testing.T) *world.Cave {
	t.Helper()

	cave, err := world.NewCave(6, 5)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	_, err = cave.AddRoom("dragon_hole", types.Pos{X: 1, Y: 1}, 2, 2, []world.RoomEntrance{
		{Pos: types.Pos{X: 1, Y: 3}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	_, err = cave.AddRoom("storage", types.Pos{X: 4, Y: 1}, 2, 1, []world.RoomEntrance{
		{Pos: types.Pos{X: 4, Y: 2}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	return cave
}

func makeOverlayInvader(id int, roomID int, state InvaderState) *Invader {
	return &Invader{
		ID:            id,
		CurrentRoomID: roomID,
		State:         state,
	}
}

func TestRenderInvasionOverlay_Advancing(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: []*Invader{
			makeOverlayInvader(1, 1, Advancing),
		}},
	}

	got := RenderInvasionOverlay(cave, waves)

	if !strings.Contains(got, ">>") {
		t.Errorf("expected >> for Advancing invader, got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_Fighting(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: []*Invader{
			makeOverlayInvader(1, 1, Fighting),
		}},
	}

	got := RenderInvasionOverlay(cave, waves)

	if !strings.Contains(got, "XX") {
		t.Errorf("expected XX for Fighting invader, got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_Retreating(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: []*Invader{
			makeOverlayInvader(1, 1, Retreating),
		}},
	}

	got := RenderInvasionOverlay(cave, waves)

	if !strings.Contains(got, "<<") {
		t.Errorf("expected << for Retreating invader, got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_GoalAchieved(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: []*Invader{
			makeOverlayInvader(1, 1, GoalAchieved),
		}},
	}

	got := RenderInvasionOverlay(cave, waves)

	if !strings.Contains(got, "$$") {
		t.Errorf("expected $$ for GoalAchieved invader, got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_DefeatedExcluded(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: []*Invader{
			makeOverlayInvader(1, 1, Defeated),
		}},
	}

	got := RenderInvasionOverlay(cave, waves)

	// Defeated invader should not appear; room should show room ID "11"
	if !strings.Contains(got, "11") {
		t.Errorf("expected 11 for room with defeated invader, got:\n%s", got)
	}
	// Should not contain any invader symbols
	for _, sym := range []string{">>", "XX", "<<", "$$"} {
		if strings.Contains(got, sym) {
			t.Errorf("defeated invader should not produce %s, got:\n%s", sym, got)
		}
	}
}

func TestRenderInvasionOverlay_MultipleInvadersCount(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: []*Invader{
			makeOverlayInvader(1, 1, Advancing),
			makeOverlayInvader(2, 1, Advancing),
			makeOverlayInvader(3, 1, Advancing),
		}},
	}

	got := RenderInvasionOverlay(cave, waves)

	if !strings.Contains(got, "3>") {
		t.Errorf("expected 3> for three advancing invaders, got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_TwoInvaders(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: []*Invader{
			makeOverlayInvader(1, 2, Retreating),
			makeOverlayInvader(2, 2, Retreating),
		}},
	}

	got := RenderInvasionOverlay(cave, waves)

	if !strings.Contains(got, "2<") {
		t.Errorf("expected 2< for two retreating invaders, got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_NoInvadersShowsRoomID(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{}

	got := RenderInvasionOverlay(cave, waves)

	if !strings.Contains(got, "11") {
		t.Errorf("expected 11 for room 1 with no invaders, got:\n%s", got)
	}
	if !strings.Contains(got, "22") {
		t.Errorf("expected 22 for room 2 with no invaders, got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_UnassignedRoomIDIgnored(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: []*Invader{
			makeOverlayInvader(1, 0, Advancing), // roomID 0 = not yet in cave
		}},
	}

	got := RenderInvasionOverlay(cave, waves)

	// Room should show room ID, not invader
	if !strings.Contains(got, "11") {
		t.Errorf("expected 11 for room (invader not assigned), got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_MultipleWaves(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: []*Invader{
			makeOverlayInvader(1, 1, Advancing),
		}},
		{ID: 2, State: Active, Invaders: []*Invader{
			makeOverlayInvader(2, 1, Fighting),
		}},
	}

	got := RenderInvasionOverlay(cave, waves)

	// Two invaders from different waves in room 1 → count 2
	if !strings.Contains(got, "2>") {
		t.Errorf("expected 2> for two invaders from different waves, got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_OtherCellTypes(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)

	got := RenderInvasionOverlay(cave, nil)

	if !strings.Contains(got, "██") {
		t.Errorf("expected ██ for rock, got:\n%s", got)
	}
	if !strings.Contains(got, "><") {
		t.Errorf("expected >< for entrance, got:\n%s", got)
	}
}

func TestRenderInvasionOverlay_NineInvadersCap(t *testing.T) {
	cave := makeSmallCaveForInvasion(t)
	invaders := make([]*Invader, 10)
	for i := range invaders {
		invaders[i] = makeOverlayInvader(i+1, 1, Advancing)
	}
	waves := []*InvasionWave{
		{ID: 1, State: Active, Invaders: invaders},
	}

	got := RenderInvasionOverlay(cave, waves)

	if !strings.Contains(got, "9+") {
		t.Errorf("expected 9+ for 10 invaders, got:\n%s", got)
	}
}

func TestStateSymbol(t *testing.T) {
	tests := []struct {
		state InvaderState
		want  string
	}{
		{Advancing, ">>"},
		{Fighting, "XX"},
		{Retreating, "<<"},
		{GoalAchieved, "$$"},
	}
	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			got := stateSymbol(tt.state)
			if got != tt.want {
				t.Errorf("stateSymbol(%v) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestInvaderTile_Single(t *testing.T) {
	tests := []struct {
		state InvaderState
		want  string
	}{
		{Advancing, ">>"},
		{Fighting, "XX"},
		{Retreating, "<<"},
		{GoalAchieved, "$$"},
	}
	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			info := &invaderOverlay{count: 1, state: tt.state}
			got := invaderTile(info)
			if got != tt.want {
				t.Errorf("invaderTile(1, %v) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestInvaderTile_Multiple(t *testing.T) {
	tests := []struct {
		count int
		state InvaderState
		want  string
	}{
		{2, Advancing, "2>"},
		{3, Fighting, "3X"},
		{5, Retreating, "5<"},
		{9, GoalAchieved, "9$"},
		{10, Advancing, "9+"},
		{99, Fighting, "9+"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			info := &invaderOverlay{count: tt.count, state: tt.state}
			got := invaderTile(info)
			if got != tt.want {
				t.Errorf("invaderTile(%d, %v) = %q, want %q", tt.count, tt.state, got, tt.want)
			}
		})
	}
}
