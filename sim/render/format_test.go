package render

import (
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func TestElementColor_AllElements(t *testing.T) {
	tests := []struct {
		elem types.Element
		want string
	}{
		{types.Fire, Red},
		{types.Water, Blue},
		{types.Wood, Green},
		{types.Metal, Yellow},
		{types.Earth, Brown},
	}
	for _, tt := range tests {
		t.Run(tt.elem.String(), func(t *testing.T) {
			got := ElementColor(tt.elem)
			if got != tt.want {
				t.Errorf("ElementColor(%v) = %q, want %q", tt.elem, got, tt.want)
			}
		})
	}
}

func TestColorize_WrapsText(t *testing.T) {
	result := Colorize("hello", Red)
	if !strings.HasPrefix(result, Red) {
		t.Errorf("should start with color code")
	}
	if !strings.HasSuffix(result, Reset) {
		t.Errorf("should end with reset code")
	}
	if !strings.Contains(result, "hello") {
		t.Errorf("should contain original text")
	}
}

func TestHPBar_FullHP(t *testing.T) {
	bar := HPBar(100, 100, 10)
	stripped := StripANSI(bar)
	if stripped != "[██████████]" {
		t.Errorf("full HP bar = %q, want full bar", stripped)
	}
	// Full HP should be green.
	if !strings.Contains(bar, Green) {
		t.Errorf("full HP bar should contain green color")
	}
}

func TestHPBar_LowHP(t *testing.T) {
	bar := HPBar(10, 100, 10)
	// 10% HP should be red.
	if !strings.Contains(bar, Red) {
		t.Errorf("low HP bar should contain red color")
	}
}

func TestHPBar_MidHP(t *testing.T) {
	bar := HPBar(40, 100, 10)
	// 40% HP should be yellow.
	if !strings.Contains(bar, Yellow) {
		t.Errorf("mid HP bar should contain yellow color")
	}
}

func TestHPBar_ZeroMax(t *testing.T) {
	bar := HPBar(0, 0, 10)
	stripped := StripANSI(bar)
	if stripped != "[░░░░░░░░░░]" {
		t.Errorf("zero max HP bar = %q, want empty bar", stripped)
	}
}

func TestProgressBar_Ratio(t *testing.T) {
	tests := []struct {
		name  string
		ratio float64
		want  int // expected filled count in width=10
	}{
		{"empty", 0.0, 0},
		{"half", 0.5, 5},
		{"full", 1.0, 10},
		{"over", 1.5, 10},
		{"negative", -0.5, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := ProgressBar(tt.ratio, 10)
			filled := strings.Count(bar, "█")
			if filled != tt.want {
				t.Errorf("ProgressBar(%v, 10) filled=%d, want %d", tt.ratio, filled, tt.want)
			}
		})
	}
}

func TestFormatHP_Colors(t *testing.T) {
	tests := []struct {
		current int
		max     int
		color   string
	}{
		{100, 100, Green},
		{50, 100, Yellow},
		{20, 100, Red},
	}
	for _, tt := range tests {
		result := FormatHP(tt.current, tt.max)
		if !strings.Contains(result, tt.color) {
			t.Errorf("FormatHP(%d, %d) should contain color %q", tt.current, tt.max, tt.color)
		}
	}
}

func TestStripANSI_RemovesEscapes(t *testing.T) {
	colored := Colorize("hello", Red)
	stripped := StripANSI(colored)
	if stripped != "hello" {
		t.Errorf("StripANSI = %q, want %q", stripped, "hello")
	}
}

func TestStripANSI_PlainText(t *testing.T) {
	plain := "hello world"
	if StripANSI(plain) != plain {
		t.Errorf("StripANSI should not modify plain text")
	}
}

func TestVisibleWidth_IgnoresANSI(t *testing.T) {
	colored := Colorize("hello", Red)
	if VisibleWidth(colored) != 5 {
		t.Errorf("VisibleWidth = %d, want 5", VisibleWidth(colored))
	}
}
