package input

import (
	"fmt"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/world"
)

// UpgradeRoomFlow manages the single-step UpgradeRoom interaction flow:
// 1. Select a room cell on the map
// 2. Generate an UpgradeRoomAction
type UpgradeRoomFlow struct {
	complete bool
}

// NewUpgradeRoomFlow creates a new UpgradeRoomFlow.
func NewUpgradeRoomFlow() *UpgradeRoomFlow {
	return &UpgradeRoomFlow{}
}

// Complete returns true if the flow has generated an action.
func (f *UpgradeRoomFlow) Complete() bool {
	return f.complete
}

// TrySelectRoom attempts to select a room for upgrading.
// Returns the generated UpgradeRoomAction on success, or an error if the
// cell is not a room cell or the roomID is invalid.
func (f *UpgradeRoomFlow) TrySelectRoom(cellType world.CellType, roomID int) (*simulation.UpgradeRoomAction, error) {
	if cellType != world.RoomFloor && cellType != world.Entrance {
		return nil, fmt.Errorf("cannot upgrade: cell is %s, not a room", cellType)
	}
	if roomID == 0 {
		return nil, fmt.Errorf("cannot upgrade: no room at this cell")
	}
	f.complete = true
	return &simulation.UpgradeRoomAction{RoomID: roomID}, nil
}

// Cancel resets the flow.
func (f *UpgradeRoomFlow) Cancel() {
	f.complete = false
}
