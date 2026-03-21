package simulation

import (
	"fmt"
	"strings"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// RenderFullStatus returns a comprehensive ASCII status display combining
// all simulation layers into a single view. The output includes:
//
//   - A unified map overlay (terrain + chi levels + beasts + invaders)
//   - A status panel below the map showing tick, core HP, chi balance,
//     beast/invader counts, wave progress, feng shui score, and economy state
//
// Priority for room cell rendering (highest wins):
//
//	Invaders present → invader tile (>>, XX, <<, $$)
//	Beasts present   → beast tile (element char or count+element)
//	Chi flow active  → chi fill level (__, ░░, ▒▒, ▓▓, ██)
//	Default          → room ID
//
// Non-room cells show dragon vein paths (~~), corridors (..), entrances (><),
// and rock (██).
func RenderFullStatus(engine *SimulationEngine) string {
	s := engine.State
	var sb strings.Builder

	// Render the unified map.
	sb.WriteString(renderUnifiedMap(s))

	// Render the status panel.
	sb.WriteString(renderStatusPanel(s))

	return sb.String()
}

// roomBeastInfo holds pre-computed per-room beast display info.
type roomBeastInfo struct {
	count   int
	element types.Element
}

// roomInvaderInfo holds pre-computed per-room invader display info.
type roomInvaderInfo struct {
	count int
	state invasion.InvaderState
}

// renderUnifiedMap produces the combined map overlay with all layers.
func renderUnifiedMap(s *GameState) string {
	if s.Cave == nil {
		return ""
	}

	// Pre-compute per-room beast info.
	roomBeasts := make(map[int]*roomBeastInfo)
	for _, b := range s.Beasts {
		if b.RoomID == 0 {
			continue
		}
		info, ok := roomBeasts[b.RoomID]
		if !ok {
			roomBeasts[b.RoomID] = &roomBeastInfo{count: 1, element: b.Element}
		} else {
			info.count++
		}
	}

	// Pre-compute per-room invader info.
	roomInvaders := make(map[int]*roomInvaderInfo)
	for _, w := range s.Waves {
		for _, inv := range w.Invaders {
			if inv.State == invasion.Defeated || inv.CurrentRoomID == 0 {
				continue
			}
			info, ok := roomInvaders[inv.CurrentRoomID]
			if !ok {
				roomInvaders[inv.CurrentRoomID] = &roomInvaderInfo{count: 1, state: inv.State}
			} else {
				info.count++
			}
		}
	}

	// Build dragon vein path set.
	veinPaths := make(map[types.Pos]bool)
	if s.ChiFlowEngine != nil {
		for _, vein := range s.ChiFlowEngine.Veins {
			for _, pos := range vein.Path {
				veinPaths[pos] = true
			}
		}
	}

	return world.RenderGrid(s.Cave.Grid, func(pos types.Pos, cell world.Cell) string {
		switch cell.Type {
		case world.RoomFloor:
			return renderRoomCell(cell.RoomID, roomInvaders, roomBeasts, s.ChiFlowEngine)
		case world.CorridorFloor, world.Rock:
			if veinPaths[pos] {
				return "~~"
			}
		}
		return ""
	})
}

// renderRoomCell returns a 2-character tile for a room cell, applying
// the priority: invaders > beasts > chi > room ID.
func renderRoomCell(roomID int, invaders map[int]*roomInvaderInfo, beasts map[int]*roomBeastInfo, chiEngine *fengshui.ChiFlowEngine) string {
	if roomID <= 0 {
		return "[]"
	}

	// Priority 1: invaders
	if info, ok := invaders[roomID]; ok {
		return invaderTile(info)
	}

	// Priority 2: beasts
	if info, ok := beasts[roomID]; ok {
		return world.CountTile(info.count, info.element.Char())
	}

	// Priority 3: chi fill level
	if chiEngine != nil {
		if rc, ok := chiEngine.RoomChi[roomID]; ok {
			ratio := rc.Ratio()
			if ratio > 0 {
				return chiLevelTile(ratio)
			}
		}
	}

	// Default: room ID
	ch := world.RoomIDChar(roomID)
	return string([]byte{ch, ch})
}

// invaderTile returns a 2-character tile for a room's invader display.
func invaderTile(info *roomInvaderInfo) string {
	sym := invaderStateSymbol(info.state)
	if info.count == 1 {
		return sym
	}
	return world.CountTile(info.count, sym[1])
}

// invaderStateSymbol returns a 2-character symbol for an invader state.
func invaderStateSymbol(state invasion.InvaderState) string {
	switch state {
	case invasion.Advancing:
		return ">>"
	case invasion.Fighting:
		return "XX"
	case invasion.Retreating:
		return "<<"
	case invasion.GoalAchieved:
		return "$$"
	default:
		return "??"
	}
}

// chiLevelTile returns a 2-character tile for a chi fill ratio.
func chiLevelTile(ratio float64) string {
	switch {
	case ratio < 0.34:
		return "░░"
	case ratio < 0.67:
		return "▒▒"
	case ratio < 1.0:
		return "▓▓"
	default:
		return "██"
	}
}

// renderStatusPanel produces the text status panel below the map.
func renderStatusPanel(s *GameState) string {
	var sb strings.Builder
	sb.WriteString("--- Status ---\n")

	// Tick and game progress
	tick := types.Tick(0)
	coreHP := 0
	scenarioID := ""
	if s.Progress != nil {
		tick = s.Progress.CurrentTick
		coreHP = s.Progress.CoreHP
	}
	if s.Scenario != nil {
		scenarioID = s.Scenario.ID
	}
	fmt.Fprintf(&sb, "Tick: %d  Scenario: %s\n", tick, scenarioID)
	fmt.Fprintf(&sb, "Core HP: %d\n", coreHP)

	// Chi economy
	chiBalance := 0.0
	if s.EconomyEngine != nil && s.EconomyEngine.ChiPool != nil {
		chiBalance = s.EconomyEngine.ChiPool.Balance()
	}
	fmt.Fprintf(&sb, "Chi Pool: %.1f  Peak: %.1f\n", chiBalance, s.PeakChi)

	// Beast summary
	alive, stunned := countBeastStates(s.Beasts)
	fmt.Fprintf(&sb, "Beasts: %d (alive: %d, stunned: %d)\n", len(s.Beasts), alive, stunned)

	// Invasion summary
	activeWaves, completedWaves, totalInvaders, activeInvaders := countInvasionState(s.Waves)
	fmt.Fprintf(&sb, "Waves: %d total, %d active, %d completed\n", len(s.Waves), activeWaves, completedWaves)
	fmt.Fprintf(&sb, "Invaders: %d total, %d active\n", totalInvaders, activeInvaders)

	// Feng shui score
	fengShuiScore := 0.0
	if s.Cave != nil && s.RoomTypeRegistry != nil && s.ChiFlowEngine != nil {
		params := s.ScoreParams
		if params == nil {
			params = fengshui.DefaultScoreParams()
		}
		ev := fengshui.NewEvaluator(s.Cave, s.RoomTypeRegistry, params)
		fengShuiScore = ev.CaveTotal(s.ChiFlowEngine)
	}
	fmt.Fprintf(&sb, "Feng Shui: %.1f\n", fengShuiScore)

	// Economy / deficit
	fmt.Fprintf(&sb, "Deficit: %d total ticks, %d consecutive\n", s.TotalDeficitTicks, s.ConsecutiveDeficitTicks)

	// Damage stats
	fmt.Fprintf(&sb, "Damage: dealt %d, received %d\n", s.TotalDamageDealt, s.TotalDamageReceived)

	// Evolutions
	fmt.Fprintf(&sb, "Evolutions: %d\n", s.EvolutionCount)

	return sb.String()
}

// countBeastStates returns the number of alive and stunned beasts.
func countBeastStates(beasts []*senju.Beast) (alive, stunned int) {
	for _, b := range beasts {
		if b.State == senju.Stunned {
			stunned++
		} else if b.HP > 0 {
			alive++
		}
	}
	return
}

// countInvasionState returns wave and invader counts.
func countInvasionState(waves []*invasion.InvasionWave) (active, completed, totalInvaders, activeInvaders int) {
	for _, w := range waves {
		switch w.State {
		case invasion.Active:
			active++
		case invasion.Completed:
			completed++
		}
		for _, inv := range w.Invaders {
			totalInvaders++
			if inv.State != invasion.Defeated {
				activeInvaders++
			}
		}
	}
	return
}
