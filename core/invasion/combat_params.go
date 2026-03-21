package invasion

import (
	"encoding/json"
	"fmt"
)

// CombatParams holds tunable parameters for the combat system.
type CombatParams struct {
	// ATKMultiplier scales the attacker's ATK stat.
	ATKMultiplier float64
	// DEFReduction scales the defender's DEF subtraction.
	DEFReduction float64
	// ElementAdvantage is the damage multiplier when the attacker's element
	// overcomes the defender's element (相克).
	ElementAdvantage float64
	// ElementDisadvantage is the damage multiplier when the attacker's element
	// is overcome by the defender's element.
	ElementDisadvantage float64
	// MinDamage is the minimum damage dealt per hit.
	MinDamage int
	// CriticalChance is the probability of a critical hit (0.0–1.0).
	CriticalChance float64
	// CriticalMultiplier scales damage on a critical hit.
	CriticalMultiplier float64
	// TrapDamageBase is the base damage dealt by traps.
	TrapDamageBase int
	// TrapElementMultiplier scales trap damage when the trap's element
	// overcomes the target's element.
	TrapElementMultiplier float64
}

// DefaultCombatParams returns the default combat parameters.
func DefaultCombatParams() CombatParams {
	return CombatParams{
		ATKMultiplier:         1.0,
		DEFReduction:          0.5,
		ElementAdvantage:      1.5,
		ElementDisadvantage:   0.7,
		MinDamage:             1,
		CriticalChance:        0.1,
		CriticalMultiplier:    2.0,
		TrapDamageBase:        20,
		TrapElementMultiplier: 1.3,
	}
}

// LoadCombatParams parses JSON data into CombatParams.
func LoadCombatParams(data []byte) (CombatParams, error) {
	var p CombatParams
	if err := json.Unmarshal(data, &p); err != nil {
		return CombatParams{}, fmt.Errorf("loading combat params: %w", err)
	}
	return p, nil
}
