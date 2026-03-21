package senju

// ActionType represents the type of action a beast can take in one tick.
type ActionType int

const (
	// Stay means the beast remains in its current room.
	Stay ActionType = iota
	// MoveToRoom means the beast moves to an adjacent room.
	MoveToRoom
	// Attack means the beast attacks a target in the same room.
	Attack
	// Retreat means the beast retreats toward safety.
	Retreat
)

// String returns the name of the action type.
func (a ActionType) String() string {
	switch a {
	case Stay:
		return "Stay"
	case MoveToRoom:
		return "MoveToRoom"
	case Attack:
		return "Attack"
	case Retreat:
		return "Retreat"
	default:
		return "Unknown"
	}
}

// Action represents the action a beast decides to take in one tick.
type Action struct {
	// Type is the kind of action to perform.
	Type ActionType
	// TargetRoomID is the destination room for MoveToRoom/Retreat actions.
	// 0 if not applicable.
	TargetRoomID int
	// TargetBeastID is the ID of the beast to attack for Attack actions.
	// 0 if not applicable.
	TargetBeastID int
}
