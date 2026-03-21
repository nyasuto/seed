package invasion

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

func setupTrapTestCave(t *testing.T) (*world.Cave, *world.RoomTypeRegistry) {
	t.Helper()
	cave, _ := world.NewCave(16, 16)

	reg := world.NewRoomTypeRegistry()
	if err := reg.Register(world.RoomType{
		ID:              "trap_room",
		Name:            "罠部屋",
		Element:         types.Metal,
		BaseChiCapacity: 30,
		Description:     "trap room",
		MaxBeasts:       0,
	}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(world.RoomType{
		ID:              "recovery_room",
		Name:            "回復室",
		Element:         types.Water,
		BaseChiCapacity: 50,
		Description:     "recovery room",
		MaxBeasts:       2,
	}); err != nil {
		t.Fatal(err)
	}
	return cave, reg
}

func TestBuildTrapEffects_TrapRoomsOnly(t *testing.T) {
	_, reg := setupTrapTestCave(t)

	rooms := []world.Room{
		{ID: 1, TypeID: "trap_room"},
		{ID: 2, TypeID: "recovery_room"},
		{ID: 3, TypeID: "trap_room"},
	}

	effects := BuildTrapEffects(nil, rooms, reg)
	if len(effects) != 2 {
		t.Fatalf("expected 2 trap effects, got %d", len(effects))
	}
	if effects[0].RoomID != 1 {
		t.Errorf("expected first trap RoomID=1, got %d", effects[0].RoomID)
	}
	if effects[1].RoomID != 3 {
		t.Errorf("expected second trap RoomID=3, got %d", effects[1].RoomID)
	}
	for _, e := range effects {
		if e.Element != types.Metal {
			t.Errorf("expected trap element Metal, got %v", e.Element)
		}
		if e.SlowTicks != 2 {
			t.Errorf("expected SlowTicks=2, got %d", e.SlowTicks)
		}
	}
}

func TestBuildTrapEffects_NoTrapRooms(t *testing.T) {
	_, reg := setupTrapTestCave(t)

	rooms := []world.Room{
		{ID: 1, TypeID: "recovery_room"},
	}

	effects := BuildTrapEffects(nil, rooms, reg)
	if len(effects) != 0 {
		t.Fatalf("expected 0 trap effects, got %d", len(effects))
	}
}

func TestApplyTrap_BaseDamage(t *testing.T) {
	params := DefaultCombatParams()
	// Metal trap vs Wood invader: no element advantage (Metal does not overcome Wood)
	invader := &Invader{
		ID:      1,
		Element: types.Wood,
		HP:      100,
		MaxHP:   100,
	}
	trap := TrapEffect{
		RoomID:           1,
		Element:          types.Metal,
		DamagePerTrigger: 30,
		SlowTicks:        2,
	}

	result := ApplyTrap(invader, trap, params)

	// Metal overcomes Wood: types.Overcomes(Metal, Wood) == true
	// So damage = 20 * 1.3 = 26
	if types.Overcomes(types.Metal, types.Wood) {
		expected := int(float64(params.TrapDamageBase) * params.TrapElementMultiplier)
		if result.Damage != expected {
			t.Errorf("expected damage %d, got %d", expected, result.Damage)
		}
	} else {
		if result.Damage != params.TrapDamageBase {
			t.Errorf("expected base damage %d, got %d", params.TrapDamageBase, result.Damage)
		}
	}

	if invader.HP != 100-result.Damage {
		t.Errorf("expected HP %d, got %d", 100-result.Damage, invader.HP)
	}
}

func TestApplyTrap_ElementAdvantage(t *testing.T) {
	params := DefaultCombatParams()
	// Wood overcomes Earth (相克: 木→土)
	invader := &Invader{
		ID:      1,
		Element: types.Earth,
		HP:      100,
		MaxHP:   100,
	}
	trap := TrapEffect{
		RoomID:           1,
		Element:          types.Wood,
		DamagePerTrigger: 30,
		SlowTicks:        2,
	}

	result := ApplyTrap(invader, trap, params)

	expectedDmg := int(float64(params.TrapDamageBase) * params.TrapElementMultiplier)
	if result.Damage != expectedDmg {
		t.Errorf("expected element advantage damage %d, got %d", expectedDmg, result.Damage)
	}
	if invader.HP != 100-expectedDmg {
		t.Errorf("expected HP %d, got %d", 100-expectedDmg, invader.HP)
	}
}

func TestApplyTrap_NoElementAdvantage(t *testing.T) {
	params := DefaultCombatParams()
	// Fire vs Water: Water overcomes Fire, but Fire does NOT overcome Water
	invader := &Invader{
		ID:      1,
		Element: types.Water,
		HP:      100,
		MaxHP:   100,
	}
	trap := TrapEffect{
		RoomID:           1,
		Element:          types.Fire,
		DamagePerTrigger: 30,
		SlowTicks:        2,
	}

	result := ApplyTrap(invader, trap, params)

	if result.Damage != params.TrapDamageBase {
		t.Errorf("expected base damage %d (no element advantage), got %d", params.TrapDamageBase, result.Damage)
	}
}

func TestApplyTrap_SlowEffect(t *testing.T) {
	params := DefaultCombatParams()
	invader := &Invader{
		ID:        1,
		Element:   types.Fire,
		HP:        100,
		MaxHP:     100,
		SlowTicks: 0,
	}
	trap := TrapEffect{
		RoomID:    1,
		Element:   types.Earth,
		SlowTicks: 3,
	}

	result := ApplyTrap(invader, trap, params)

	if !result.IsSlowed {
		t.Error("expected IsSlowed=true")
	}
	if result.SlowTicksApplied != 3 {
		t.Errorf("expected SlowTicksApplied=3, got %d", result.SlowTicksApplied)
	}
	if invader.SlowTicks != 3 {
		t.Errorf("expected invader SlowTicks=3, got %d", invader.SlowTicks)
	}

	// Apply again — slow ticks should stack
	result2 := ApplyTrap(invader, trap, params)
	if invader.SlowTicks != 6 {
		t.Errorf("expected invader SlowTicks=6 after second trap, got %d", invader.SlowTicks)
	}
	_ = result2
}

func TestApplyTrap_KillsInvader(t *testing.T) {
	params := DefaultCombatParams()
	invader := &Invader{
		ID:      1,
		Element: types.Fire,
		HP:      5,
		MaxHP:   100,
	}
	trap := TrapEffect{
		RoomID:    1,
		Element:   types.Water, // Water overcomes Fire
		SlowTicks: 2,
	}

	result := ApplyTrap(invader, trap, params)

	if invader.HP != 0 {
		t.Errorf("expected HP=0, got %d", invader.HP)
	}
	expectedDmg := int(float64(params.TrapDamageBase) * params.TrapElementMultiplier)
	if result.Damage != expectedDmg {
		t.Errorf("expected damage %d, got %d", expectedDmg, result.Damage)
	}
}

func TestApplyTrap_ZeroSlowTicks(t *testing.T) {
	params := DefaultCombatParams()
	invader := &Invader{
		ID:      1,
		Element: types.Fire,
		HP:      100,
		MaxHP:   100,
	}
	trap := TrapEffect{
		RoomID:    1,
		Element:   types.Fire,
		SlowTicks: 0,
	}

	result := ApplyTrap(invader, trap, params)

	if result.IsSlowed {
		t.Error("expected IsSlowed=false when SlowTicks=0")
	}
	if result.SlowTicksApplied != 0 {
		t.Errorf("expected SlowTicksApplied=0, got %d", result.SlowTicksApplied)
	}
}

func TestApplyTrap_AllElementCombinations(t *testing.T) {
	params := DefaultCombatParams()
	elements := []types.Element{types.Wood, types.Fire, types.Earth, types.Metal, types.Water}

	for _, trapElem := range elements {
		for _, invaderElem := range elements {
			invader := &Invader{
				ID:      1,
				Element: invaderElem,
				HP:      1000,
				MaxHP:   1000,
			}
			trap := TrapEffect{
				RoomID:    1,
				Element:   trapElem,
				SlowTicks: 1,
			}

			result := ApplyTrap(invader, trap, params)

			if types.Overcomes(trapElem, invaderElem) {
				expected := int(float64(params.TrapDamageBase) * params.TrapElementMultiplier)
				if result.Damage != expected {
					t.Errorf("trap %v vs invader %v: expected advantage damage %d, got %d",
						trapElem, invaderElem, expected, result.Damage)
				}
			} else {
				if result.Damage != params.TrapDamageBase {
					t.Errorf("trap %v vs invader %v: expected base damage %d, got %d",
						trapElem, invaderElem, params.TrapDamageBase, result.Damage)
				}
			}

			if result.Damage < params.MinDamage {
				t.Errorf("trap %v vs invader %v: damage %d below minimum %d",
					trapElem, invaderElem, result.Damage, params.MinDamage)
			}
		}
	}
}
