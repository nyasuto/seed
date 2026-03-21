package invasion

import (
	"testing"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/testutil"
	"github.com/nyasuto/seed/core/types"
)

// makeBeast creates a beast with specific stats for testing.
func makeBeast(element types.Element, hp, atk, def, spd int) *senju.Beast {
	return &senju.Beast{
		ID:      1,
		Element: element,
		HP:      hp,
		MaxHP:   hp,
		ATK:     atk,
		DEF:     def,
		SPD:     spd,
	}
}

// makeInvader creates an invader with specific stats for testing.
func makeInvader(element types.Element, hp, atk, def, spd int) *Invader {
	return &Invader{
		ID:      1,
		Element: element,
		HP:      hp,
		MaxHP:   hp,
		ATK:     atk,
		DEF:     def,
		SPD:     spd,
		State:   Fighting,
	}
}

func TestResolveCombatRound_BasicDamage(t *testing.T) {
	// No crits (FloatValue=0.5 > CriticalChance=0.1)
	rng := &testutil.FixedRNG{IntValue: 0, FloatValue: 0.5}
	params := DefaultCombatParams()
	engine := NewCombatEngine(params, rng)

	// Beast: ATK=20, Wood; Invader: ATK=15, DEF=10, SPD lower
	beast := makeBeast(types.Wood, 100, 20, 10, 15)
	invader := makeInvader(types.Wood, 80, 15, 10, 10)

	// Beast faster (15>10), so beast attacks first
	// Beast damage: 20*1.0 - 10*0.5 = 15 (same element, no advantage)
	// Invader damage: 15*1.0 - 10*0.5 = 10
	result := engine.ResolveCombatRound(beast, invader, nil)

	if result.FirstAttacker != "beast" {
		t.Errorf("expected beast first, got %s", result.FirstAttacker)
	}
	if result.InvaderDamageTaken != 15 {
		t.Errorf("expected invader damage 15, got %d", result.InvaderDamageTaken)
	}
	if result.BeastDamageTaken != 10 {
		t.Errorf("expected beast damage 10, got %d", result.BeastDamageTaken)
	}
	if result.InvaderHP != 65 {
		t.Errorf("expected invader HP 65, got %d", result.InvaderHP)
	}
	if result.BeastHP != 90 {
		t.Errorf("expected beast HP 90, got %d", result.BeastHP)
	}
}

func TestResolveCombatRound_FirstAttacker(t *testing.T) {
	tests := []struct {
		name          string
		beastSPD      int
		invaderSPD    int
		rngInt        int
		wantFirst     string
	}{
		{"beast faster", 20, 10, 0, "beast"},
		{"invader faster", 10, 20, 0, "invader"},
		{"tie, rng=0 => beast", 15, 15, 0, "beast"},
		{"tie, rng=1 => invader", 15, 15, 1, "invader"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := &testutil.FixedRNG{IntValue: tt.rngInt, FloatValue: 0.5}
			engine := NewCombatEngine(DefaultCombatParams(), rng)
			beast := makeBeast(types.Wood, 100, 10, 10, tt.beastSPD)
			invader := makeInvader(types.Wood, 100, 10, 10, tt.invaderSPD)

			result := engine.ResolveCombatRound(beast, invader, nil)
			if result.FirstAttacker != tt.wantFirst {
				t.Errorf("got FirstAttacker=%s, want %s", result.FirstAttacker, tt.wantFirst)
			}
		})
	}
}

func TestResolveCombatRound_ElementAdvantage(t *testing.T) {
	// Wood overcomes Earth
	rng := &testutil.FixedRNG{IntValue: 0, FloatValue: 0.5}
	params := DefaultCombatParams()
	engine := NewCombatEngine(params, rng)

	beast := makeBeast(types.Wood, 100, 20, 10, 20)
	invader := makeInvader(types.Earth, 100, 20, 10, 10)

	// Beast attacks: 20*1.0 - 10*0.5 = 15, * 1.5 (advantage) = 22.5 -> 23 (rounded)
	result := engine.ResolveCombatRound(beast, invader, nil)
	if result.InvaderDamageTaken != 23 {
		t.Errorf("expected invader damage 23 (element advantage), got %d", result.InvaderDamageTaken)
	}
}

func TestResolveCombatRound_ElementDisadvantage(t *testing.T) {
	// Earth is overcome by Wood, so Earth attacking Wood gets disadvantage
	rng := &testutil.FixedRNG{IntValue: 0, FloatValue: 0.5}
	params := DefaultCombatParams()
	engine := NewCombatEngine(params, rng)

	beast := makeBeast(types.Earth, 100, 20, 10, 20)
	invader := makeInvader(types.Wood, 100, 20, 10, 10)

	// Beast attacks: 20*1.0 - 10*0.5 = 15, * 0.7 (disadvantage) = 10.5 -> 11 (rounded)
	result := engine.ResolveCombatRound(beast, invader, nil)
	if result.InvaderDamageTaken != 11 {
		t.Errorf("expected invader damage 11 (element disadvantage), got %d", result.InvaderDamageTaken)
	}
}

func TestResolveCombatRound_Critical(t *testing.T) {
	// FloatValue=0.05 < CriticalChance=0.1, so crit fires
	rng := &testutil.FixedRNG{IntValue: 0, FloatValue: 0.05}
	params := DefaultCombatParams()
	engine := NewCombatEngine(params, rng)

	beast := makeBeast(types.Wood, 100, 20, 10, 20)
	invader := makeInvader(types.Wood, 100, 20, 10, 10)

	// Beast attacks: 15 * 2.0 (crit) = 30
	result := engine.ResolveCombatRound(beast, invader, nil)
	if !result.WasBeastCritical {
		t.Error("expected beast critical hit")
	}
	if result.InvaderDamageTaken != 30 {
		t.Errorf("expected invader damage 30 (critical), got %d", result.InvaderDamageTaken)
	}
	// Invader also crits (same FixedRNG)
	if !result.WasInvaderCritical {
		t.Error("expected invader critical hit")
	}
	if result.BeastDamageTaken != 30 {
		t.Errorf("expected beast damage 30 (critical), got %d", result.BeastDamageTaken)
	}
}

func TestResolveCombatRound_MinDamage(t *testing.T) {
	// Very high DEF to make raw damage negative
	rng := &testutil.FixedRNG{IntValue: 0, FloatValue: 0.5}
	engine := NewCombatEngine(DefaultCombatParams(), rng)

	beast := makeBeast(types.Wood, 100, 5, 100, 20)
	invader := makeInvader(types.Wood, 100, 5, 100, 10)

	// 5*1.0 - 100*0.5 = -45 -> MinDamage = 1
	result := engine.ResolveCombatRound(beast, invader, nil)
	if result.InvaderDamageTaken != 1 {
		t.Errorf("expected min damage 1, got %d", result.InvaderDamageTaken)
	}
	if result.BeastDamageTaken != 1 {
		t.Errorf("expected min damage 1, got %d", result.BeastDamageTaken)
	}
}

func TestResolveCombatRound_FirstAttackerKills(t *testing.T) {
	rng := &testutil.FixedRNG{IntValue: 0, FloatValue: 0.5}
	engine := NewCombatEngine(DefaultCombatParams(), rng)

	// Beast has very high ATK, invader has low HP
	beast := makeBeast(types.Wood, 100, 50, 10, 20)
	invader := makeInvader(types.Wood, 10, 50, 10, 10)

	// Beast damage: 50 - 5 = 45, invader HP=10 -> defeated
	result := engine.ResolveCombatRound(beast, invader, nil)
	if !result.IsInvaderDefeated {
		t.Error("expected invader defeated")
	}
	if result.BeastDamageTaken != 0 {
		t.Errorf("expected no beast damage (invader died first), got %d", result.BeastDamageTaken)
	}
	if invader.HP != 0 {
		t.Errorf("expected invader HP 0, got %d", invader.HP)
	}
}

func TestResolveRoomCombat_MultipleMatchups(t *testing.T) {
	rng := &testutil.FixedRNG{IntValue: 0, FloatValue: 0.5}
	engine := NewCombatEngine(DefaultCombatParams(), rng)

	// 2 beasts vs 3 invaders -> 2 pairs
	b1 := makeBeast(types.Wood, 100, 20, 10, 30)
	b1.ID = 1
	b2 := makeBeast(types.Fire, 100, 20, 10, 10)
	b2.ID = 2

	i1 := makeInvader(types.Water, 100, 15, 10, 25)
	i1.ID = 1
	i2 := makeInvader(types.Earth, 100, 15, 10, 20)
	i2.ID = 2
	i3 := makeInvader(types.Metal, 100, 15, 10, 5)
	i3.ID = 3

	results := engine.ResolveRoomCombat(
		[]*senju.Beast{b1, b2},
		[]*Invader{i1, i2, i3},
		nil,
	)

	// Should have 2 results (min of 2 beasts, 3 invaders)
	if len(results) != 2 {
		t.Fatalf("expected 2 combat results, got %d", len(results))
	}

	// Fastest beast (b1, SPD=30) pairs with fastest invader (i1, SPD=25)
	// Second pair: b2 (SPD=10) vs i2 (SPD=20)
	// i3 (SPD=5) is unpaired

	// Verify that combat happened (HP changed)
	if b1.HP == 100 && i1.HP == 100 {
		t.Error("expected combat to change HP for first pair")
	}
	if b2.HP == 100 && i2.HP == 100 {
		t.Error("expected combat to change HP for second pair")
	}
	// Third invader should be untouched
	if i3.HP != 100 {
		t.Errorf("expected unpaired invader HP unchanged, got %d", i3.HP)
	}
}

func TestResolveRoomCombat_Empty(t *testing.T) {
	rng := &testutil.FixedRNG{IntValue: 0, FloatValue: 0.5}
	engine := NewCombatEngine(DefaultCombatParams(), rng)

	results := engine.ResolveRoomCombat(nil, nil, nil)
	if results != nil {
		t.Errorf("expected nil for empty combatants, got %v", results)
	}

	b := makeBeast(types.Wood, 100, 10, 10, 10)
	results = engine.ResolveRoomCombat([]*senju.Beast{b}, nil, nil)
	if results != nil {
		t.Errorf("expected nil for no invaders, got %v", results)
	}
}

func TestResolveCombatRound_WithRoomChi(t *testing.T) {
	rng := &testutil.FixedRNG{IntValue: 0, FloatValue: 0.5}
	engine := NewCombatEngine(DefaultCombatParams(), rng)

	// Water room generates Wood beast -> 1.3x multiplier on ATK/DEF/SPD
	beast := makeBeast(types.Wood, 100, 20, 10, 20)
	invader := makeInvader(types.Wood, 100, 20, 10, 10)
	roomChi := &fengshui.RoomChi{
		RoomID:   1,
		Current:  50,
		Capacity: 100,
		Element:  types.Water, // Water generates Wood
	}

	// Beast effective ATK = round(20*1.3) = 26
	// Beast effective DEF = round(10*1.3) = 13
	// Beast damage to invader: 26*1.0 - 10*0.5 = 21
	// Invader damage to beast: 20*1.0 - 13*0.5 = 13.5 -> 14 (rounded)
	result := engine.ResolveCombatRound(beast, invader, roomChi)
	if result.InvaderDamageTaken != 21 {
		t.Errorf("expected invader damage 21 (room affinity), got %d", result.InvaderDamageTaken)
	}
	if result.BeastDamageTaken != 14 {
		t.Errorf("expected beast damage 14 (room affinity DEF), got %d", result.BeastDamageTaken)
	}
}
