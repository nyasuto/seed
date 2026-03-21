package economy

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestDefaultBeastCost(t *testing.T) {
	bc := DefaultBeastCost()
	if bc == nil {
		t.Fatal("DefaultBeastCost() returned nil")
	}
	if len(bc.SummonCostByElement) != types.ElementCount {
		t.Errorf("expected %d elements, got %d", types.ElementCount, len(bc.SummonCostByElement))
	}
}

func TestCalcSummonCost_ByElement(t *testing.T) {
	bc := DefaultBeastCost()

	tests := []struct {
		element types.Element
		want    float64
	}{
		{types.Wood, 30.0},
		{types.Fire, 35.0},
		{types.Earth, 25.0},
		{types.Metal, 40.0},
		{types.Water, 30.0},
	}

	for _, tt := range tests {
		t.Run(tt.element.String(), func(t *testing.T) {
			got := bc.CalcSummonCost(tt.element)
			if got != tt.want {
				t.Errorf("CalcSummonCost(%v) = %v, want %v", tt.element, got, tt.want)
			}
		})
	}
}

func TestCalcSummonCost_UnknownElement(t *testing.T) {
	bc := &BeastCost{
		SummonCostByElement: map[types.Element]float64{},
	}
	got := bc.CalcSummonCost(types.Wood)
	if got != 0 {
		t.Errorf("CalcSummonCost for missing element = %v, want 0", got)
	}
}

func TestLoadBeastCost_InvalidJSON(t *testing.T) {
	_, err := LoadBeastCost([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadBeastCost_UnknownElement(t *testing.T) {
	data := []byte(`{"summon_cost_by_element": {"Void": 10.0}}`)
	_, err := LoadBeastCost(data)
	if err == nil {
		t.Error("expected error for unknown element")
	}
}
