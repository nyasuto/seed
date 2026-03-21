package senju

// ChaseBehavior implements the Chase AI pattern.
// The beast pursues a detected intruder by moving through adjacent rooms
// toward the intruder's location. If the beast reaches the same room as
// the intruder, it attacks. If the intruder is lost for a timeout period,
// the beast returns to Patrol behavior.
type ChaseBehavior struct {
	// TargetInvaderID is the ID of the intruder being chased.
	TargetInvaderID int
	// TimeoutTicks is the maximum number of ticks to chase before giving up.
	TimeoutTicks int
	// ticksChasing tracks how many ticks have been spent chasing.
	ticksChasing int
}

// NewChaseBehavior creates a ChaseBehavior targeting a specific invader.
func NewChaseBehavior(targetInvaderID int, timeoutTicks int) *ChaseBehavior {
	return &ChaseBehavior{
		TargetInvaderID: targetInvaderID,
		TimeoutTicks:    timeoutTicks,
		ticksChasing:    0,
	}
}

// DecideAction determines the chase action for this tick.
// If the intruder is in the same room, attack.
// If the intruder is in an adjacent room, move there.
// If the intruder cannot be found or timeout is reached, return Stay
// (the behavior engine should transition back to Patrol).
func (c *ChaseBehavior) DecideAction(ctx BehaviorContext) Action {
	c.ticksChasing++

	// Timeout — give up the chase.
	if c.ticksChasing > c.TimeoutTicks {
		return Action{Type: Stay}
	}

	// Intruder in the same room — attack.
	if invaders := ctx.InvaderRoomIDs[ctx.RoomID]; len(invaders) > 0 {
		for _, id := range invaders {
			if id == c.TargetInvaderID {
				return Action{
					Type:          Attack,
					TargetBeastID: c.TargetInvaderID,
				}
			}
		}
		// Target not here but other invaders are — attack the first one.
		return Action{
			Type:          Attack,
			TargetBeastID: invaders[0],
		}
	}

	// Look for the target invader in adjacent rooms.
	for _, adjID := range ctx.AdjacentRoomIDs {
		if invaders := ctx.InvaderRoomIDs[adjID]; len(invaders) > 0 {
			for _, id := range invaders {
				if id == c.TargetInvaderID {
					return Action{
						Type:         MoveToRoom,
						TargetRoomID: adjID,
					}
				}
			}
		}
	}

	// Check for any invader in adjacent rooms (even if not the original target).
	for _, adjID := range ctx.AdjacentRoomIDs {
		if invaders := ctx.InvaderRoomIDs[adjID]; len(invaders) > 0 {
			return Action{
				Type:         MoveToRoom,
				TargetRoomID: adjID,
			}
		}
	}

	// Invader not visible — stay and wait (or behavior engine transitions back).
	return Action{Type: Stay}
}

// TimedOut reports whether the chase has exceeded its timeout.
func (c *ChaseBehavior) TimedOut() bool {
	return c.ticksChasing > c.TimeoutTicks
}

// Type returns Chase.
func (c *ChaseBehavior) Type() BehaviorType {
	return Chase
}
