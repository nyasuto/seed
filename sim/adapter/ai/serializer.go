package ai

import (
	"encoding/json"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// SnapshotToJSON converts a GameSnapshot to a JSON-encoded byte slice.
func SnapshotToJSON(snapshot scenario.GameSnapshot) (json.RawMessage, error) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// BuildValidActions generates the list of valid actions available to the
// player in the current game state. Actions whose chi cost exceeds the
// current balance are excluded to prevent the LLM from wasting attempts.
func BuildValidActions(state *simulation.GameState) []ValidAction {
	var actions []ValidAction

	actions = append(actions, buildDigRoomActions(state)...)
	actions = append(actions, buildDigCorridorActions(state)...)
	actions = append(actions, buildSummonBeastActions(state)...)
	actions = append(actions, buildUpgradeRoomActions(state)...)
	actions = append(actions, buildEvolveBeastActions(state)...)
	actions = append(actions, buildPlaceBeastActions(state)...)

	// wait is always available
	actions = append(actions, ValidAction{Kind: "wait", Params: map[string]any{}})

	return actions
}

// buildDigRoomActions returns valid dig_room actions for each room type
// at each position where a 3x3 room can be placed.
func buildDigRoomActions(state *simulation.GameState) []ValidAction {
	// Check MaxRooms constraint.
	if state.Scenario != nil && state.Scenario.Constraints.MaxRooms > 0 {
		if len(state.Cave.Rooms) >= state.Scenario.Constraints.MaxRooms {
			return nil
		}
	}

	balance := state.EconomyEngine.ChiPool.Balance()
	allTypes := state.RoomTypeRegistry.All()

	// Filter affordable room types first.
	type affordableType struct {
		id   string
		cost float64
	}
	var affordable []affordableType
	for _, rt := range allTypes {
		cost := state.EconomyEngine.Construction.CalcRoomCost(rt.ID)
		if cost > 0 && cost <= balance {
			affordable = append(affordable, affordableType{id: rt.ID, cost: cost})
		}
	}
	if len(affordable) == 0 {
		return nil
	}

	// Find all valid placement positions for 3x3 rooms.
	const roomW, roomH = 3, 3
	grid := state.Cave.Grid
	var positions []types.Pos
	for y := 0; y <= grid.Height-roomH; y++ {
		for x := 0; x <= grid.Width-roomW; x++ {
			pos := types.Pos{X: x, Y: y}
			candidate := &world.Room{Pos: pos, Width: roomW, Height: roomH}
			if world.CanPlaceRoom(grid, candidate) {
				positions = append(positions, pos)
			}
		}
	}
	if len(positions) == 0 {
		return nil
	}

	var actions []ValidAction
	for _, at := range affordable {
		for _, pos := range positions {
			actions = append(actions, ValidAction{
				Kind: "dig_room",
				Params: map[string]any{
					"room_type_id": at.id,
					"x":            pos.X,
					"y":            pos.Y,
					"width":        roomW,
					"height":       roomH,
					"cost":         at.cost,
				},
			})
		}
	}
	return actions
}

// buildDigCorridorActions returns valid dig_corridor actions for all
// pairs of rooms that both have entrances and are not yet connected.
func buildDigCorridorActions(state *simulation.GameState) []ValidAction {
	rooms := state.Cave.Rooms
	if len(rooms) < 2 {
		return nil
	}

	// Build a set of existing corridor connections for fast lookup.
	type pair struct{ a, b int }
	connected := make(map[pair]bool)
	for _, c := range state.Cave.Corridors {
		connected[pair{c.FromRoomID, c.ToRoomID}] = true
		connected[pair{c.ToRoomID, c.FromRoomID}] = true
	}

	var actions []ValidAction
	for i := 0; i < len(rooms); i++ {
		if len(rooms[i].Entrances) == 0 {
			continue
		}
		for j := i + 1; j < len(rooms); j++ {
			if len(rooms[j].Entrances) == 0 {
				continue
			}
			if connected[pair{rooms[i].ID, rooms[j].ID}] {
				continue
			}
			actions = append(actions, ValidAction{
				Kind: "dig_corridor",
				Params: map[string]any{
					"from_room_id": rooms[i].ID,
					"to_room_id":   rooms[j].ID,
				},
			})
		}
	}
	return actions
}

// buildSummonBeastActions returns valid summon_beast actions for each
// element whose summoning cost can be afforded.
func buildSummonBeastActions(state *simulation.GameState) []ValidAction {
	balance := state.EconomyEngine.ChiPool.Balance()
	var actions []ValidAction
	for e := range types.Element(types.ElementCount) {
		cost := state.EconomyEngine.Beast.CalcSummonCost(e)
		if cost > 0 && cost <= balance {
			actions = append(actions, ValidAction{
				Kind: "summon_beast",
				Params: map[string]any{
					"element": e.String(),
					"cost":    cost,
				},
			})
		}
	}
	return actions
}

// buildUpgradeRoomActions returns valid upgrade_room actions for each
// room that can be upgraded and whose cost can be afforded.
func buildUpgradeRoomActions(state *simulation.GameState) []ValidAction {
	balance := state.EconomyEngine.ChiPool.Balance()
	var actions []ValidAction
	for _, room := range state.Cave.Rooms {
		cost := state.EconomyEngine.Construction.CalcUpgradeCost(room.TypeID, room.Level)
		if cost > 0 && cost <= balance {
			actions = append(actions, ValidAction{
				Kind: "upgrade_room",
				Params: map[string]any{
					"room_id": room.ID,
					"cost":    cost,
				},
			})
		}
	}
	return actions
}

// buildEvolveBeastActions returns valid evolve_beast actions for each
// beast that has an evolution path available.
func buildEvolveBeastActions(state *simulation.GameState) []ValidAction {
	if state.EvolutionRegistry == nil {
		return nil
	}
	balance := state.EconomyEngine.ChiPool.Balance()
	var actions []ValidAction
	for _, beast := range state.Beasts {
		paths := state.EvolutionRegistry.GetPaths(beast.SpeciesID)
		if len(paths) == 0 {
			continue
		}
		// Include each available evolution path.
		for _, path := range paths {
			if path.ChiCost > 0 && path.ChiCost > balance {
				continue
			}
			actions = append(actions, ValidAction{
				Kind: "evolve_beast",
				Params: map[string]any{
					"beast_id":       beast.ID,
					"to_species_id":  path.ToSpeciesID,
					"cost":           path.ChiCost,
				},
			})
		}
	}
	return actions
}

// buildPlaceBeastActions returns valid place_beast actions for each
// unassigned beast that can be placed in a room with capacity.
func buildPlaceBeastActions(state *simulation.GameState) []ValidAction {
	// Find unassigned beasts.
	type unassigned struct {
		speciesID string
		id        int
	}
	var free []unassigned
	for _, b := range state.Beasts {
		if b.RoomID == 0 {
			free = append(free, unassigned{speciesID: b.SpeciesID, id: b.ID})
		}
	}
	if len(free) == 0 {
		return nil
	}

	// Find rooms with beast capacity.
	type roomSlot struct {
		id int
	}
	var slots []roomSlot
	for _, room := range state.Cave.Rooms {
		rt, err := state.RoomTypeRegistry.Get(room.TypeID)
		if err != nil {
			continue
		}
		if room.HasBeastCapacity(rt) {
			slots = append(slots, roomSlot{id: room.ID})
		}
	}
	if len(slots) == 0 {
		return nil
	}

	var actions []ValidAction
	// Deduplicate by speciesID (only one action per species, since PlaceBeast
	// picks the first matching unassigned beast).
	seen := make(map[string]bool)
	for _, b := range free {
		if seen[b.speciesID] {
			continue
		}
		seen[b.speciesID] = true
		for _, s := range slots {
			actions = append(actions, ValidAction{
				Kind: "place_beast",
				Params: map[string]any{
					"species_id": b.speciesID,
					"room_id":    s.id,
				},
			})
		}
	}
	return actions
}

// StateBuilder provides methods to build the game state needed by the
// AI adapter from a GameServer's engine. It wraps a function that
// returns the current SimulationEngine.
type StateBuilder struct {
	engineFn func() *simulation.SimulationEngine
}

// NewStateBuilder creates a StateBuilder. The engineFn should return
// the current SimulationEngine (typically GameServer.Engine).
func NewStateBuilder(engineFn func() *simulation.SimulationEngine) *StateBuilder {
	return &StateBuilder{engineFn: engineFn}
}

// BuildStateMessage constructs a complete StateMessage for the current
// game state. Returns nil if no engine is active.
func (sb *StateBuilder) BuildStateMessage(snapshot scenario.GameSnapshot) (*StateMessage, error) {
	engine := sb.engineFn()
	if engine == nil {
		return nil, nil
	}

	snapJSON, err := SnapshotToJSON(snapshot)
	if err != nil {
		return nil, err
	}

	validActions := BuildValidActions(engine.State)
	msg := NewStateMessage(int(snapshot.Tick), snapJSON, validActions)
	return &msg, nil
}

