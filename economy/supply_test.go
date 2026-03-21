package economy

import (
	"math"
	"testing"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/types"
)

func defaultTestSupplyParams() *SupplyParams {
	return &SupplyParams{
		BaseSupplyPerVein:     5.0,
		FengShuiMinMultiplier: 0.8,
		FengShuiMaxMultiplier: 1.3,
		ChiRatioSupplyWeight:  0.5,
	}
}

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestCalcTickSupply_NoVeins(t *testing.T) {
	sc := NewSupplyCalculator(defaultTestSupplyParams())
	got := sc.CalcTickSupply(nil, nil, 0.5)
	if got != 0 {
		t.Errorf("expected 0 supply with no veins, got %f", got)
	}
}

func TestCalcTickSupply_MultipleVeins(t *testing.T) {
	sc := NewSupplyCalculator(defaultTestSupplyParams())
	veins := []fengshui.DragonVein{
		{ID: 1, SourcePos: types.Pos{X: 0, Y: 0}, Element: types.Wood, FlowRate: 1.0},
		{ID: 2, SourcePos: types.Pos{X: 1, Y: 1}, Element: types.Fire, FlowRate: 1.0},
		{ID: 3, SourcePos: types.Pos{X: 2, Y: 2}, Element: types.Water, FlowRate: 1.0},
	}
	// No rooms, caveScore=0.5 → midpoint multiplier
	// base = 3 * 5.0 = 15.0
	// fillBonus = 0 (no rooms)
	// fengShuiMul = 0.8 + 0.5*(1.3-0.8) = 1.05
	// total = 15.0 * 1.05 = 15.75
	got := sc.CalcTickSupply(veins, nil, 0.5)
	if !almostEqual(got, 15.75, 0.001) {
		t.Errorf("expected 15.75, got %f", got)
	}
}

func TestCalcTickSupply_FengShuiBonusAndPenalty(t *testing.T) {
	sc := NewSupplyCalculator(defaultTestSupplyParams())
	veins := []fengshui.DragonVein{
		{ID: 1, SourcePos: types.Pos{X: 0, Y: 0}, Element: types.Wood, FlowRate: 1.0},
	}

	tests := []struct {
		name      string
		caveScore float64
		want      float64
	}{
		{
			name:      "lowest feng shui score gives penalty",
			caveScore: 0.0,
			// base=5.0, fillBonus=0, mul=0.8 → 5.0*0.8=4.0
			want: 4.0,
		},
		{
			name:      "highest feng shui score gives bonus",
			caveScore: 1.0,
			// base=5.0, fillBonus=0, mul=1.3 → 5.0*1.3=6.5
			want: 6.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sc.CalcTickSupply(veins, nil, tt.caveScore)
			if !almostEqual(got, tt.want, 0.001) {
				t.Errorf("caveScore=%f: expected %f, got %f", tt.caveScore, tt.want, got)
			}
		})
	}
}

func TestCalcTickSupply_LowChiFillRatio(t *testing.T) {
	sc := NewSupplyCalculator(defaultTestSupplyParams())
	veins := []fengshui.DragonVein{
		{ID: 1, SourcePos: types.Pos{X: 0, Y: 0}, Element: types.Wood, FlowRate: 1.0},
	}

	fullRooms := map[int]*fengshui.RoomChi{
		1: {RoomID: 1, Current: 100, Capacity: 100, Element: types.Wood},
	}
	emptyRooms := map[int]*fengshui.RoomChi{
		1: {RoomID: 1, Current: 0, Capacity: 100, Element: types.Wood},
	}

	caveScore := 0.5
	// fengShuiMul = 0.8 + 0.5*0.5 = 1.05

	supplyFull := sc.CalcTickSupply(veins, fullRooms, caveScore)
	// base=5.0, fillBonus=1.0*0.5=0.5, total=(5.0+0.5)*1.05=5.775
	if !almostEqual(supplyFull, 5.775, 0.001) {
		t.Errorf("full rooms: expected 5.775, got %f", supplyFull)
	}

	supplyEmpty := sc.CalcTickSupply(veins, emptyRooms, caveScore)
	// base=5.0, fillBonus=0*0.5=0, total=5.0*1.05=5.25
	if !almostEqual(supplyEmpty, 5.25, 0.001) {
		t.Errorf("empty rooms: expected 5.25, got %f", supplyEmpty)
	}

	if supplyEmpty >= supplyFull {
		t.Errorf("empty rooms supply (%f) should be less than full rooms supply (%f)", supplyEmpty, supplyFull)
	}
}

func TestDefaultSupplyParams(t *testing.T) {
	p := DefaultSupplyParams()
	if p == nil {
		t.Fatal("DefaultSupplyParams returned nil")
	}
	if p.BaseSupplyPerVein != 5.0 {
		t.Errorf("BaseSupplyPerVein: expected 5.0, got %f", p.BaseSupplyPerVein)
	}
	if p.FengShuiMinMultiplier != 0.8 {
		t.Errorf("FengShuiMinMultiplier: expected 0.8, got %f", p.FengShuiMinMultiplier)
	}
}

func TestLoadSupplyParams_InvalidJSON(t *testing.T) {
	_, err := LoadSupplyParams([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
