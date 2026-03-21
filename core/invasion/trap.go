package invasion

import (
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// TrapEffect represents the trap installed in a specific room.
type TrapEffect struct {
	// RoomID is the room where this trap is active.
	RoomID int
	// Element is the elemental attribute of the trap (inherited from the room type).
	Element types.Element
	// DamagePerTrigger is the base damage dealt when the trap triggers.
	DamagePerTrigger int
	// SlowTicks is the number of ticks the invader is slowed after triggering.
	SlowTicks int
}

// TrapResult describes the outcome of applying a trap to an invader.
type TrapResult struct {
	// InvaderID is the ID of the affected invader.
	InvaderID int
	// Damage is the actual damage dealt.
	Damage int
	// IsSlowed indicates whether the invader was slowed.
	IsSlowed bool
	// SlowTicksApplied is the number of slow ticks applied.
	SlowTicksApplied int
}

// trapRoomTypeID is the room type ID for trap rooms.
const trapRoomTypeID = "trap_room"

// BuildTrapEffects constructs TrapEffect entries from all trap rooms in the cave.
// Only rooms whose type ID is "trap_room" produce a TrapEffect.
func BuildTrapEffects(cave *world.Cave, rooms []world.Room, roomTypes *world.RoomTypeRegistry) []TrapEffect {
	var effects []TrapEffect
	for _, room := range rooms {
		if room.TypeID != trapRoomTypeID {
			continue
		}
		rt, err := roomTypes.Get(room.TypeID)
		if err != nil {
			continue
		}
		effects = append(effects, TrapEffect{
			RoomID:           room.ID,
			Element:          rt.Element,
			DamagePerTrigger: rt.BaseChiCapacity, // use chi capacity as base damage scaling
			SlowTicks:        2,
		})
	}
	return effects
}

// ApplyTrap applies a trap effect to an invader and returns the result.
// Damage is calculated as: TrapDamageBase × element multiplier.
// If the trap's element overcomes the invader's element, TrapElementMultiplier is applied.
// The invader's HP is reduced and SlowTicks are added.
func ApplyTrap(invader *Invader, trap TrapEffect, params CombatParams) TrapResult {
	damage := params.TrapDamageBase

	// Apply element multiplier if trap overcomes invader
	if types.Overcomes(trap.Element, invader.Element) {
		damage = int(float64(damage) * params.TrapElementMultiplier)
	}

	// Ensure minimum damage
	if damage < params.MinDamage {
		damage = params.MinDamage
	}

	// Apply damage
	invader.HP -= damage
	if invader.HP < 0 {
		invader.HP = 0
	}

	// Apply slow effect
	invader.SlowTicks += trap.SlowTicks

	return TrapResult{
		InvaderID:        invader.ID,
		Damage:           damage,
		IsSlowed:         trap.SlowTicks > 0,
		SlowTicksApplied: trap.SlowTicks,
	}
}
