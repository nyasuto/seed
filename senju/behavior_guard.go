package senju

// GuardBehavior implements the Guard AI pattern.
// The beast stays in its assigned room and attacks intruders that enter.
type GuardBehavior struct{}

// DecideAction returns Attack if there are intruders in the beast's room,
// otherwise Stay.
func (g *GuardBehavior) DecideAction(ctx BehaviorContext) Action {
	invaders := ctx.InvaderRoomIDs[ctx.RoomID]
	if len(invaders) > 0 {
		return Action{
			Type:          Attack,
			TargetBeastID: invaders[0],
		}
	}
	return Action{Type: Stay}
}

// Type returns Guard.
func (g *GuardBehavior) Type() BehaviorType {
	return Guard
}
