package controller

import "github.com/nyasuto/seed/core/simulation"

// DefaultFastForwardSpeed is the number of ticks advanced per Update call
// when in FastForward mode.
const DefaultFastForwardSpeed = 1

// TogglePause toggles between Paused and Playing states.
// Has no effect when the game is over.
func (gc *GameController) TogglePause() {
	switch gc.state {
	case Playing:
		gc.state = Paused
	case Paused:
		gc.state = Playing
	}
}

// StartFastForward enters FastForward mode with the given speed (ticks per
// Update call). If speed < 1 it defaults to DefaultFastForwardSpeed.
// Has no effect when the game is over.
func (gc *GameController) StartFastForward(speed int) {
	if gc.state == GameOver {
		return
	}
	if speed < 1 {
		speed = DefaultFastForwardSpeed
	}
	gc.state = FastForward
	gc.ffSpeed = speed
}

// StopFastForward exits FastForward mode, returning to Playing.
// Has no effect unless currently in FastForward.
func (gc *GameController) StopFastForward() {
	if gc.state == FastForward {
		gc.state = Playing
	}
}

// FastForwardSpeed returns the current fast-forward speed. Returns 0 when
// not in FastForward mode.
func (gc *GameController) FastForwardSpeed() int {
	if gc.state != FastForward {
		return 0
	}
	return gc.ffSpeed
}

// UpdateTick is intended to be called once per frame. It advances ticks
// according to the current game state:
//   - Playing / Paused: no automatic advancement (manual AdvanceTick only)
//   - FastForward: advances ffSpeed ticks, stopping early on game over
//
// It returns the number of ticks actually advanced and any error.
func (gc *GameController) UpdateTick() (int, error) {
	if gc.state != FastForward {
		return 0, nil
	}

	advanced := 0
	for i := 0; i < gc.ffSpeed; i++ {
		result, err := gc.AdvanceTick()
		if err != nil {
			return advanced, err
		}
		advanced++
		if result.Status != simulation.Running {
			// Game ended — state is already set to GameOver by AdvanceTick.
			return advanced, nil
		}
	}
	return advanced, nil
}
