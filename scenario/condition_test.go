package scenario

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func TestNewCondition_UnknownType(t *testing.T) {
	_, err := NewCondition(ConditionDef{Type: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !errors.Is(err, ErrUnknownConditionType) {
		t.Fatalf("expected ErrUnknownConditionType, got %v", err)
	}
}

func TestSurviveUntil_Evaluate(t *testing.T) {
	cond, err := NewCondition(ConditionDef{
		Type:   "survive_until",
		Params: json.RawMessage(`{"ticks": 100}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		snap GameSnapshot
		want bool
	}{
		{"before target tick", GameSnapshot{Tick: 50, CoreHP: 10}, false},
		{"at target tick alive", GameSnapshot{Tick: 100, CoreHP: 1}, true},
		{"past target tick alive", GameSnapshot{Tick: 200, CoreHP: 5}, true},
		{"at target tick dead", GameSnapshot{Tick: 100, CoreHP: 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cond.Evaluate(tt.snap); got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSurviveUntil_MissingParam(t *testing.T) {
	_, err := NewCondition(ConditionDef{
		Type:   "survive_until",
		Params: json.RawMessage(`{}`),
	})
	if err == nil {
		t.Fatal("expected error for missing ticks param")
	}
}

func TestDefeatAllWaves_Evaluate(t *testing.T) {
	cond, err := NewCondition(ConditionDef{Type: "defeat_all_waves"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		snap GameSnapshot
		want bool
	}{
		{"no waves", GameSnapshot{TotalWaves: 0, DefeatedWaves: 0}, false},
		{"partial", GameSnapshot{TotalWaves: 5, DefeatedWaves: 3}, false},
		{"all defeated", GameSnapshot{TotalWaves: 5, DefeatedWaves: 5}, true},
		{"over defeated", GameSnapshot{TotalWaves: 3, DefeatedWaves: 4}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cond.Evaluate(tt.snap); got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFengshuiScore_Evaluate(t *testing.T) {
	cond, err := NewCondition(ConditionDef{
		Type:   "fengshui_score",
		Params: json.RawMessage(`{"threshold": 80.0}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		snap GameSnapshot
		want bool
	}{
		{"below threshold", GameSnapshot{CaveFengShuiScore: 79.9}, false},
		{"at threshold", GameSnapshot{CaveFengShuiScore: 80.0}, true},
		{"above threshold", GameSnapshot{CaveFengShuiScore: 95.0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cond.Evaluate(tt.snap); got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChiPool_Evaluate(t *testing.T) {
	cond, err := NewCondition(ConditionDef{
		Type:   "chi_pool",
		Params: json.RawMessage(`{"threshold": 500.0}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		snap GameSnapshot
		want bool
	}{
		{"below", GameSnapshot{ChiPoolBalance: 499.9}, false},
		{"at", GameSnapshot{ChiPoolBalance: 500.0}, true},
		{"above", GameSnapshot{ChiPoolBalance: 1000.0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cond.Evaluate(tt.snap); got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoreDestroyed_Evaluate(t *testing.T) {
	cond, err := NewCondition(ConditionDef{Type: "core_destroyed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		snap GameSnapshot
		want bool
	}{
		{"alive", GameSnapshot{CoreHP: 10}, false},
		{"zero", GameSnapshot{CoreHP: 0}, true},
		{"negative", GameSnapshot{CoreHP: -5}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cond.Evaluate(tt.snap); got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllBeastsDefeated_Evaluate(t *testing.T) {
	cond, err := NewCondition(ConditionDef{Type: "all_beasts_defeated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		snap GameSnapshot
		want bool
	}{
		{"some alive", GameSnapshot{AliveBeasts: 3}, false},
		{"none alive", GameSnapshot{AliveBeasts: 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cond.Evaluate(tt.snap); got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBankrupt_Evaluate(t *testing.T) {
	cond, err := NewCondition(ConditionDef{
		Type:   "bankrupt",
		Params: json.RawMessage(`{"ticks": 10}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		snap GameSnapshot
		want bool
	}{
		{"below threshold", GameSnapshot{ConsecutiveDeficitTicks: 9}, false},
		{"at threshold", GameSnapshot{ConsecutiveDeficitTicks: 10}, true},
		{"above threshold", GameSnapshot{ConsecutiveDeficitTicks: 20}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cond.Evaluate(tt.snap); got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCondition_InvalidParamType(t *testing.T) {
	_, err := NewCondition(ConditionDef{
		Type:   "survive_until",
		Params: json.RawMessage(`{"ticks": "not a number"}`),
	})
	if err == nil {
		t.Fatal("expected error for non-numeric param")
	}
}

// Verify that types.Tick is used correctly in survive_until.
func TestSurviveUntil_TickType(t *testing.T) {
	cond, err := NewCondition(ConditionDef{
		Type:   "survive_until",
		Params: json.RawMessage(`{"ticks": 3000}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := GameSnapshot{Tick: types.Tick(3000), CoreHP: 1}
	if !cond.Evaluate(snap) {
		t.Error("expected true at exact tick boundary")
	}
}
