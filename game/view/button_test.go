package view

import "testing"

func TestButton_Contains_Inside(t *testing.T) {
	btn := NewButton(100, 200, 80, 30, "Test")

	tests := []struct {
		name string
		px   int
		py   int
		want bool
	}{
		{"center", 140, 215, true},
		{"top-left corner", 100, 200, true},
		{"bottom-right inside", 179, 229, true},
		{"left edge", 100, 215, true},
		{"top edge", 140, 200, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := btn.Contains(tt.px, tt.py); got != tt.want {
				t.Errorf("Contains(%d, %d) = %v, want %v", tt.px, tt.py, got, tt.want)
			}
		})
	}
}

func TestButton_Contains_Outside(t *testing.T) {
	btn := NewButton(100, 200, 80, 30, "Test")

	tests := []struct {
		name string
		px   int
		py   int
	}{
		{"left of button", 99, 215},
		{"above button", 140, 199},
		{"right of button", 180, 215},
		{"below button", 140, 230},
		{"far away", 0, 0},
		{"negative coords", -10, -10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if btn.Contains(tt.px, tt.py) {
				t.Errorf("Contains(%d, %d) = true, want false", tt.px, tt.py)
			}
		})
	}
}

func TestNewButton_Rect(t *testing.T) {
	btn := NewButton(10, 20, 50, 30, "OK")
	if btn.Rect.Min.X != 10 || btn.Rect.Min.Y != 20 {
		t.Errorf("Rect.Min = (%d,%d), want (10,20)", btn.Rect.Min.X, btn.Rect.Min.Y)
	}
	if btn.Rect.Max.X != 60 || btn.Rect.Max.Y != 50 {
		t.Errorf("Rect.Max = (%d,%d), want (60,50)", btn.Rect.Max.X, btn.Rect.Max.Y)
	}
	if btn.Label != "OK" {
		t.Errorf("Label = %q, want %q", btn.Label, "OK")
	}
}
