package senju

import (
	"math"

	"github.com/ponpoko/chaosseed-core/fengshui"
)

// CombatStats holds the effective combat statistics of a beast,
// after applying room affinity modifiers.
type CombatStats struct {
	// HP is the effective hit points.
	HP int
	// ATK is the effective attack power.
	ATK int
	// DEF is the effective defense power.
	DEF int
	// SPD is the effective speed.
	SPD int
}

// CalcCombatStats computes the effective combat stats for this beast,
// applying the element affinity multiplier based on the room's chi element.
// If roomChi is nil, base stats are returned unmodified.
func (b *Beast) CalcCombatStats(roomChi *fengshui.RoomChi) CombatStats {
	if roomChi == nil {
		return CombatStats{
			HP:  b.HP,
			ATK: b.ATK,
			DEF: b.DEF,
			SPD: b.SPD,
		}
	}

	mult := RoomAffinity(b.Element, roomChi.Element)
	return CombatStats{
		HP:  b.HP,
		ATK: int(math.Round(float64(b.ATK) * mult)),
		DEF: int(math.Round(float64(b.DEF) * mult)),
		SPD: int(math.Round(float64(b.SPD) * mult)),
	}
}
