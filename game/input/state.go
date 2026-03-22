package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputStateMachine manages the current ActionMode and handles mode transitions
// based on keyboard input.
type InputStateMachine struct {
	mode ActionMode
}

// NewInputStateMachine creates an InputStateMachine starting in ModeNormal.
func NewInputStateMachine() *InputStateMachine {
	return &InputStateMachine{mode: ModeNormal}
}

// Mode returns the current ActionMode.
func (sm *InputStateMachine) Mode() ActionMode {
	return sm.mode
}

// SetMode sets the current ActionMode directly.
func (sm *InputStateMachine) SetMode(mode ActionMode) {
	sm.mode = mode
}

// modeKeys maps ebiten keys to action modes for mode transitions.
var modeKeys = []struct {
	key  ebiten.Key
	mode ActionMode
}{
	{ebiten.KeyD, ModeDigRoom},
	{ebiten.KeyC, ModeDigCorridor},
	{ebiten.KeyS, ModeSummon},
	{ebiten.KeyU, ModeUpgrade},
}

// Update checks keyboard input and transitions the mode accordingly.
// Key mappings:
//   - Escape → ModeNormal (cancel current mode)
//   - D → ModeDigRoom
//   - C → ModeDigCorridor
//   - S → ModeSummon
//   - U → ModeUpgrade
//
// Call this once per frame in Game.Update().
func (sm *InputStateMachine) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		sm.mode = ModeNormal
		return
	}
	for _, mk := range modeKeys {
		if inpututil.IsKeyJustPressed(mk.key) {
			sm.mode = mk.mode
			return
		}
	}
}
