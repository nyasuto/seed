package render

import (
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/economy"
	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// newTestState creates a minimal GameState for rendering tests.
func newTestState() *simulation.GameState {
	grid, _ := world.NewGrid(5, 3)
	cave := &world.Cave{
		Grid: grid,
	}

	// Place a small room at position (1,1).
	room := &world.Room{
		ID:     1,
		TypeID: "fire_room",
		Pos:    types.Pos{X: 1, Y: 1},
		Width:  2,
		Height: 1,
		Level:  1,
	}
	cave.Rooms = append(cave.Rooms, room)

	// Set grid cells for the room.
	for x := 1; x <= 2; x++ {
		_ = grid.Set(types.Pos{X: x, Y: 1}, world.Cell{Type: world.RoomFloor, RoomID: 1})
	}
	// Set an entrance.
	_ = grid.Set(types.Pos{X: 0, Y: 1}, world.Cell{Type: world.Entrance})

	return &simulation.GameState{
		Cave: cave,
		Progress: &scenario.ScenarioProgress{
			CurrentTick: 42,
			CoreHP:      85,
		},
		Scenario: &scenario.Scenario{
			ID:         "test_scenario",
			Difficulty: "standard",
		},
		Beasts: []*senju.Beast{
			{
				ID:      1,
				Name:    "FireDrake",
				Element: types.Fire,
				RoomID:  1,
				Level:   3,
				HP:      45,
				MaxHP:   50,
				State:   senju.Idle,
			},
			{
				ID:      2,
				Name:    "WaterSprite",
				Element: types.Water,
				RoomID:  1,
				Level:   2,
				HP:      0,
				MaxHP:   30,
				State:   senju.Stunned,
			},
		},
		Waves: []*invasion.InvasionWave{
			{
				ID:    1,
				State: invasion.Active,
				Invaders: []*invasion.Invader{
					{
						ID:            1,
						Name:          "Warrior",
						Element:       types.Metal,
						Level:         5,
						CurrentRoomID: 1,
						State:         invasion.Advancing,
					},
				},
			},
		},
	}
}

func TestRenderFullStatus_ContainsANSI(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	// Should contain ANSI escape sequences.
	if !strings.Contains(output, "\033[") {
		t.Error("output should contain ANSI escape sequences")
	}
}

func TestRenderFullStatus_ContainsStatusHeader(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	if !strings.Contains(output, "Status") {
		t.Error("output should contain 'Status' header")
	}
}

func TestRenderFullStatus_ContainsTick(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	if !strings.Contains(output, "Tick: 42") {
		t.Error("output should contain 'Tick: 42'")
	}
}

func TestRenderFullStatus_ContainsCoreHP(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	if !strings.Contains(output, "Core HP:") {
		t.Error("output should contain 'Core HP:'")
	}
	// HP bar should be present.
	stripped := StripANSI(output)
	if !strings.Contains(stripped, "85/100") {
		t.Error("output should contain '85/100'")
	}
}

func TestRenderFullStatus_ContainsBeastList(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	if !strings.Contains(output, "FireDrake") {
		t.Error("output should contain beast name 'FireDrake'")
	}
	if !strings.Contains(output, "WaterSprite") {
		t.Error("output should contain beast name 'WaterSprite'")
	}
	if !strings.Contains(output, "Lv3") {
		t.Error("output should contain beast level 'Lv3'")
	}
}

func TestRenderFullStatus_ContainsBeastState(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	if !strings.Contains(output, "Idle") {
		t.Error("output should contain beast state 'Idle'")
	}
	if !strings.Contains(output, "Stunned") {
		t.Error("output should contain beast state 'Stunned'")
	}
}

func TestRenderFullStatus_ContainsInvaders(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	if !strings.Contains(output, "Warrior") {
		t.Error("output should contain invader name 'Warrior'")
	}
	if !strings.Contains(output, "Advancing") {
		t.Error("output should contain invader state 'Advancing'")
	}
}

func TestRenderFullStatus_InvaderColoredRed(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	// Invader tiles should be red on the map.
	if !strings.Contains(output, Red) {
		t.Error("output should contain red color for invaders")
	}
}

func TestRenderFullStatus_BeastColoredByElement(t *testing.T) {
	// Remove invaders so beasts show on map.
	state := newTestState()
	state.Waves = nil

	output := RenderFullStatus(state)

	// Fire beast should produce red color.
	if !strings.Contains(output, Red) {
		t.Error("fire beast should produce red color")
	}
}

func TestRenderFullStatus_LayoutWidth(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	for i, line := range strings.Split(output, "\n") {
		visible := VisibleWidth(line)
		if visible > 80 {
			t.Errorf("line %d exceeds 80 columns: visible width=%d, line=%q",
				i+1, visible, StripANSI(line))
		}
	}
}

func TestRenderFullStatus_WithEconomy(t *testing.T) {
	state := newTestState()

	chiPool := economy.NewChiPool(200.0)
	_ = chiPool.Deposit(50.0, economy.Supply, "test", 0)
	state.EconomyEngine = &economy.EconomyEngine{
		ChiPool: chiPool,
	}

	output := RenderFullStatus(state)

	if !strings.Contains(output, "Chi Pool:") {
		t.Error("output should contain 'Chi Pool:'")
	}
	stripped := StripANSI(output)
	if !strings.Contains(stripped, "50.0/200.0") {
		t.Error("output should show chi balance/cap")
	}
}

func TestRenderFullStatus_WithChiFlow(t *testing.T) {
	state := newTestState()

	chiEngine := &fengshui.ChiFlowEngine{
		RoomChi: map[int]*fengshui.RoomChi{
			1: {RoomID: 1, Current: 50, Capacity: 100, Element: types.Fire},
		},
	}
	state.ChiFlowEngine = chiEngine

	// Remove beasts and invaders so chi shows on map.
	state.Beasts = nil
	state.Waves = nil

	output := RenderFullStatus(state)

	// Chi overlay should produce cyan color.
	if !strings.Contains(output, Cyan) {
		t.Error("chi overlay should produce cyan color")
	}
}

func TestRenderFullStatus_DragonVeinColored(t *testing.T) {
	state := newTestState()

	chiEngine := &fengshui.ChiFlowEngine{
		RoomChi: map[int]*fengshui.RoomChi{},
		Veins: []*fengshui.DragonVein{
			{
				ID:       1,
				Element:  types.Water,
				Path:     []types.Pos{{X: 3, Y: 0}},
				FlowRate: 1.0,
			},
		},
	}
	state.ChiFlowEngine = chiEngine

	output := RenderFullStatus(state)

	// Dragon vein path should be colored cyan.
	if !strings.Contains(output, Cyan) {
		t.Error("dragon vein path should produce cyan color")
	}
}

func TestRenderFullStatus_NilState(t *testing.T) {
	state := &simulation.GameState{}
	output := RenderFullStatus(state)

	// Should not panic and should produce some output.
	if !strings.Contains(output, "Status") {
		t.Error("output should contain 'Status' even with minimal state")
	}
}

func TestRenderFullStatus_DeficitWarning(t *testing.T) {
	state := newTestState()
	state.ConsecutiveDeficitTicks = 5
	state.TotalDeficitTicks = 10
	state.EconomyEngine = &economy.EconomyEngine{
		ChiPool: economy.NewChiPool(100.0),
	}

	output := RenderFullStatus(state)

	if !strings.Contains(output, "WARNING") {
		t.Error("output should contain deficit WARNING")
	}
	if !strings.Contains(output, Red) {
		t.Error("deficit warning should be red")
	}
}

func TestRenderColoredMap_EntranceColored(t *testing.T) {
	state := newTestState()
	output := RenderFullStatus(state)

	// Entrance should be bold.
	if !strings.Contains(output, Bold) {
		t.Error("entrance should be rendered with bold")
	}
}
