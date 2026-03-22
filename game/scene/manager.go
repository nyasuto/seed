package scene

import (
	"image"
)

// Scene represents a game scene with lifecycle callbacks.
// Draw receives an image.Image to avoid a direct ebiten dependency,
// enabling headless testing. Concrete scenes cast it to *ebiten.Image.
type Scene interface {
	Update() error
	Draw(screen image.Image)
	OnEnter()
	OnExit()
}

// SceneManager manages the current scene and handles transitions.
type SceneManager struct {
	current Scene
}

// NewSceneManager creates a SceneManager with an initial scene.
// OnEnter is called on the initial scene.
func NewSceneManager(initial Scene) *SceneManager {
	sm := &SceneManager{}
	if initial != nil {
		sm.current = initial
		initial.OnEnter()
	}
	return sm
}

// Current returns the active scene.
func (sm *SceneManager) Current() Scene {
	return sm.current
}

// Switch transitions from the current scene to a new scene.
// OnExit is called on the old scene, then OnEnter on the new scene.
func (sm *SceneManager) Switch(next Scene) {
	if sm.current != nil {
		sm.current.OnExit()
	}
	sm.current = next
	if next != nil {
		next.OnEnter()
	}
}

// Update delegates to the current scene's Update.
func (sm *SceneManager) Update() error {
	if sm.current == nil {
		return nil
	}
	return sm.current.Update()
}

// Draw delegates to the current scene's Draw.
func (sm *SceneManager) Draw(screen image.Image) {
	if sm.current == nil {
		return
	}
	sm.current.Draw(screen)
}
