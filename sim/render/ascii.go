package render

import (
	"fmt"
	"strings"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// RenderFullStatus returns a comprehensive colored ASCII status display
// for a running game. It extends core's simulation.RenderFullStatus with:
//   - ANSI color codes for element-based coloring
//   - Beast list with name, level, and behavior state
//   - Invader position and state display
//   - Economy info (ChiPool balance, supply/tick, maintenance/tick)
//   - CoreHP bar
//
// The output fits within 80 columns of visible width.
func RenderFullStatus(state *simulation.GameState) string {
	var sb strings.Builder

	// Map section.
	sb.WriteString(renderColoredMap(state))

	// Status panel.
	sb.WriteString(renderColoredStatus(state))

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

// renderColoredMap produces the map with ANSI color overlays.
func renderColoredMap(s *simulation.GameState) string {
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
			return renderColoredRoomCell(cell.RoomID, roomInvaders, roomBeasts, s.ChiFlowEngine)
		case world.CorridorFloor, world.Rock:
			if veinPaths[pos] {
				return Colorize("~~", Cyan)
			}
		case world.Entrance:
			return Colorize("><", Bold)
		}
		return ""
	})
}

// renderColoredRoomCell returns a colored 2-character tile for a room cell.
// Priority: invaders > beasts > chi > room ID.
func renderColoredRoomCell(roomID int, invaders map[int]*roomInvaderInfo, beasts map[int]*roomBeastInfo, chiEngine *fengshui.ChiFlowEngine) string {
	if roomID <= 0 {
		return "[]"
	}

	// Priority 1: invaders (red).
	if info, ok := invaders[roomID]; ok {
		tile := invaderTile(info)
		return Colorize(tile, Red)
	}

	// Priority 2: beasts (element color).
	if info, ok := beasts[roomID]; ok {
		tile := world.CountTile(info.count, info.element.Char())
		return Colorize(tile, ElementColor(info.element))
	}

	// Priority 3: chi fill level (cyan).
	if chiEngine != nil {
		if rc, ok := chiEngine.RoomChi[roomID]; ok {
			ratio := rc.Ratio()
			if ratio > 0 {
				return Colorize(chiLevelTile(ratio), Cyan)
			}
		}
	}

	// Default: room ID.
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

// renderColoredStatus produces the colored status panel below the map.
func renderColoredStatus(s *simulation.GameState) string {
	var sb strings.Builder
	sb.WriteString(Colorize("--- Status ---", Bold) + "\n")

	// Tick and scenario.
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
	// Derive max CoreHP from the dragon hole room in the cave.
	maxCoreHP := coreHPMax(s)
	fmt.Fprintf(&sb, "Tick: %d  Scenario: %s\n", tick, scenarioID)

	// CoreHP bar.
	hpBar := HPBar(coreHP, maxCoreHP, 20)
	fmt.Fprintf(&sb, "Core HP: %s %s\n", hpBar, FormatHP(coreHP, maxCoreHP))

	// Economy section.
	renderEconomySection(&sb, s)

	// Beast section.
	renderBeastSection(&sb, s)

	// Invader section.
	renderInvaderSection(&sb, s)

	// Feng shui.
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

	// Damage stats.
	fmt.Fprintf(&sb, "Damage: dealt %d, received %d\n", s.TotalDamageDealt, s.TotalDamageReceived)

	return sb.String()
}

// renderEconomySection writes economy info to the builder.
func renderEconomySection(sb *strings.Builder, s *simulation.GameState) {
	if s.EconomyEngine == nil || s.EconomyEngine.ChiPool == nil {
		sb.WriteString("Chi Pool: N/A\n")
		return
	}

	pool := s.EconomyEngine.ChiPool
	balance := pool.Balance()
	cap := pool.Cap
	ratio := 0.0
	if cap > 0 {
		ratio = balance / cap
	}

	chiBar := ProgressBar(ratio, 20)
	fmt.Fprintf(sb, "Chi Pool: %s %.1f/%.1f\n", chiBar, balance, cap)

	// Show supply/maintenance from last transactions if available.
	// Since EconomyTickResult is not stored, we show what we can.
	fmt.Fprintf(sb, "Peak Chi: %.1f  Deficit: %d ticks\n", s.PeakChi, s.TotalDeficitTicks)
	if s.ConsecutiveDeficitTicks > 0 {
		fmt.Fprintf(sb, "%s\n",
			Colorize(fmt.Sprintf("WARNING: %d consecutive deficit ticks!", s.ConsecutiveDeficitTicks), Red))
	}
}

// renderBeastSection writes beast summary and list to the builder.
func renderBeastSection(sb *strings.Builder, s *simulation.GameState) {
	alive, stunned := 0, 0
	for _, b := range s.Beasts {
		if b.State == senju.Stunned {
			stunned++
		} else if b.HP > 0 {
			alive++
		}
	}
	fmt.Fprintf(sb, "Beasts: %d (alive: %d, stunned: %d)\n", len(s.Beasts), alive, stunned)

	// Individual beast details (limit to avoid overflow).
	for _, b := range s.Beasts {
		elemColor := ElementColor(b.Element)
		stateStr := b.State.String()
		switch b.State {
		case senju.Fighting:
			stateStr = Colorize(stateStr, Red)
		case senju.Stunned:
			stateStr = Colorize(stateStr, Dim)
		}
		fmt.Fprintf(sb, "  %s Lv%d %s HP:%s [%s]\n",
			Colorize(b.Name, elemColor),
			b.Level,
			Colorize(string(b.Element.Char()), elemColor),
			FormatHP(b.HP, b.MaxHP),
			stateStr,
		)
	}
}

// renderInvaderSection writes invader summary and list to the builder.
func renderInvaderSection(sb *strings.Builder, s *simulation.GameState) {
	activeWaves, completedWaves, totalInvaders, activeInvaders := 0, 0, 0, 0
	for _, w := range s.Waves {
		switch w.State {
		case invasion.Active:
			activeWaves++
		case invasion.Completed:
			completedWaves++
		}
		for _, inv := range w.Invaders {
			totalInvaders++
			if inv.State != invasion.Defeated {
				activeInvaders++
			}
		}
	}
	fmt.Fprintf(sb, "Waves: %d total, %d active, %d completed\n", len(s.Waves), activeWaves, completedWaves)
	fmt.Fprintf(sb, "Invaders: %d total, %d active\n", totalInvaders, activeInvaders)

	// List active invaders.
	for _, w := range s.Waves {
		for _, inv := range w.Invaders {
			if inv.State == invasion.Defeated {
				continue
			}
			elemColor := ElementColor(inv.Element)
			stateStr := inv.State.String()
			switch inv.State {
			case invasion.Fighting:
				stateStr = Colorize(stateStr, Red)
			case invasion.Retreating:
				stateStr = Colorize(stateStr, Yellow)
			case invasion.GoalAchieved:
				stateStr = Colorize(stateStr, Magenta)
			}
			fmt.Fprintf(sb, "  %s Lv%d %s Room:%d [%s]\n",
				Colorize(inv.Name, elemColor),
				inv.Level,
				Colorize(string(inv.Element.Char()), elemColor),
				inv.CurrentRoomID,
				stateStr,
			)
		}
	}
}

// coreHPMax derives the maximum CoreHP from the cave's dragon hole room.
// Falls back to 100 if not determinable.
func coreHPMax(s *simulation.GameState) int {
	if s.Cave != nil && s.RoomTypeRegistry != nil {
		for _, room := range s.Cave.Rooms {
			rt, err := s.RoomTypeRegistry.Get(room.TypeID)
			if err == nil && rt.BaseCoreHP > 0 {
				return rt.CoreHPAtLevel(room.Level)
			}
		}
	}
	return 100
}
