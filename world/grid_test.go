package world

import (
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestNewGrid(t *testing.T) {
	tests := []struct {
		name    string
		w, h    int
		wantErr bool
	}{
		{"valid grid", 10, 8, false},
		{"1x1 grid", 1, 1, false},
		{"zero width", 0, 5, true},
		{"zero height", 5, 0, true},
		{"negative width", -1, 5, true},
		{"negative height", 5, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGrid(tt.w, tt.h)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NewGrid(%d, %d) expected error, got nil", tt.w, tt.h)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewGrid(%d, %d) unexpected error: %v", tt.w, tt.h, err)
			}
			if g.Width != tt.w || g.Height != tt.h {
				t.Errorf("got dimensions %dx%d, want %dx%d", g.Width, g.Height, tt.w, tt.h)
			}
		})
	}
}

func TestGrid_DefaultCellsAreRock(t *testing.T) {
	g, err := NewGrid(3, 3)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			cell, err := g.At(types.Pos{X: x, Y: y})
			if err != nil {
				t.Fatalf("At(%d,%d): %v", x, y, err)
			}
			if cell.Type != Rock {
				t.Errorf("At(%d,%d).Type = %v, want Rock", x, y, cell.Type)
			}
			if cell.RoomID != 0 {
				t.Errorf("At(%d,%d).RoomID = %d, want 0", x, y, cell.RoomID)
			}
		}
	}
}

func TestGrid_SetAndAt(t *testing.T) {
	g, err := NewGrid(5, 5)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	pos := types.Pos{X: 2, Y: 3}
	want := Cell{Type: RoomFloor, RoomID: 42}

	if err := g.Set(pos, want); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := g.At(pos)
	if err != nil {
		t.Fatalf("At: %v", err)
	}
	if got != want {
		t.Errorf("At(%v) = %+v, want %+v", pos, got, want)
	}
}

func TestGrid_InBounds(t *testing.T) {
	g, err := NewGrid(4, 3)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	tests := []struct {
		pos  types.Pos
		want bool
	}{
		{types.Pos{X: 0, Y: 0}, true},
		{types.Pos{X: 3, Y: 2}, true},
		{types.Pos{X: 4, Y: 0}, false},
		{types.Pos{X: 0, Y: 3}, false},
		{types.Pos{X: -1, Y: 0}, false},
		{types.Pos{X: 0, Y: -1}, false},
	}

	for _, tt := range tests {
		got := g.InBounds(tt.pos)
		if got != tt.want {
			t.Errorf("InBounds(%v) = %v, want %v", tt.pos, got, tt.want)
		}
	}
}

func TestGrid_OutOfBoundsError(t *testing.T) {
	g, err := NewGrid(3, 3)
	if err != nil {
		t.Fatalf("NewGrid: %v", err)
	}

	outPositions := []types.Pos{
		{X: -1, Y: 0},
		{X: 0, Y: -1},
		{X: 3, Y: 0},
		{X: 0, Y: 3},
	}

	for _, pos := range outPositions {
		_, err := g.At(pos)
		if err == nil {
			t.Errorf("At(%v) expected error, got nil", pos)
		}
		if !errors.Is(err, ErrOutOfBounds) {
			t.Errorf("At(%v) error = %v, want ErrOutOfBounds", pos, err)
		}

		err = g.Set(pos, Cell{Type: Corridor})
		if err == nil {
			t.Errorf("Set(%v) expected error, got nil", pos)
		}
		if !errors.Is(err, ErrOutOfBounds) {
			t.Errorf("Set(%v) error = %v, want ErrOutOfBounds", pos, err)
		}
	}
}
