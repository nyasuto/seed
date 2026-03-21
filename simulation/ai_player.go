package simulation

import (
	"math"
	"slices"
	"sort"

	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// AIPlayer is the interface for automated game players. Each tick the
// simulation calls DecideActions with the current snapshot to obtain the
// player actions to submit.
type AIPlayer interface {
	// DecideActions returns the actions to perform for the current tick.
	DecideActions(snapshot scenario.GameSnapshot) []PlayerAction
}

// SimpleAIPlayer implements a minimal strategy:
//  1. If chi pool has余裕, build the cheapest affordable room.
//  2. If a beast room has capacity and there are unassigned beasts, place one.
//  3. If there are no beasts, summon the cheapest one.
//  4. Otherwise, do nothing.
type SimpleAIPlayer struct {
	state *GameState
}

// NewSimpleAIPlayer creates a SimpleAIPlayer that decides actions based on
// the given game state. The state must remain the same instance used by the
// SimulationEngine throughout the game.
func NewSimpleAIPlayer(state *GameState) *SimpleAIPlayer {
	return &SimpleAIPlayer{state: state}
}

// DecideActions implements AIPlayer.
func (ai *SimpleAIPlayer) DecideActions(snapshot scenario.GameSnapshot) []PlayerAction {
	var actions []PlayerAction

	// 1. Place unassigned beasts into rooms with capacity.
	if action := ai.tryPlaceBeast(); action != nil {
		actions = append(actions, action)
	}

	// 2. If chi is sufficient, build the cheapest room.
	if action := ai.tryBuildRoom(snapshot); action != nil {
		actions = append(actions, action)
	}

	// 3. If beast rooms have space but no unassigned beasts, summon one.
	if action := ai.trySummonBeast(snapshot); action != nil {
		actions = append(actions, action)
	}

	if len(actions) == 0 {
		return []PlayerAction{NoAction{}}
	}
	return actions
}

// tryBuildRoom attempts to find the cheapest affordable room type and a valid
// position to build it. Returns nil if no room can be built.
func (ai *SimpleAIPlayer) tryBuildRoom(snapshot scenario.GameSnapshot) PlayerAction {
	s := ai.state
	construction := s.EconomyEngine.Construction
	balance := s.EconomyEngine.ChiPool.Balance()

	// Collect buildable room types sorted by cost.
	type roomCandidate struct {
		typeID string
		cost   float64
	}
	var candidates []roomCandidate
	for typeID, cost := range construction.RoomCost {
		if cost > 0 && cost <= balance {
			// Only consider room types that exist in the registry.
			if _, err := s.RoomTypeRegistry.Get(typeID); err != nil {
				continue
			}
			candidates = append(candidates, roomCandidate{typeID: typeID, cost: cost})
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].cost != candidates[j].cost {
			return candidates[i].cost < candidates[j].cost
		}
		return candidates[i].typeID < candidates[j].typeID
	})

	// Try to find a valid placement position for the cheapest room type.
	roomWidth, roomHeight := 3, 3
	for _, c := range candidates {
		pos, ok := findPlacement(s.Cave, roomWidth, roomHeight)
		if ok {
			return DigRoomAction{
				RoomTypeID: c.typeID,
				Pos:        pos,
				Width:      roomWidth,
				Height:     roomHeight,
			}
		}
	}
	return nil
}

// tryPlaceBeast looks for an unassigned beast and a room with capacity to
// place it in. Returns nil if no placement is possible.
func (ai *SimpleAIPlayer) tryPlaceBeast() PlayerAction {
	s := ai.state

	// Find unassigned beasts.
	var unassigned []*senju.Beast
	for _, b := range s.Beasts {
		if b.RoomID == 0 && b.State != senju.Stunned {
			unassigned = append(unassigned, b)
		}
	}
	if len(unassigned) == 0 {
		return nil
	}

	// Find a room with beast capacity, preferring rooms close to dragon hole.
	dragonHoleRoomID := findDragonHoleRoomID(s)
	bestRoom := findBestPlacementRoom(s, dragonHoleRoomID)
	if bestRoom == nil {
		return nil
	}

	return PlaceBeastAction{
		SpeciesID: unassigned[0].SpeciesID,
		RoomID:    bestRoom.ID,
	}
}

// trySummonBeast summons the cheapest beast element if there are rooms with
// capacity but no unassigned beasts. Returns nil if summoning is not needed or
// not affordable.
func (ai *SimpleAIPlayer) trySummonBeast(snapshot scenario.GameSnapshot) PlayerAction {
	s := ai.state

	// Check if we have unassigned beasts already.
	for _, b := range s.Beasts {
		if b.RoomID == 0 && b.State != senju.Stunned {
			return nil
		}
	}

	// Check if any room has beast capacity.
	hasCapacity := false
	for _, room := range s.Cave.Rooms {
		rt, err := s.RoomTypeRegistry.Get(room.TypeID)
		if err != nil {
			continue
		}
		if room.HasBeastCapacity(rt) {
			hasCapacity = true
			break
		}
	}
	if !hasCapacity {
		return nil
	}

	// Find cheapest element to summon.
	balance := s.EconomyEngine.ChiPool.Balance()
	bestCost := math.MaxFloat64
	bestElement := types.Wood
	found := false
	for elem, cost := range s.EconomyEngine.Beast.SummonCostByElement {
		if cost > 0 && cost <= balance && cost < bestCost {
			bestCost = cost
			bestElement = elem
			found = true
		}
	}
	if !found {
		return nil
	}

	return SummonBeastAction{Element: bestElement}
}

// findDragonHoleRoomID returns the room ID of the dragon hole (core room).
// Returns 0 if not found.
func findDragonHoleRoomID(s *GameState) int {
	for _, room := range s.Cave.Rooms {
		if room.CoreHP > 0 {
			return room.ID
		}
	}
	return 0
}

// findBestPlacementRoom finds a room with beast capacity, preferring rooms
// closer to the dragon hole for defensive positioning.
func findBestPlacementRoom(s *GameState, dragonHoleRoomID int) *world.Room {
	var dragonHolePos types.Pos
	if dragonHoleRoomID > 0 {
		if r := s.Cave.RoomByID(dragonHoleRoomID); r != nil {
			dragonHolePos = r.Pos
		}
	}

	var best *world.Room
	bestDist := math.MaxInt32
	for _, room := range s.Cave.Rooms {
		rt, err := s.RoomTypeRegistry.Get(room.TypeID)
		if err != nil {
			continue
		}
		if !room.HasBeastCapacity(rt) {
			continue
		}
		dist := room.Pos.Distance(dragonHolePos)
		if dist < bestDist {
			bestDist = dist
			best = room
		}
	}
	return best
}

// findPlacement scans the cave grid for a position where a room of the given
// size can be placed (all cells must be Rock).
func findPlacement(cave *world.Cave, w, h int) (types.Pos, bool) {
	grid := cave.Grid
	for y := 1; y < grid.Height-h-1; y++ {
		for x := 1; x < grid.Width-w-1; x++ {
			pos := types.Pos{X: x, Y: y}
			candidate := &world.Room{Pos: pos, Width: w, Height: h}
			if world.CanPlaceRoom(grid, candidate) {
				return pos, true
			}
		}
	}
	return types.Pos{}, false
}

// RandomAIPlayer selects random actions each tick using the provided RNG.
// It is intended for fuzz-like balance testing to verify that arbitrary
// actions do not crash the simulation.
type RandomAIPlayer struct {
	state *GameState
	rng   types.RNG
}

// NewRandomAIPlayer creates a RandomAIPlayer. The rng should be a separate
// instance from the engine's RNG to avoid disturbing determinism of the
// simulation itself.
func NewRandomAIPlayer(state *GameState, rng types.RNG) *RandomAIPlayer {
	return &RandomAIPlayer{state: state, rng: rng}
}

// DecideActions implements AIPlayer. It picks a random action type and attempts
// to construct a valid action. If the chosen action cannot be performed, it
// falls back to NoAction.
func (ai *RandomAIPlayer) DecideActions(snapshot scenario.GameSnapshot) []PlayerAction {
	// 7 action types: 0=no_action, 1=dig_room, 2=summon, 3=place, 4=upgrade, 5=evolve, 6=corridor
	choice := ai.rng.Intn(7)
	switch choice {
	case 1:
		if a := ai.randomDigRoom(); a != nil {
			return []PlayerAction{a}
		}
	case 2:
		if a := ai.randomSummon(); a != nil {
			return []PlayerAction{a}
		}
	case 3:
		if a := ai.randomPlace(); a != nil {
			return []PlayerAction{a}
		}
	case 4:
		if a := ai.randomUpgrade(); a != nil {
			return []PlayerAction{a}
		}
	case 5:
		if a := ai.randomEvolve(); a != nil {
			return []PlayerAction{a}
		}
	case 6:
		if a := ai.randomCorridor(); a != nil {
			return []PlayerAction{a}
		}
	}
	return []PlayerAction{NoAction{}}
}

func (ai *RandomAIPlayer) randomDigRoom() PlayerAction {
	s := ai.state
	construction := s.EconomyEngine.Construction
	balance := s.EconomyEngine.ChiPool.Balance()

	// Collect affordable room types.
	var typeIDs []string
	for typeID, cost := range construction.RoomCost {
		if cost > 0 && cost <= balance {
			if _, err := s.RoomTypeRegistry.Get(typeID); err != nil {
				continue
			}
			typeIDs = append(typeIDs, typeID)
		}
	}
	if len(typeIDs) == 0 {
		return nil
	}
	sort.Strings(typeIDs) // deterministic order
	typeID := typeIDs[ai.rng.Intn(len(typeIDs))]

	pos, ok := findPlacement(s.Cave, 3, 3)
	if !ok {
		return nil
	}
	return DigRoomAction{RoomTypeID: typeID, Pos: pos, Width: 3, Height: 3}
}

func (ai *RandomAIPlayer) randomSummon() PlayerAction {
	s := ai.state
	balance := s.EconomyEngine.ChiPool.Balance()

	var elements []types.Element
	for elem, cost := range s.EconomyEngine.Beast.SummonCostByElement {
		if cost > 0 && cost <= balance {
			elements = append(elements, elem)
		}
	}
	if len(elements) == 0 {
		return nil
	}
	slices.Sort(elements)
	return SummonBeastAction{Element: elements[ai.rng.Intn(len(elements))]}
}

func (ai *RandomAIPlayer) randomPlace() PlayerAction {
	s := ai.state

	// Find unassigned beasts.
	var unassigned []*senju.Beast
	for _, b := range s.Beasts {
		if b.RoomID == 0 && b.State != senju.Stunned {
			unassigned = append(unassigned, b)
		}
	}
	if len(unassigned) == 0 {
		return nil
	}

	// Find rooms with capacity.
	var rooms []*world.Room
	for _, room := range s.Cave.Rooms {
		rt, err := s.RoomTypeRegistry.Get(room.TypeID)
		if err != nil {
			continue
		}
		if room.HasBeastCapacity(rt) {
			rooms = append(rooms, room)
		}
	}
	if len(rooms) == 0 {
		return nil
	}

	beast := unassigned[ai.rng.Intn(len(unassigned))]
	room := rooms[ai.rng.Intn(len(rooms))]
	return PlaceBeastAction{SpeciesID: beast.SpeciesID, RoomID: room.ID}
}

func (ai *RandomAIPlayer) randomUpgrade() PlayerAction {
	s := ai.state
	if len(s.Cave.Rooms) == 0 {
		return nil
	}
	room := s.Cave.Rooms[ai.rng.Intn(len(s.Cave.Rooms))]
	cost := s.EconomyEngine.Construction.CalcUpgradeCost(room.TypeID, room.Level)
	if cost <= 0 || !s.EconomyEngine.ChiPool.CanAfford(cost) {
		return nil
	}
	return UpgradeRoomAction{RoomID: room.ID}
}

func (ai *RandomAIPlayer) randomEvolve() PlayerAction {
	s := ai.state
	if len(s.Beasts) == 0 || s.EvolutionRegistry == nil {
		return nil
	}

	// Collect beasts with evolution paths.
	var candidates []*senju.Beast
	for _, b := range s.Beasts {
		if b.State == senju.Stunned || b.RoomID == 0 {
			continue
		}
		paths := s.EvolutionRegistry.GetPaths(b.SpeciesID)
		if len(paths) > 0 {
			candidates = append(candidates, b)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	return EvolveBeastAction{BeastID: candidates[ai.rng.Intn(len(candidates))].ID}
}

func (ai *RandomAIPlayer) randomCorridor() PlayerAction {
	s := ai.state
	rooms := s.Cave.Rooms
	if len(rooms) < 2 {
		return nil
	}
	i := ai.rng.Intn(len(rooms))
	j := ai.rng.Intn(len(rooms))
	if i == j {
		return nil
	}
	from := rooms[i]
	to := rooms[j]
	if len(from.Entrances) == 0 || len(to.Entrances) == 0 {
		return nil
	}
	return DigCorridorAction{FromRoomID: from.ID, ToRoomID: to.ID}
}
