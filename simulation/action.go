package simulation

import (
	"errors"
	"fmt"

	"github.com/ponpoko/chaosseed-core/economy"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// Errors returned by ValidateAction.
var (
	// ErrUnknownAction is returned when the action type is not recognized.
	ErrUnknownAction = errors.New("unknown action type")
	// ErrRoomTypeNotFound is returned when the room type ID is not registered.
	ErrRoomTypeNotFound = errors.New("room type not found")
	// ErrBeastNotFound is returned when a beast ID is not found in the game state.
	ErrBeastNotFound = errors.New("beast not found")
	// ErrNoEvolutionPath is returned when no evolution path is available for a beast.
	ErrNoEvolutionPath = errors.New("no evolution path available")
)

// PlayerAction is the interface that all player actions must implement.
// Each tick, a player (or AI) may submit one or more actions to the
// simulation engine for validation and execution.
type PlayerAction interface {
	// ActionType returns a string identifier for this action type.
	ActionType() string
}

// ActionResult captures the outcome of applying a player action.
type ActionResult struct {
	// Success indicates whether the action was executed successfully.
	Success bool
	// Cost is the chi cost that was spent (0 if no cost or failed).
	Cost float64
	// Description is a human-readable summary of what happened.
	Description string
}

// DigRoomAction requests digging a new room in the cave.
type DigRoomAction struct {
	// RoomTypeID is the type of room to build (e.g. "dragon_lair").
	RoomTypeID string `json:"room_type_id"`
	// Pos is the top-left position where the room will be placed.
	Pos types.Pos `json:"pos"`
	// Width is the room width in cells.
	Width int `json:"width"`
	// Height is the room height in cells.
	Height int `json:"height"`
}

// ActionType returns the action type identifier.
func (a DigRoomAction) ActionType() string { return "dig_room" }

// DigCorridorAction requests connecting two rooms with a corridor.
type DigCorridorAction struct {
	// FromRoomID is the source room ID.
	FromRoomID int `json:"from_room_id"`
	// ToRoomID is the destination room ID.
	ToRoomID int `json:"to_room_id"`
}

// ActionType returns the action type identifier.
func (a DigCorridorAction) ActionType() string { return "dig_corridor" }

// PlaceBeastAction requests placing an existing beast into a room.
type PlaceBeastAction struct {
	// SpeciesID is the species of beast to place.
	SpeciesID string `json:"species_id"`
	// RoomID is the target room ID.
	RoomID int `json:"room_id"`
}

// ActionType returns the action type identifier.
func (a PlaceBeastAction) ActionType() string { return "place_beast" }

// UpgradeRoomAction requests upgrading a room to the next level.
type UpgradeRoomAction struct {
	// RoomID is the room to upgrade.
	RoomID int `json:"room_id"`
}

// ActionType returns the action type identifier.
func (a UpgradeRoomAction) ActionType() string { return "upgrade_room" }

// SummonBeastAction requests summoning a new beast of the given element.
type SummonBeastAction struct {
	// Element is the element of the beast to summon.
	Element types.Element `json:"element"`
}

// ActionType returns the action type identifier.
func (a SummonBeastAction) ActionType() string { return "summon_beast" }

// EvolveBeastAction requests evolving a beast along its evolution path.
type EvolveBeastAction struct {
	// BeastID is the ID of the beast to evolve.
	BeastID int `json:"beast_id"`
}

// ActionType returns the action type identifier.
func (a EvolveBeastAction) ActionType() string { return "evolve_beast" }

// NoAction represents a deliberate choice to do nothing this tick.
type NoAction struct{}

// ActionType returns the action type identifier.
func (a NoAction) ActionType() string { return "no_action" }

// ValidateAction checks whether a player action can be executed against the
// current game state without modifying anything. It returns nil if the action
// is valid, or a descriptive error explaining why it cannot be performed.
func ValidateAction(action PlayerAction, state *GameState) error {
	switch a := action.(type) {
	case DigRoomAction:
		return validateDigRoom(a, state)
	case DigCorridorAction:
		return validateDigCorridor(a, state)
	case PlaceBeastAction:
		return validatePlaceBeast(a, state)
	case UpgradeRoomAction:
		return validateUpgradeRoom(a, state)
	case SummonBeastAction:
		return validateSummonBeast(a, state)
	case EvolveBeastAction:
		return validateEvolveBeast(a, state)
	case NoAction:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownAction, action.ActionType())
	}
}

func validateDigRoom(a DigRoomAction, state *GameState) error {
	// Check room type exists.
	if _, err := state.RoomTypeRegistry.Get(a.RoomTypeID); err != nil {
		return fmt.Errorf("%w: %s", ErrRoomTypeNotFound, a.RoomTypeID)
	}

	// Check placement is valid (no overlap, within bounds).
	tempRoom := &world.Room{
		Pos:    a.Pos,
		Width:  a.Width,
		Height: a.Height,
	}
	if !world.CanPlaceRoom(state.Cave.Grid, tempRoom) {
		return fmt.Errorf("cannot place room at (%d,%d) size %dx%d", a.Pos.X, a.Pos.Y, a.Width, a.Height)
	}

	// Check cost affordability.
	cost := state.EconomyEngine.Construction.CalcRoomCost(a.RoomTypeID)
	if !state.EconomyEngine.ChiPool.CanAfford(cost) {
		return fmt.Errorf("insufficient chi: need %.1f, have %.1f", cost, state.EconomyEngine.ChiPool.Balance())
	}

	return nil
}

func validateDigCorridor(a DigCorridorAction, state *GameState) error {
	// Check both rooms exist.
	room1 := state.Cave.RoomByID(a.FromRoomID)
	if room1 == nil {
		return fmt.Errorf("from room %d: %w", a.FromRoomID, world.ErrRoomNotFound)
	}
	room2 := state.Cave.RoomByID(a.ToRoomID)
	if room2 == nil {
		return fmt.Errorf("to room %d: %w", a.ToRoomID, world.ErrRoomNotFound)
	}

	// Check both rooms have entrances.
	if len(room1.Entrances) == 0 {
		return fmt.Errorf("from room %d: %w", a.FromRoomID, world.ErrNoEntrance)
	}
	if len(room2.Entrances) == 0 {
		return fmt.Errorf("to room %d: %w", a.ToRoomID, world.ErrNoEntrance)
	}

	// Cost check requires knowing the path length, which we can't determine
	// without actually pathfinding. We defer the full cost check to ApplyAction.
	return nil
}

func validatePlaceBeast(a PlaceBeastAction, state *GameState) error {
	// Find the beast by species ID among unassigned beasts.
	found := false
	for _, b := range state.Beasts {
		if b.SpeciesID == a.SpeciesID && b.RoomID == 0 {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("no unassigned beast with species %s", a.SpeciesID)
	}

	// Check room exists.
	room := state.Cave.RoomByID(a.RoomID)
	if room == nil {
		return fmt.Errorf("room %d: %w", a.RoomID, world.ErrRoomNotFound)
	}

	// Check room type allows beasts and has capacity.
	rt, err := state.RoomTypeRegistry.Get(room.TypeID)
	if err != nil {
		return fmt.Errorf("room type %s: %w", room.TypeID, ErrRoomTypeNotFound)
	}
	if !room.HasBeastCapacity(rt) {
		return fmt.Errorf("room %d is at beast capacity", a.RoomID)
	}

	return nil
}

func validateUpgradeRoom(a UpgradeRoomAction, state *GameState) error {
	// Check room exists.
	room := state.Cave.RoomByID(a.RoomID)
	if room == nil {
		return fmt.Errorf("room %d: %w", a.RoomID, world.ErrRoomNotFound)
	}

	// Check cost affordability.
	cost := state.EconomyEngine.Construction.CalcUpgradeCost(room.TypeID, room.Level)
	if cost <= 0 {
		return fmt.Errorf("cannot upgrade room type %s", room.TypeID)
	}
	if !state.EconomyEngine.ChiPool.CanAfford(cost) {
		return fmt.Errorf("insufficient chi: need %.1f, have %.1f", cost, state.EconomyEngine.ChiPool.Balance())
	}

	return nil
}

func validateSummonBeast(a SummonBeastAction, state *GameState) error {
	// Check element cost exists.
	cost := state.EconomyEngine.Beast.CalcSummonCost(a.Element)
	if cost <= 0 {
		return fmt.Errorf("unknown element for summoning: %v", a.Element)
	}

	// Check cost affordability.
	if !state.EconomyEngine.ChiPool.CanAfford(cost) {
		return fmt.Errorf("insufficient chi: need %.1f, have %.1f", cost, state.EconomyEngine.ChiPool.Balance())
	}

	return nil
}

func validateEvolveBeast(a EvolveBeastAction, state *GameState) error {
	// Find the beast.
	var beast *senju.Beast
	for _, b := range state.Beasts {
		if b.ID == a.BeastID {
			beast = b
			break
		}
	}
	if beast == nil {
		return fmt.Errorf("beast %d: %w", a.BeastID, ErrBeastNotFound)
	}

	// Check evolution registry exists and has paths.
	if state.EvolutionRegistry == nil {
		return fmt.Errorf("beast %d: %w", a.BeastID, ErrNoEvolutionPath)
	}

	paths := state.EvolutionRegistry.GetPaths(beast.SpeciesID)
	if len(paths) == 0 {
		return fmt.Errorf("beast %d (species %s): %w", a.BeastID, beast.SpeciesID, ErrNoEvolutionPath)
	}

	return nil
}

// ApplyAction validates and executes a player action against the current game
// state, returning an ActionResult describing the outcome. The state is mutated
// on success. On validation failure the state is unchanged and an error is returned.
func ApplyAction(action PlayerAction, state *GameState) (ActionResult, error) {
	if err := ValidateAction(action, state); err != nil {
		return ActionResult{}, err
	}

	switch a := action.(type) {
	case DigRoomAction:
		return applyDigRoom(a, state)
	case DigCorridorAction:
		return applyDigCorridor(a, state)
	case PlaceBeastAction:
		return applyPlaceBeast(a, state)
	case UpgradeRoomAction:
		return applyUpgradeRoom(a, state)
	case SummonBeastAction:
		return applySummonBeast(a, state)
	case EvolveBeastAction:
		return applyEvolveBeast(a, state)
	case NoAction:
		return ActionResult{Success: true, Description: "no action taken"}, nil
	default:
		return ActionResult{}, fmt.Errorf("%w: %s", ErrUnknownAction, action.ActionType())
	}
}

func currentTick(state *GameState) types.Tick {
	if state.Progress != nil {
		return state.Progress.CurrentTick
	}
	return 0
}

func applyDigRoom(a DigRoomAction, state *GameState) (ActionResult, error) {
	tick := currentTick(state)

	cost, err := state.EconomyEngine.TryBuildRoom(a.RoomTypeID, tick)
	if err != nil {
		return ActionResult{}, fmt.Errorf("dig room: %w", err)
	}

	room, err := state.Cave.AddRoom(a.RoomTypeID, a.Pos, a.Width, a.Height, nil)
	if err != nil {
		return ActionResult{}, fmt.Errorf("dig room: %w", err)
	}

	return ActionResult{
		Success:     true,
		Cost:        cost,
		Description: fmt.Sprintf("dug room %d (%s) at (%d,%d)", room.ID, a.RoomTypeID, a.Pos.X, a.Pos.Y),
	}, nil
}

func applyDigCorridor(a DigCorridorAction, state *GameState) (ActionResult, error) {
	tick := currentTick(state)

	corridor, err := state.Cave.ConnectRooms(a.FromRoomID, a.ToRoomID)
	if err != nil {
		return ActionResult{}, fmt.Errorf("dig corridor: %w", err)
	}

	cost, err := state.EconomyEngine.TryDigCorridor(len(corridor.Path), tick)
	if err != nil {
		return ActionResult{}, fmt.Errorf("dig corridor: %w", err)
	}

	return ActionResult{
		Success:     true,
		Cost:        cost,
		Description: fmt.Sprintf("dug corridor %d from room %d to room %d (length %d)", corridor.ID, a.FromRoomID, a.ToRoomID, len(corridor.Path)),
	}, nil
}

func applyPlaceBeast(a PlaceBeastAction, state *GameState) (ActionResult, error) {
	// Find the first unassigned beast with the matching species.
	var target *senju.Beast
	for _, b := range state.Beasts {
		if b.SpeciesID == a.SpeciesID && b.RoomID == 0 {
			target = b
			break
		}
	}
	if target == nil {
		return ActionResult{}, fmt.Errorf("no unassigned beast with species %s", a.SpeciesID)
	}

	// Assign beast to room.
	target.RoomID = a.RoomID
	room := state.Cave.RoomByID(a.RoomID)
	room.BeastIDs = append(room.BeastIDs, target.ID)

	return ActionResult{
		Success:     true,
		Cost:        0,
		Description: fmt.Sprintf("placed beast %d (%s) in room %d", target.ID, a.SpeciesID, a.RoomID),
	}, nil
}

func applyUpgradeRoom(a UpgradeRoomAction, state *GameState) (ActionResult, error) {
	tick := currentTick(state)
	room := state.Cave.RoomByID(a.RoomID)

	cost, err := state.EconomyEngine.TryUpgradeRoom(room.TypeID, room.Level, tick)
	if err != nil {
		return ActionResult{}, fmt.Errorf("upgrade room: %w", err)
	}

	room.Level++

	return ActionResult{
		Success:     true,
		Cost:        cost,
		Description: fmt.Sprintf("upgraded room %d to level %d", a.RoomID, room.Level),
	}, nil
}

func applySummonBeast(a SummonBeastAction, state *GameState) (ActionResult, error) {
	tick := currentTick(state)

	cost, err := state.EconomyEngine.TrySummonBeast(a.Element, tick)
	if err != nil {
		return ActionResult{}, fmt.Errorf("summon beast: %w", err)
	}

	// Find a species of the requested element.
	var species *senju.Species
	for _, s := range state.SpeciesRegistry.All() {
		if s.Element == a.Element {
			species = s
			break
		}
	}
	if species == nil {
		return ActionResult{}, fmt.Errorf("no species found for element %v", a.Element)
	}

	beast := senju.NewBeast(state.NextBeastID, species, tick)
	state.NextBeastID++
	state.Beasts = append(state.Beasts, beast)

	return ActionResult{
		Success:     true,
		Cost:        cost,
		Description: fmt.Sprintf("summoned beast %d (%s, %v)", beast.ID, species.ID, a.Element),
	}, nil
}

func applyEvolveBeast(a EvolveBeastAction, state *GameState) (ActionResult, error) {
	tick := currentTick(state)

	// Find the beast.
	var beast *senju.Beast
	for _, b := range state.Beasts {
		if b.ID == a.BeastID {
			beast = b
			break
		}
	}
	if beast == nil {
		return ActionResult{}, fmt.Errorf("beast %d: %w", a.BeastID, ErrBeastNotFound)
	}

	// Get the first available evolution path.
	paths := state.EvolutionRegistry.GetPaths(beast.SpeciesID)
	if len(paths) == 0 {
		return ActionResult{}, fmt.Errorf("beast %d: %w", a.BeastID, ErrNoEvolutionPath)
	}
	path := &paths[0]

	// Pay the evolution chi cost.
	if path.ChiCost > 0 {
		if err := state.EconomyEngine.ChiPool.Withdraw(path.ChiCost, economy.BeastSummon, fmt.Sprintf("evolve beast %d", a.BeastID), tick); err != nil {
			return ActionResult{}, fmt.Errorf("evolve beast: %w", err)
		}
	}

	oldSpecies := beast.SpeciesID
	if err := senju.Evolve(beast, path, state.SpeciesRegistry); err != nil {
		return ActionResult{}, fmt.Errorf("evolve beast: %w", err)
	}

	return ActionResult{
		Success:     true,
		Cost:        path.ChiCost,
		Description: fmt.Sprintf("evolved beast %d from %s to %s", a.BeastID, oldSpecies, beast.SpeciesID),
	}, nil
}
