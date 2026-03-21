package invasion

import (
	"math"
	"sort"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
)

// CombatRoundResult holds the outcome of a single combat round
// between one beast and one invader.
type CombatRoundResult struct {
	// BeastDamageTaken is the damage dealt to the beast this round.
	BeastDamageTaken int
	// InvaderDamageTaken is the damage dealt to the invader this round.
	InvaderDamageTaken int
	// BeastHP is the beast's remaining HP after this round.
	BeastHP int
	// InvaderHP is the invader's remaining HP after this round.
	InvaderHP int
	// IsBeastDefeated is true if the beast's HP reached zero.
	IsBeastDefeated bool
	// IsInvaderDefeated is true if the invader's HP reached zero.
	IsInvaderDefeated bool
	// WasBeastCritical is true if the beast landed a critical hit.
	WasBeastCritical bool
	// WasInvaderCritical is true if the invader landed a critical hit.
	WasInvaderCritical bool
	// FirstAttacker is "beast" or "invader" indicating who attacked first.
	FirstAttacker string
}

// CombatEngine resolves combat between beasts and invaders.
type CombatEngine struct {
	params CombatParams
	rng    types.RNG
}

// NewCombatEngine creates a new CombatEngine with the given parameters and RNG.
func NewCombatEngine(params CombatParams, rng types.RNG) *CombatEngine {
	return &CombatEngine{
		params: params,
		rng:    rng,
	}
}

// calcDamage computes damage from an attacker to a defender.
// Returns the damage amount and whether it was a critical hit.
func (ce *CombatEngine) calcDamage(attackerATK int, defenderDEF int, attackerElement, defenderElement types.Element) (int, bool) {
	// Base damage: ATK * ATKMultiplier - DEF * DEFReduction
	raw := float64(attackerATK)*ce.params.ATKMultiplier - float64(defenderDEF)*ce.params.DEFReduction

	// Element modifier
	elemMult := 1.0
	if types.Overcomes(attackerElement, defenderElement) {
		elemMult = ce.params.ElementAdvantage
	} else if types.Overcomes(defenderElement, attackerElement) {
		elemMult = ce.params.ElementDisadvantage
	}
	raw *= elemMult

	// Critical hit
	isCrit := ce.rng.Float64() < ce.params.CriticalChance
	if isCrit {
		raw *= ce.params.CriticalMultiplier
	}

	dmg := int(math.Round(raw))
	if dmg < ce.params.MinDamage {
		dmg = ce.params.MinDamage
	}
	return dmg, isCrit
}

// ResolveCombatRound resolves one round of combat between a beast and an invader.
// The beast's effective stats are computed from room chi affinity.
func (ce *CombatEngine) ResolveCombatRound(beast *senju.Beast, invader *Invader, roomChi *fengshui.RoomChi) CombatRoundResult {
	stats := beast.CalcCombatStats(roomChi)

	// Determine first attacker by speed; ties broken by RNG
	beastFirst := stats.SPD > invader.SPD
	if stats.SPD == invader.SPD {
		beastFirst = ce.rng.Intn(2) == 0
	}

	result := CombatRoundResult{}
	if beastFirst {
		result.FirstAttacker = "beast"
	} else {
		result.FirstAttacker = "invader"
	}

	if beastFirst {
		// Beast attacks first
		dmg, crit := ce.calcDamage(stats.ATK, invader.DEF, beast.Element, invader.Element)
		result.WasBeastCritical = crit
		result.InvaderDamageTaken = dmg
		invader.HP -= dmg
		if invader.HP <= 0 {
			invader.HP = 0
			result.InvaderHP = 0
			result.BeastHP = beast.HP
			result.IsInvaderDefeated = true
			return result
		}

		// Invader counterattacks
		dmg, crit = ce.calcDamage(invader.ATK, stats.DEF, invader.Element, beast.Element)
		result.WasInvaderCritical = crit
		result.BeastDamageTaken = dmg
		beast.HP -= dmg
		if beast.HP <= 0 {
			beast.HP = 0
		}
	} else {
		// Invader attacks first
		dmg, crit := ce.calcDamage(invader.ATK, stats.DEF, invader.Element, beast.Element)
		result.WasInvaderCritical = crit
		result.BeastDamageTaken = dmg
		beast.HP -= dmg
		if beast.HP <= 0 {
			beast.HP = 0
			result.BeastHP = 0
			result.InvaderHP = invader.HP
			result.IsBeastDefeated = true
			return result
		}

		// Beast counterattacks
		dmg, crit = ce.calcDamage(stats.ATK, invader.DEF, beast.Element, invader.Element)
		result.WasBeastCritical = crit
		result.InvaderDamageTaken = dmg
		invader.HP -= dmg
		if invader.HP <= 0 {
			invader.HP = 0
		}
	}

	result.BeastHP = beast.HP
	result.InvaderHP = invader.HP
	result.IsBeastDefeated = beast.HP <= 0
	result.IsInvaderDefeated = invader.HP <= 0
	return result
}

// ResolveRoomCombat resolves one round of combat for all beasts vs all invaders
// in the same room. Combatants are paired by descending speed; unpaired units
// do not fight this round.
func (ce *CombatEngine) ResolveRoomCombat(beasts []*senju.Beast, invaders []*Invader, roomChi *fengshui.RoomChi) []CombatRoundResult {
	if len(beasts) == 0 || len(invaders) == 0 {
		return nil
	}

	// Sort beasts by effective SPD descending
	type beastWithSPD struct {
		beast *senju.Beast
		spd   int
	}
	beastsSorted := make([]beastWithSPD, len(beasts))
	for i, b := range beasts {
		stats := b.CalcCombatStats(roomChi)
		beastsSorted[i] = beastWithSPD{beast: b, spd: stats.SPD}
	}
	sort.SliceStable(beastsSorted, func(i, j int) bool {
		return beastsSorted[i].spd > beastsSorted[j].spd
	})

	// Sort invaders by SPD descending
	invadersSorted := make([]*Invader, len(invaders))
	copy(invadersSorted, invaders)
	sort.SliceStable(invadersSorted, func(i, j int) bool {
		return invadersSorted[i].SPD > invadersSorted[j].SPD
	})

	// Pair up by index; min(len, len) pairs
	pairs := len(beastsSorted)
	if len(invadersSorted) < pairs {
		pairs = len(invadersSorted)
	}

	results := make([]CombatRoundResult, 0, pairs)
	for i := 0; i < pairs; i++ {
		r := ce.ResolveCombatRound(beastsSorted[i].beast, invadersSorted[i], roomChi)
		results = append(results, r)
	}
	return results
}
