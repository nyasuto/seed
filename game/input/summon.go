package input

import (
	"fmt"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// SummonStep represents the current step in the SummonBeast interaction flow.
type SummonStep int

const (
	// SummonStepSelectRoom waits for the player to click a valid room cell.
	SummonStepSelectRoom SummonStep = iota
	// SummonStepSelectElement waits for the player to choose an element.
	SummonStepSelectElement
	// SummonStepComplete indicates the flow finished and an action was generated.
	SummonStepComplete
)

// SummonBeastFlow manages the multi-step SummonBeast interaction flow:
// 1. Select a room cell on the map (validates room exists and has beast capacity)
// 2. Choose an element from the selection panel
// 3. Generate a SummonBeastAction
//
// Note: The core SummonBeastAction only requires Element. The room selection
// step is for UX validation (checking beast capacity) but the roomID is not
// included in the generated action. The summoned beast goes to the unassigned
// pool and can later be placed with PlaceBeastAction.
type SummonBeastFlow struct {
	step   SummonStep
	roomID int
}

// NewSummonBeastFlow creates a new SummonBeastFlow starting at room selection.
func NewSummonBeastFlow() *SummonBeastFlow {
	return &SummonBeastFlow{step: SummonStepSelectRoom}
}

// Step returns the current step in the flow.
func (f *SummonBeastFlow) Step() SummonStep {
	return f.step
}

// RoomID returns the selected room ID.
// Only meaningful after SummonStepSelectRoom completes successfully.
func (f *SummonBeastFlow) RoomID() int {
	return f.roomID
}

// TrySelectRoom attempts to select a room for context validation.
// Returns an error if the cell is not a room cell, the room has no beast
// capacity, or the room already has beasts at capacity.
// On success, advances the flow to SummonStepSelectElement.
func (f *SummonBeastFlow) TrySelectRoom(cellType world.CellType, roomID int, roomHasCapacity bool) error {
	if cellType != world.RoomFloor && cellType != world.Entrance {
		return fmt.Errorf("cannot summon: cell is %s, not a room", cellType)
	}
	if roomID == 0 {
		return fmt.Errorf("cannot summon: no room at this cell")
	}
	if !roomHasCapacity {
		return fmt.Errorf("cannot summon: room %d already has a beast", roomID)
	}
	f.roomID = roomID
	f.step = SummonStepSelectElement
	return nil
}

// SelectElement completes the flow by choosing an element and building
// the SummonBeastAction.
func (f *SummonBeastFlow) SelectElement(elem types.Element) *simulation.SummonBeastAction {
	f.step = SummonStepComplete
	return &simulation.SummonBeastAction{Element: elem}
}

// Cancel resets the flow back to room selection step.
func (f *SummonBeastFlow) Cancel() {
	f.step = SummonStepSelectRoom
	f.roomID = 0
}
