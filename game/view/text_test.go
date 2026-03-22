package view

import "testing"

func TestTextWidth(t *testing.T) {
	tests := []struct {
		str  string
		want int
	}{
		{"", 0},
		{"A", 6},
		{"Hello", 30},
		{"Chi: 150/500", 72},
	}
	for _, tt := range tests {
		got := TextWidth(tt.str)
		if got != tt.want {
			t.Errorf("TextWidth(%q) = %d, want %d", tt.str, got, tt.want)
		}
	}
}
