package senju

import "github.com/ponpoko/chaosseed-core/fengshui"

// BehaviorType represents the type of AI behavior pattern assigned to a beast.
type BehaviorType int

const (
	// Guard is a stationary defense behavior. The beast stays in its room
	// and attacks intruders that enter.
	Guard BehaviorType = iota
	// Patrol is a roaming behavior. The beast moves through adjacent rooms
	// on a set route and chases intruders it encounters.
	Patrol
	// Chase is a pursuit behavior. The beast moves toward a detected intruder
	// through adjacent rooms.
	Chase
	// Flee is an escape behavior. The beast retreats when HP is critically low,
	// seeking a recovery room.
	Flee
)

// String returns the name of the behavior type.
func (b BehaviorType) String() string {
	switch b {
	case Guard:
		return "Guard"
	case Patrol:
		return "Patrol"
	case Chase:
		return "Chase"
	case Flee:
		return "Flee"
	default:
		return "Unknown"
	}
}

// BehaviorContext provides environmental information for a beast's AI decision.
type BehaviorContext struct {
	// Beast is the beast making the decision.
	Beast *Beast
	// RoomID is the current room the beast occupies.
	RoomID int
	// AdjacentRoomIDs is the list of room IDs directly connected to the current room.
	AdjacentRoomIDs []int
	// RoomBeasts maps room ID to the list of beast IDs in that room (friendly beasts).
	RoomBeasts map[int][]int
	// InvaderRoomIDs maps room ID to the list of invader IDs in that room.
	InvaderRoomIDs map[int][]int
	// RoomChi is the chi state of the beast's current room. May be nil.
	RoomChi *fengshui.RoomChi
}

// Behavior is the interface for beast AI behavior patterns.
// Each implementation decides what action a beast should take given its context.
type Behavior interface {
	// DecideAction determines the action a beast should take this tick.
	DecideAction(ctx BehaviorContext) Action
	// Type returns the BehaviorType of this behavior.
	Type() BehaviorType
}
