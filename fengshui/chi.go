package fengshui

import "github.com/ponpoko/chaosseed-core/types"

// RoomChi tracks the current chi energy level of a single room.
type RoomChi struct {
	// RoomID is the ID of the room this chi state belongs to.
	RoomID int
	// Current is the current amount of chi stored in the room.
	Current float64
	// Capacity is the maximum amount of chi this room can hold.
	Capacity float64
	// Element is the elemental attribute of this room's chi,
	// inherited from the room's RoomType.
	Element types.Element
}

// IsFull reports whether the room's chi is at maximum capacity.
// A room with zero capacity is always considered full.
func (rc *RoomChi) IsFull() bool {
	if rc.Capacity <= 0 {
		return true
	}
	return rc.Current >= rc.Capacity
}

// IsEmpty reports whether the room has no chi.
func (rc *RoomChi) IsEmpty() bool {
	return rc.Current <= 0
}

// Ratio returns the chi fill ratio (0.0 to 1.0).
// Returns 0 if capacity is zero or negative.
func (rc *RoomChi) Ratio() float64 {
	if rc.Capacity <= 0 {
		return 0
	}
	r := rc.Current / rc.Capacity
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}
