package view

import "testing"

func TestFormatChiPool(t *testing.T) {
	data := TopBarData{ChiPool: 150, MaxChiPool: 500}
	got := FormatChiPool(data)
	want := "Chi: 150/500"
	if got != want {
		t.Errorf("FormatChiPool = %q, want %q", got, want)
	}
}

func TestFormatChiPool_Zero(t *testing.T) {
	data := TopBarData{ChiPool: 0, MaxChiPool: 0}
	got := FormatChiPool(data)
	want := "Chi: 0/0"
	if got != want {
		t.Errorf("FormatChiPool = %q, want %q", got, want)
	}
}

func TestFormatCoreHP(t *testing.T) {
	data := TopBarData{CoreHP: 80, MaxCoreHP: 100}
	got := FormatCoreHP(data)
	want := "CoreHP: 80/100"
	if got != want {
		t.Errorf("FormatCoreHP = %q, want %q", got, want)
	}
}

func TestFormatTick(t *testing.T) {
	data := TopBarData{Tick: 42}
	got := FormatTick(data)
	want := "Tick: 42"
	if got != want {
		t.Errorf("FormatTick = %q, want %q", got, want)
	}
}

func TestBarWidth(t *testing.T) {
	tests := []struct {
		name       string
		current    int
		max        int
		totalWidth int
		want       int
	}{
		{name: "80 percent", current: 80, max: 100, totalWidth: 100, want: 80},
		{name: "full", current: 100, max: 100, totalWidth: 100, want: 100},
		{name: "zero", current: 0, max: 100, totalWidth: 100, want: 0},
		{name: "half of 200", current: 50, max: 100, totalWidth: 200, want: 100},
		{name: "max zero", current: 50, max: 0, totalWidth: 100, want: 0},
		{name: "over max", current: 120, max: 100, totalWidth: 100, want: 100},
		{name: "negative current", current: -10, max: 100, totalWidth: 100, want: 0},
		{name: "one third", current: 1, max: 3, totalWidth: 90, want: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BarWidth(tt.current, tt.max, tt.totalWidth)
			if got != tt.want {
				t.Errorf("BarWidth(%d, %d, %d) = %d, want %d",
					tt.current, tt.max, tt.totalWidth, got, tt.want)
			}
		})
	}
}
