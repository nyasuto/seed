package invasion

import (
	_ "embed"
	"testing"
)

//go:embed combat_params_data.json
var combatParamsJSON []byte

func TestDefaultCombatParams(t *testing.T) {
	p := DefaultCombatParams()
	if p.ATKMultiplier != 1.0 {
		t.Errorf("ATKMultiplier = %v, want 1.0", p.ATKMultiplier)
	}
	if p.DEFReduction != 0.5 {
		t.Errorf("DEFReduction = %v, want 0.5", p.DEFReduction)
	}
	if p.ElementAdvantage != 1.5 {
		t.Errorf("ElementAdvantage = %v, want 1.5", p.ElementAdvantage)
	}
	if p.ElementDisadvantage != 0.7 {
		t.Errorf("ElementDisadvantage = %v, want 0.7", p.ElementDisadvantage)
	}
	if p.MinDamage != 1 {
		t.Errorf("MinDamage = %v, want 1", p.MinDamage)
	}
	if p.CriticalChance != 0.1 {
		t.Errorf("CriticalChance = %v, want 0.1", p.CriticalChance)
	}
	if p.CriticalMultiplier != 2.0 {
		t.Errorf("CriticalMultiplier = %v, want 2.0", p.CriticalMultiplier)
	}
	if p.TrapDamageBase != 20 {
		t.Errorf("TrapDamageBase = %v, want 20", p.TrapDamageBase)
	}
	if p.TrapElementMultiplier != 1.3 {
		t.Errorf("TrapElementMultiplier = %v, want 1.3", p.TrapElementMultiplier)
	}
}

func TestLoadCombatParams(t *testing.T) {
	p, err := LoadCombatParams(combatParamsJSON)
	if err != nil {
		t.Fatalf("LoadCombatParams() error: %v", err)
	}

	def := DefaultCombatParams()
	if p != def {
		t.Errorf("loaded params differ from defaults:\n  got  %+v\n  want %+v", p, def)
	}
}

func TestLoadCombatParams_InvalidJSON(t *testing.T) {
	_, err := LoadCombatParams([]byte("{invalid"))
	if err == nil {
		t.Error("LoadCombatParams() with invalid JSON should return error")
	}
}
