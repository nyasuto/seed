package view

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/asset"
)

const (
	// BattleBlinkPeriod is the number of frames per blink cycle for battle rooms.
	BattleBlinkPeriod = 20
	// WaveAlertDuration is the number of frames a wave alert is displayed (~3s at 60 FPS).
	WaveAlertDuration = 180
	// waveAlertFadeStart is the frame count at which fade-out begins.
	waveAlertFadeStart = 60
	// HPBlinkDuration is the number of frames the CoreHP bar blinks after damage.
	HPBlinkDuration = 60
	// HPBlinkPeriod is the number of frames per blink cycle for CoreHP bar.
	HPBlinkPeriod = 10
	// BeastDefeatBlinkDuration is how many frames a defeated beast blinks.
	BeastDefeatBlinkDuration = 90
	// BeastDefeatBlinkPeriod is the number of frames per blink cycle for defeated beasts.
	BeastDefeatBlinkPeriod = 15
)

// BattleFeedback tracks combat-related visual feedback state.
type BattleFeedback struct {
	// frame is the global frame counter for blink calculations.
	frame int

	// waveAlertMsg is the current wave arrival message.
	waveAlertMsg string
	// waveAlertTimer counts down frames for wave alert display.
	waveAlertTimer int

	// prevActiveWaves tracks the number of active waves from the previous tick.
	prevActiveWaves int

	// prevCoreHP tracks CoreHP from the previous tick for damage detection.
	prevCoreHP int
	// hpBlinkTimer counts down frames for HP bar blinking.
	hpBlinkTimer int

	// defeatedBeastIDs tracks recently defeated beast IDs with blink timers.
	defeatedBeastIDs map[int]int

	// initialized is set after the first Update call with valid state.
	initialized bool
}

// NewBattleFeedback creates a new BattleFeedback.
func NewBattleFeedback() *BattleFeedback {
	return &BattleFeedback{
		defeatedBeastIDs: make(map[int]int),
	}
}

// Update processes one frame of battle feedback state. It should be called
// each frame with the current game state.
func (bf *BattleFeedback) Update(coreHP int, waves []*invasion.InvasionWave, beasts []*senju.Beast) {
	bf.frame++

	activeWaves := countActiveWaves(waves)

	if !bf.initialized {
		bf.prevActiveWaves = activeWaves
		bf.prevCoreHP = coreHP
		bf.initialized = true
	}

	// Decrement timers first, before detecting new events.
	if bf.waveAlertTimer > 0 {
		bf.waveAlertTimer--
		if bf.waveAlertTimer == 0 {
			bf.waveAlertMsg = ""
		}
	}
	if bf.hpBlinkTimer > 0 {
		bf.hpBlinkTimer--
	}
	for id, timer := range bf.defeatedBeastIDs {
		if timer <= 1 {
			delete(bf.defeatedBeastIDs, id)
		} else {
			bf.defeatedBeastIDs[id] = timer - 1
		}
	}

	// Detect new wave arrivals.
	if activeWaves > bf.prevActiveWaves {
		newCount := activeWaves - bf.prevActiveWaves
		bf.waveAlertMsg = FormatWaveAlert(newCount, activeWaves)
		bf.waveAlertTimer = WaveAlertDuration
	}
	bf.prevActiveWaves = activeWaves

	// Detect CoreHP decrease.
	if coreHP < bf.prevCoreHP {
		bf.hpBlinkTimer = HPBlinkDuration
	}
	bf.prevCoreHP = coreHP

	// Detect newly stunned beasts.
	for _, beast := range beasts {
		if beast.State == senju.Stunned {
			if _, exists := bf.defeatedBeastIDs[beast.ID]; !exists {
				bf.defeatedBeastIDs[beast.ID] = BeastDefeatBlinkDuration
			}
		}
	}
}

// BattleRoomIDs returns the set of room IDs where combat is currently happening.
// A room is considered in battle if it contains a Fighting invader.
func BattleRoomIDs(waves []*invasion.InvasionWave) map[int]bool {
	rooms := make(map[int]bool)
	for _, wave := range waves {
		if !wave.IsActive() {
			continue
		}
		for _, inv := range wave.Invaders {
			if inv.State == invasion.Fighting && inv.CurrentRoomID > 0 {
				rooms[inv.CurrentRoomID] = true
			}
		}
	}
	return rooms
}

// IsBlinkOn returns whether the blink is in its "on" phase for the given period.
func (bf *BattleFeedback) IsBlinkOn(period int) bool {
	if period <= 0 {
		return false
	}
	return (bf.frame/period)%2 == 0
}

// DrawBattleRoomOverlays draws red border overlays on rooms where combat is happening.
func (bf *BattleFeedback) DrawBattleRoomOverlays(screen *ebiten.Image, cave *world.Cave, mv *MapView, battleRooms map[int]bool) {
	if len(battleRooms) == 0 {
		return
	}
	if !bf.IsBlinkOn(BattleBlinkPeriod) {
		return
	}

	borderColor := color.RGBA{R: 0xFF, G: 0x22, B: 0x22, A: 0x60}
	for _, room := range cave.Rooms {
		if !battleRooms[room.ID] {
			continue
		}
		// Draw red overlay on each cell of the room's border.
		drawRoomBorderOverlay(screen, mv, room, borderColor)
	}
}

// drawRoomBorderOverlay draws a colored overlay on the border cells of a room.
func drawRoomBorderOverlay(screen *ebiten.Image, mv *MapView, room *world.Room, c color.RGBA) {
	tile := ebiten.NewImage(asset.TileSize, asset.TileSize)
	tile.Fill(c)

	for dy := 0; dy < room.Height; dy++ {
		for dx := 0; dx < room.Width; dx++ {
			// Only border cells (edges of the room area).
			if dy > 0 && dy < room.Height-1 && dx > 0 && dx < room.Width-1 {
				continue
			}
			px, py := mv.CellToScreen(room.Pos.X+dx, room.Pos.Y+dy)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(px), float64(py))
			screen.DrawImage(tile, op)
		}
	}
}

// WaveAlertMessage returns the current wave alert message, or "" if none.
func (bf *BattleFeedback) WaveAlertMessage() string {
	return bf.waveAlertMsg
}

// WaveAlertAlpha returns the opacity for the wave alert text.
func (bf *BattleFeedback) WaveAlertAlpha() float64 {
	if bf.waveAlertTimer <= 0 {
		return 0
	}
	if bf.waveAlertTimer > waveAlertFadeStart {
		return 1.0
	}
	return float64(bf.waveAlertTimer) / float64(waveAlertFadeStart)
}

// DrawWaveAlert renders the wave arrival warning at the top of the screen.
func (bf *BattleFeedback) DrawWaveAlert(screen *ebiten.Image, screenWidth int) {
	if bf.waveAlertMsg == "" {
		return
	}
	yellow := color.RGBA{R: 0xFF, G: 0xD7, B: 0x00, A: 0xFF}
	textW := TextWidth(bf.waveAlertMsg)
	x := (screenWidth - textW) / 2
	y := 34 // Just below the top bar.
	DrawColoredText(screen, bf.waveAlertMsg, x, y, yellow, bf.WaveAlertAlpha())
}

// IsHPBlinking returns whether the CoreHP bar should currently be blinking.
func (bf *BattleFeedback) IsHPBlinking() bool {
	return bf.hpBlinkTimer > 0 && bf.IsBlinkOn(HPBlinkPeriod)
}

// HPBlinkTimer returns the remaining blink timer for CoreHP.
func (bf *BattleFeedback) HPBlinkTimer() int {
	return bf.hpBlinkTimer
}

// IsBeastBlinking returns whether the given beast should be blinking (recently defeated).
func (bf *BattleFeedback) IsBeastBlinking(beastID int) bool {
	timer, ok := bf.defeatedBeastIDs[beastID]
	if !ok {
		return false
	}
	_ = timer
	// Blink: alternate visibility.
	return !bf.IsBlinkOn(BeastDefeatBlinkPeriod)
}

// BlinkHiddenBeastIDs returns the set of beast IDs that should be hidden
// in the current frame due to defeat blink effect.
func (bf *BattleFeedback) BlinkHiddenBeastIDs() map[int]bool {
	if len(bf.defeatedBeastIDs) == 0 {
		return nil
	}
	if bf.IsBlinkOn(BeastDefeatBlinkPeriod) {
		return nil // visible phase
	}
	hidden := make(map[int]bool, len(bf.defeatedBeastIDs))
	for id := range bf.defeatedBeastIDs {
		hidden[id] = true
	}
	return hidden
}

// FormatWaveAlert generates the wave alert message.
func FormatWaveAlert(newWaves, totalActive int) string {
	if newWaves <= 1 {
		return fmt.Sprintf("Wave incoming! (%d active)", totalActive)
	}
	return fmt.Sprintf("%d Waves incoming! (%d active)", newWaves, totalActive)
}

// WaveArrivalDetected returns true if new waves arrived this tick compared to
// the previous tick. This is a pure helper for testing wave detection logic.
func WaveArrivalDetected(prevActiveWaves, currentActiveWaves int) bool {
	return currentActiveWaves > prevActiveWaves
}

// countActiveWaves counts the number of active waves.
func countActiveWaves(waves []*invasion.InvasionWave) int {
	count := 0
	for _, wave := range waves {
		if wave.IsActive() {
			count++
		}
	}
	return count
}
