package input

// ActionMode represents the current player interaction mode.
type ActionMode int

const (
	// ModeNormal is the default mode for viewing and selecting.
	ModeNormal ActionMode = iota
	// ModeDigRoom is the mode for placing a new room.
	ModeDigRoom
	// ModeDigCorridor is the mode for connecting two rooms with a corridor.
	ModeDigCorridor
	// ModeSummon is the mode for summoning a beast into a room.
	ModeSummon
	// ModeUpgrade is the mode for upgrading a room.
	ModeUpgrade
)

// String returns a human-readable name for the ActionMode.
func (m ActionMode) String() string {
	switch m {
	case ModeNormal:
		return "Normal"
	case ModeDigRoom:
		return "DigRoom"
	case ModeDigCorridor:
		return "DigCorridor"
	case ModeSummon:
		return "Summon"
	case ModeUpgrade:
		return "Upgrade"
	default:
		return "Unknown"
	}
}
