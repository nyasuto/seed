package senju

// FleeHPThreshold is the default HP ratio (HP/MaxHP) below which a beast
// will attempt to flee. This can be overridden by BehaviorParams.
const FleeHPThreshold = 0.25

// RecoveryRoomTypeID is the room type ID for recovery rooms.
const RecoveryRoomTypeID = "recovery_room"

// FleeBehavior implements the Flee AI pattern.
// The beast retreats from invaders when HP drops below a threshold,
// seeking a recovery room to heal.
type FleeBehavior struct {
	// HPThreshold is the HP ratio (HP/MaxHP) below which flee triggers.
	HPThreshold float64
	// RoomTypeIDs maps room ID to room type ID, used to find recovery rooms.
	RoomTypeIDs map[int]string
}

// NewFleeBehavior creates a FleeBehavior with the given HP threshold.
// roomTypeIDs maps room ID to room type ID string for recovery room detection.
func NewFleeBehavior(hpThreshold float64, roomTypeIDs map[int]string) *FleeBehavior {
	return &FleeBehavior{
		HPThreshold: hpThreshold,
		RoomTypeIDs: roomTypeIDs,
	}
}

// ShouldFlee checks whether the beast's HP is low enough to trigger flee.
func ShouldFlee(beast *Beast, hpThreshold float64) bool {
	if beast.MaxHP <= 0 {
		return false
	}
	return float64(beast.HP)/float64(beast.MaxHP) <= hpThreshold
}

// DecideAction determines the flee action for this tick.
// The beast tries to move to an adjacent recovery room if available.
// Otherwise, it moves to the adjacent room farthest from any invader.
// If it is already in a recovery room, it stays and transitions to Recovering.
func (f *FleeBehavior) DecideAction(ctx BehaviorContext) Action {
	// Already in a recovery room — stay and recover.
	if f.RoomTypeIDs[ctx.RoomID] == RecoveryRoomTypeID {
		return Action{Type: Stay}
	}

	// Look for an adjacent recovery room.
	for _, adjID := range ctx.AdjacentRoomIDs {
		if f.RoomTypeIDs[adjID] == RecoveryRoomTypeID {
			return Action{
				Type:         Retreat,
				TargetRoomID: adjID,
			}
		}
	}

	// No recovery room adjacent — move to the room farthest from invaders.
	bestRoom := 0
	bestScore := -1

	for _, adjID := range ctx.AdjacentRoomIDs {
		// Score: prefer rooms with no invaders. Count invaders as negative.
		invaderCount := len(ctx.InvaderRoomIDs[adjID])
		score := 100 - invaderCount
		if score > bestScore {
			bestScore = score
			bestRoom = adjID
		}
	}

	if bestRoom != 0 {
		return Action{
			Type:         Retreat,
			TargetRoomID: bestRoom,
		}
	}

	// No adjacent rooms — stay.
	return Action{Type: Stay}
}

// Type returns Flee.
func (f *FleeBehavior) Type() BehaviorType {
	return Flee
}
