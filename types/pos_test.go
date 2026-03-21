package types

import "testing"

func TestPos_Add(t *testing.T) {
	tests := []struct {
		name string
		a, b Pos
		want Pos
	}{
		{"origin", Pos{0, 0}, Pos{0, 0}, Pos{0, 0}},
		{"positive", Pos{1, 2}, Pos{3, 4}, Pos{4, 6}},
		{"negative", Pos{1, 2}, Pos{-3, -4}, Pos{-2, -2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Add(tt.b)
			if got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPos_Sub(t *testing.T) {
	tests := []struct {
		name string
		a, b Pos
		want Pos
	}{
		{"same", Pos{3, 4}, Pos{3, 4}, Pos{0, 0}},
		{"positive", Pos{5, 7}, Pos{2, 3}, Pos{3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Sub(tt.b)
			if got != tt.want {
				t.Errorf("Sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPos_Distance(t *testing.T) {
	tests := []struct {
		name string
		a, b Pos
		want int
	}{
		{"same", Pos{0, 0}, Pos{0, 0}, 0},
		{"adjacent", Pos{0, 0}, Pos{1, 0}, 1},
		{"diagonal", Pos{0, 0}, Pos{3, 4}, 7},
		{"negative", Pos{-1, -2}, Pos{1, 2}, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Distance(tt.b)
			if got != tt.want {
				t.Errorf("Distance() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPos_Neighbors(t *testing.T) {
	p := Pos{5, 5}
	neighbors := p.Neighbors()
	expected := [4]Pos{
		{5, 4}, // North
		{5, 6}, // South
		{6, 5}, // East
		{4, 5}, // West
	}
	if neighbors != expected {
		t.Errorf("Neighbors() = %v, want %v", neighbors, expected)
	}
}
