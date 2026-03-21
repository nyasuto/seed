package senju

// PatrolBehavior implements the Patrol AI pattern.
// The beast roams through adjacent rooms in order. When it encounters
// an intruder, it transitions to Chase behavior.
type PatrolBehavior struct {
	// PatrolRoute is the ordered list of room IDs to visit.
	PatrolRoute []int
	// RouteIndex is the current position in the patrol route.
	RouteIndex int
	// RestTicks is the number of ticks to stay in each room before moving.
	RestTicks int
	// currentRest tracks how many ticks the beast has stayed in the current room.
	currentRest int
}

// NewPatrolBehavior creates a PatrolBehavior with a route generated from
// the beast's home room and its adjacent rooms.
func NewPatrolBehavior(homeRoomID int, adjacentRoomIDs []int, restTicks int) *PatrolBehavior {
	route := make([]int, 0, 1+len(adjacentRoomIDs))
	route = append(route, homeRoomID)
	route = append(route, adjacentRoomIDs...)
	return &PatrolBehavior{
		PatrolRoute: route,
		RouteIndex:  0,
		RestTicks:   restTicks,
		currentRest: 0,
	}
}

// DecideAction returns the patrol action for this tick.
// If an intruder is detected in the current room or adjacent rooms,
// the beast moves toward the intruder (signaling a Chase transition).
// Otherwise, it follows its patrol route.
func (p *PatrolBehavior) DecideAction(ctx BehaviorContext) Action {
	// Check for invaders in the current room — attack immediately.
	if invaders := ctx.InvaderRoomIDs[ctx.RoomID]; len(invaders) > 0 {
		return Action{
			Type:          Attack,
			TargetBeastID: invaders[0],
		}
	}

	// Check for invaders in adjacent rooms — move toward the first one found.
	for _, adjID := range ctx.AdjacentRoomIDs {
		if invaders := ctx.InvaderRoomIDs[adjID]; len(invaders) > 0 {
			return Action{
				Type:         MoveToRoom,
				TargetRoomID: adjID,
			}
		}
	}

	// Normal patrol: stay for RestTicks, then move to next room in route.
	if p.currentRest < p.RestTicks {
		p.currentRest++
		return Action{Type: Stay}
	}

	// Advance to next room in patrol route.
	p.RouteIndex = (p.RouteIndex + 1) % len(p.PatrolRoute)
	nextRoom := p.PatrolRoute[p.RouteIndex]
	p.currentRest = 0

	// If we are already in the next room, stay.
	if nextRoom == ctx.RoomID {
		return Action{Type: Stay}
	}

	return Action{
		Type:         MoveToRoom,
		TargetRoomID: nextRoom,
	}
}

// Type returns Patrol.
func (p *PatrolBehavior) Type() BehaviorType {
	return Patrol
}
