package input

import (
	"fmt"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/world"
)

// DigCorridorStep represents the current step in the DigCorridor interaction flow.
type DigCorridorStep int

const (
	// CorridorStepSelectFirst waits for the player to click a room cell as the starting point.
	CorridorStepSelectFirst DigCorridorStep = iota
	// CorridorStepSelectSecond waits for the player to click another room cell as the endpoint.
	CorridorStepSelectSecond
	// CorridorStepComplete indicates the flow finished and an action was generated.
	CorridorStepComplete
)

// DigCorridorFlow manages the multi-step DigCorridor interaction flow:
// 1. Select a room cell as the starting point (FromRoomID)
// 2. Select another room cell as the endpoint (ToRoomID)
// 3. Generate a DigCorridorAction
type DigCorridorFlow struct {
	step       DigCorridorStep
	fromRoomID int
}

// NewDigCorridorFlow creates a new DigCorridorFlow starting at first room selection.
func NewDigCorridorFlow() *DigCorridorFlow {
	return &DigCorridorFlow{step: CorridorStepSelectFirst}
}

// Step returns the current step in the flow.
func (f *DigCorridorFlow) Step() DigCorridorStep {
	return f.step
}

// FromRoomID returns the selected starting room ID.
// Only meaningful after CorridorStepSelectFirst completes successfully.
func (f *DigCorridorFlow) FromRoomID() int {
	return f.fromRoomID
}

// TrySelectRoom attempts to select a room for corridor connection.
// cellType must be a room-related cell type (RoomFloor or Entrance).
// roomID must be a valid room ID (> 0).
// On the first call, records the starting room and advances to CorridorStepSelectSecond.
// On the second call, validates the room is different from the first and generates the action.
func (f *DigCorridorFlow) TrySelectRoom(cellType world.CellType, roomID int) (*simulation.DigCorridorAction, error) {
	if !isRoomCell(cellType) {
		return nil, fmt.Errorf("cannot connect: cell is %s, not a room", cellType)
	}
	if roomID <= 0 {
		return nil, fmt.Errorf("cannot connect: no room at this cell")
	}

	switch f.step {
	case CorridorStepSelectFirst:
		f.fromRoomID = roomID
		f.step = CorridorStepSelectSecond
		return nil, nil

	case CorridorStepSelectSecond:
		if roomID == f.fromRoomID {
			return nil, fmt.Errorf("cannot connect: start and end room are the same (room %d)", roomID)
		}
		action := &simulation.DigCorridorAction{
			FromRoomID: f.fromRoomID,
			ToRoomID:   roomID,
		}
		f.step = CorridorStepComplete
		return action, nil

	default:
		return nil, fmt.Errorf("flow already complete")
	}
}

// Cancel resets the flow back to first room selection step.
func (f *DigCorridorFlow) Cancel() {
	f.step = CorridorStepSelectFirst
	f.fromRoomID = 0
}

// isRoomCell returns true if the cell type belongs to a room.
func isRoomCell(ct world.CellType) bool {
	return ct == world.RoomFloor || ct == world.Entrance
}
