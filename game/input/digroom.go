package input

import (
	"fmt"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// DigRoomStep represents the current step in the DigRoom interaction flow.
type DigRoomStep int

const (
	// StepSelectCell waits for the player to click a valid Rock cell.
	StepSelectCell DigRoomStep = iota
	// StepSelectElement waits for the player to choose an element.
	StepSelectElement
	// StepComplete indicates the flow finished and an action was generated.
	StepComplete
)

const (
	// DefaultRoomWidth is the default width for a new room.
	DefaultRoomWidth = 3
	// DefaultRoomHeight is the default height for a new room.
	DefaultRoomHeight = 3
)

// DigRoomFlow manages the multi-step DigRoom interaction flow:
// 1. Select a Rock cell on the map
// 2. Choose an element from the selection panel
// 3. Generate a DigRoomAction
type DigRoomFlow struct {
	step  DigRoomStep
	cellX int
	cellY int
}

// NewDigRoomFlow creates a new DigRoomFlow starting at cell selection.
func NewDigRoomFlow() *DigRoomFlow {
	return &DigRoomFlow{step: StepSelectCell}
}

// Step returns the current step in the flow.
func (f *DigRoomFlow) Step() DigRoomStep {
	return f.step
}

// SelectedCell returns the selected cell coordinates.
// Only meaningful after StepSelectCell completes successfully.
func (f *DigRoomFlow) SelectedCell() (int, int) {
	return f.cellX, f.cellY
}

// TrySelectCell attempts to select a cell for room placement.
// Returns an error if the cell is not a valid Rock cell.
// On success, advances the flow to StepSelectElement.
func (f *DigRoomFlow) TrySelectCell(cx, cy int, cellType world.CellType) error {
	if cellType == world.HardRock {
		return fmt.Errorf("cannot dig: hard rock at (%d,%d)", cx, cy)
	}
	if cellType == world.Water {
		return fmt.Errorf("cannot dig: water at (%d,%d)", cx, cy)
	}
	if cellType != world.Rock {
		return fmt.Errorf("cannot dig: cell at (%d,%d) is %s, not rock", cx, cy, cellType)
	}
	f.cellX = cx
	f.cellY = cy
	f.step = StepSelectElement
	return nil
}

// SelectElement completes the flow by choosing an element and building
// the DigRoomAction. It finds a room type matching the element from the
// registry (excluding dragon hole rooms with BaseCoreHP > 0).
// Returns the generated DigRoomAction, or an error if no matching room type exists.
func (f *DigRoomFlow) SelectElement(elem types.Element, registry *world.RoomTypeRegistry) (*simulation.DigRoomAction, error) {
	roomTypeID := findRoomTypeByElement(elem, registry)
	if roomTypeID == "" {
		return nil, fmt.Errorf("no room type found for element %s", elem)
	}

	action := &simulation.DigRoomAction{
		RoomTypeID: roomTypeID,
		Pos:        types.Pos{X: f.cellX, Y: f.cellY},
		Width:      DefaultRoomWidth,
		Height:     DefaultRoomHeight,
	}
	f.step = StepComplete
	return action, nil
}

// Cancel resets the flow back to cell selection step.
func (f *DigRoomFlow) Cancel() {
	f.step = StepSelectCell
	f.cellX = 0
	f.cellY = 0
}

// findRoomTypeByElement returns the ID of a room type matching the given element,
// excluding special rooms (dragon hole). Returns empty string if none found.
func findRoomTypeByElement(elem types.Element, registry *world.RoomTypeRegistry) string {
	for _, rt := range registry.All() {
		if rt.Element == elem && rt.BaseCoreHP == 0 {
			return rt.ID
		}
	}
	return ""
}
