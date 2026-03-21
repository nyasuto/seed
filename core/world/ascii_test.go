package world

import (
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func TestRenderASCII_EmptyGrid(t *testing.T) {
	cave, err := NewCave(3, 2)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	got := cave.RenderASCII()
	// 3x2 all rock
	want := "██████\n██████\n"
	if got != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderASCII_WithRoomAndCorridor(t *testing.T) {
	// 8x8 cave with one room and a corridor cell
	cave, err := NewCave(8, 8)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Place a 2x2 room at (1,1) with entrance at (1,3)
	_, err = cave.AddRoom("test", types.Pos{X: 1, Y: 1}, 2, 2, []RoomEntrance{
		{Pos: types.Pos{X: 1, Y: 3}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	got := cave.RenderASCII()

	// Verify room cells show RoomID '1'
	if !strings.Contains(got, "11") {
		t.Error("expected room ID '11' in output")
	}
	// Verify entrance shows ><
	if !strings.Contains(got, "><") {
		t.Error("expected entrance '><' in output")
	}

	// Manually set a corridor cell to verify corridor rendering
	err = cave.Grid.Set(types.Pos{X: 1, Y: 4}, Cell{Type: CorridorFloor})
	if err != nil {
		t.Fatalf("Set corridor: %v", err)
	}
	got = cave.RenderASCII()
	if !strings.Contains(got, "..") {
		t.Error("expected corridor '..' in output")
	}
}

func TestRenderASCII_MultipleRooms(t *testing.T) {
	cave, err := NewCave(8, 6)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	// Room 1 at (0,0) 2x2
	_, err = cave.AddRoom("a", types.Pos{X: 0, Y: 0}, 2, 2, nil)
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	// Room 2 at (4,0) 2x2
	_, err = cave.AddRoom("b", types.Pos{X: 4, Y: 0}, 2, 2, nil)
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	got := cave.RenderASCII()

	if !strings.Contains(got, "11") {
		t.Error("expected room ID '1' in output")
	}
	if !strings.Contains(got, "22") {
		t.Error("expected room ID '2' in output")
	}
}

func TestRoomIDChar(t *testing.T) {
	tests := []struct {
		id   int
		want byte
	}{
		{1, '1'},
		{9, '9'},
		{10, 'A'},
		{11, 'B'},
		{35, 'Z'},
	}
	for _, tt := range tests {
		got := RoomIDChar(tt.id)
		if got != tt.want {
			t.Errorf("RoomIDChar(%d) = %c, want %c", tt.id, got, tt.want)
		}
	}
}
