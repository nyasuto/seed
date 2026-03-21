package senju

import "github.com/nyasuto/seed/core/types"

// Affinity multiplier constants for beast-room element interactions.
const (
	// GeneratesMultiplier is applied when the room element generates the beast element.
	GeneratesMultiplier = 1.3
	// SameElementMultiplier is applied when beast and room share the same element.
	SameElementMultiplier = 1.1
	// NeutralMultiplier is the default multiplier when no special relationship exists.
	NeutralMultiplier = 1.0
	// OvercomesMultiplier is applied when the room element overcomes the beast element.
	OvercomesMultiplier = 0.7
)

// RoomAffinity returns the affinity multiplier between a beast's element
// and the room's element. This affects combat stats.
//
// Relationships checked (in priority order):
//   - Room generates beast element (相生): 1.3
//   - Same element: 1.1
//   - Room overcomes beast element (相克): 0.7
//   - Otherwise neutral: 1.0
func RoomAffinity(beastElement, roomElement types.Element) float64 {
	if types.Generates(roomElement, beastElement) {
		return GeneratesMultiplier
	}
	if beastElement == roomElement {
		return SameElementMultiplier
	}
	if types.Overcomes(roomElement, beastElement) {
		return OvercomesMultiplier
	}
	return NeutralMultiplier
}

// GrowthAffinity returns the growth speed multiplier between a beast's element
// and the room's element. Currently uses the same values as RoomAffinity,
// but is separated to allow independent tuning in the future.
func GrowthAffinity(beastElement, roomElement types.Element) float64 {
	return RoomAffinity(beastElement, roomElement)
}
