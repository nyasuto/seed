package types

import "testing"

func TestDirection_Opposite(t *testing.T) {
	tests := []struct {
		dir  Direction
		want Direction
	}{
		{North, South},
		{South, North},
		{East, West},
		{West, East},
	}
	for _, tt := range tests {
		t.Run(tt.dir.String(), func(t *testing.T) {
			got := tt.dir.Opposite()
			if got != tt.want {
				t.Errorf("%v.Opposite() = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestDirection_Delta(t *testing.T) {
	tests := []struct {
		dir  Direction
		want Pos
	}{
		{North, Pos{0, -1}},
		{South, Pos{0, 1}},
		{East, Pos{1, 0}},
		{West, Pos{-1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.dir.String(), func(t *testing.T) {
			got := tt.dir.Delta()
			if got != tt.want {
				t.Errorf("%v.Delta() = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestElement_String(t *testing.T) {
	tests := []struct {
		e    Element
		want string
	}{
		{Wood, "Wood"},
		{Fire, "Fire"},
		{Earth, "Earth"},
		{Metal, "Metal"},
		{Water, "Water"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("Element(%d).String() = %q, want %q", tt.e, got, tt.want)
			}
		})
	}
}

// TestGenerates_AllCombinations tests the productive cycle (相生):
// Wood → Fire → Earth → Metal → Water → Wood
func TestGenerates_AllCombinations(t *testing.T) {
	generating := [][2]Element{
		{Wood, Fire},
		{Fire, Earth},
		{Earth, Metal},
		{Metal, Water},
		{Water, Wood},
	}

	generatingSet := make(map[[2]Element]bool)
	for _, pair := range generating {
		generatingSet[pair] = true
	}

	for from := range Element(ElementCount) {
		for to := range Element(ElementCount) {
			want := generatingSet[[2]Element{from, to}]
			got := Generates(from, to)
			if got != want {
				t.Errorf("Generates(%v, %v) = %v, want %v", from, to, got, want)
			}
		}
	}
}

// TestOvercomes_AllCombinations tests the destructive cycle (相克):
// Wood → Earth → Water → Fire → Metal → Wood
func TestOvercomes_AllCombinations(t *testing.T) {
	overcoming := [][2]Element{
		{Wood, Earth},
		{Earth, Water},
		{Water, Fire},
		{Fire, Metal},
		{Metal, Wood},
	}

	overcomingSet := make(map[[2]Element]bool)
	for _, pair := range overcoming {
		overcomingSet[pair] = true
	}

	for from := range Element(ElementCount) {
		for to := range Element(ElementCount) {
			want := overcomingSet[[2]Element{from, to}]
			got := Overcomes(from, to)
			if got != want {
				t.Errorf("Overcomes(%v, %v) = %v, want %v", from, to, got, want)
			}
		}
	}
}
