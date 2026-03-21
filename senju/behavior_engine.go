package senju

import (
	"errors"
	"fmt"
	"sort"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/world"
)

// BehaviorEngine manages beast AI behaviors and produces actions each tick.
type BehaviorEngine struct {
	cave           *world.Cave
	adjacencyGraph world.AdjacencyGraph
	roomTypeReg    *world.RoomTypeRegistry
	params         *BehaviorParams
	behaviors      map[int]Behavior // beastID -> assigned behavior
}

// NewBehaviorEngine creates a BehaviorEngine with the given cave and adjacency graph.
// If params is nil, DefaultBehaviorParams() is used.
func NewBehaviorEngine(cave *world.Cave, adjacencyGraph world.AdjacencyGraph, roomTypeReg *world.RoomTypeRegistry, params *BehaviorParams) *BehaviorEngine {
	if params == nil {
		params = DefaultBehaviorParams()
	}
	return &BehaviorEngine{
		cave:           cave,
		adjacencyGraph: adjacencyGraph,
		roomTypeReg:    roomTypeReg,
		params:         params,
		behaviors:      make(map[int]Behavior),
	}
}

// AssignBehavior assigns a behavior AI pattern to a beast.
// The behaviorType determines which Behavior implementation is created.
// For Patrol, the beast's current room and adjacency graph are used to build the route.
// For Chase, a default target of 0 and timeout of 10 is used; the engine
// will reassign with a real target when an invader is detected.
func (be *BehaviorEngine) AssignBehavior(beast *Beast, behaviorType BehaviorType) {
	var b Behavior
	switch behaviorType {
	case Guard:
		b = &GuardBehavior{}
	case Patrol:
		adj := be.adjacencyGraph.Neighbors(beast.RoomID)
		b = NewPatrolBehavior(beast.RoomID, adj, be.params.PatrolRestTicks)
	case Chase:
		b = NewChaseBehavior(0, be.params.ChaseTimeoutTicks)
	case Flee:
		roomTypeIDs := be.buildRoomTypeMap()
		b = NewFleeBehavior(be.params.FleeHPThreshold, roomTypeIDs)
	default:
		b = &GuardBehavior{}
	}
	be.behaviors[beast.ID] = b
}

// SetBehavior directly sets a specific behavior instance for a beast.
func (be *BehaviorEngine) SetBehavior(beastID int, behavior Behavior) {
	be.behaviors[beastID] = behavior
}

// GetBehavior returns the currently assigned behavior for a beast, or nil if none.
func (be *BehaviorEngine) GetBehavior(beastID int) Behavior {
	return be.behaviors[beastID]
}

// RemoveBehavior removes the behavior entry for a beast.
// Call this when a beast is defeated or removed from the cave to prevent
// stale entries from accumulating in the behaviors map.
func (be *BehaviorEngine) RemoveBehavior(beastID int) {
	delete(be.behaviors, beastID)
}

// BeastAction records the action decided for a beast and its context.
type BeastAction struct {
	// BeastID is the ID of the beast that took the action.
	BeastID int
	// Action is the decided action.
	Action Action
	// PreviousRoomID is the room the beast was in before the action.
	PreviousRoomID int
	// ResultingState is the beast state after the action is applied.
	ResultingState BeastState
}

// Tick runs one tick of behavior AI for all beasts.
// It builds BehaviorContext for each beast, calls DecideAction, checks HP
// for automatic Flee transitions, and resolves movement conflicts.
func (be *BehaviorEngine) Tick(beasts []*Beast, invaderPositions map[int][]int, roomChi map[int]*fengshui.RoomChi) []BeastAction {
	// Build room-beast map for context.
	roomBeasts := make(map[int][]int)
	for _, beast := range beasts {
		if beast.RoomID != 0 {
			roomBeasts[beast.RoomID] = append(roomBeasts[beast.RoomID], beast.ID)
		}
	}

	// Phase 1: Decide actions for each beast.
	type pendingAction struct {
		beast  *Beast
		action Action
	}
	pending := make([]pendingAction, 0, len(beasts))

	for _, beast := range beasts {
		if beast.RoomID == 0 {
			continue
		}

		// Stunned beasts cannot act or fight; skip them entirely.
		if beast.State == Stunned {
			continue
		}

		// HP threshold check: auto-transition to Flee if needed.
		be.checkFleeTransition(beast)

		b := be.behaviors[beast.ID]
		if b == nil {
			continue
		}

		// Check chase timeout: revert to Guard if timed out.
		if chase, ok := b.(*ChaseBehavior); ok && chase.TimedOut() {
			b = &GuardBehavior{}
			be.behaviors[beast.ID] = b
		}

		ctx := BehaviorContext{
			Beast:           beast,
			RoomID:          beast.RoomID,
			AdjacentRoomIDs: be.adjacencyGraph.Neighbors(beast.RoomID),
			RoomBeasts:      roomBeasts,
			InvaderRoomIDs:  invaderPositions,
			RoomChi:         roomChi[beast.RoomID],
		}

		action := b.DecideAction(ctx)

		// Patrol -> Chase transition: if patrol decided to move toward invader.
		if b.Type() == Patrol && action.Type == MoveToRoom {
			// Check if moving toward an invader room.
			if invaders, ok := invaderPositions[action.TargetRoomID]; ok && len(invaders) > 0 {
				chase := NewChaseBehavior(invaders[0], be.params.ChaseTimeoutTicks)
				be.behaviors[beast.ID] = chase
			}
		}

		pending = append(pending, pendingAction{beast: beast, action: action})
	}

	// Phase 2: Resolve movement conflicts (first-come by beast ID order).
	// Sort by beast ID for deterministic ordering.
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].beast.ID < pending[j].beast.ID
	})

	movedTo := make(map[int]bool) // room ID -> already claimed by a moving beast this tick
	actions := make([]BeastAction, 0, len(pending))

	for _, p := range pending {
		action := p.action
		resultState := p.beast.State

		switch action.Type {
		case MoveToRoom, Retreat:
			if movedTo[action.TargetRoomID] {
				// Conflict: another beast already moving to this room this tick.
				action = Action{Type: Stay}
				resultState = p.beast.State
			} else {
				movedTo[action.TargetRoomID] = true
				if action.Type == Retreat {
					resultState = Recovering
				} else {
					resultState = Patrolling
				}
			}
		case Attack:
			resultState = Fighting
		case Stay:
			// Determine resulting state based on behavior.
			b := be.behaviors[p.beast.ID]
			if b != nil {
				switch b.Type() {
				case Guard:
					resultState = Idle
				case Patrol:
					resultState = Patrolling
				case Flee:
					resultState = Recovering
				default:
					resultState = Idle
				}
			}
		}

		actions = append(actions, BeastAction{
			BeastID:        p.beast.ID,
			Action:         action,
			PreviousRoomID: p.beast.RoomID,
			ResultingState: resultState,
		})
	}

	return actions
}

// Behavior engine errors.
var (
	ErrBeastNotFound = errors.New("beast not found")
	ErrRoomNotFound  = errors.New("room not found")
)

// ApplyActions applies a list of BeastAction results, updating beast positions and states.
func ApplyActions(beasts []*Beast, rooms map[int]*world.Room, roomTypeReg *world.RoomTypeRegistry, actions []BeastAction) error {
	beastMap := make(map[int]*Beast, len(beasts))
	for _, b := range beasts {
		beastMap[b.ID] = b
	}

	for _, ba := range actions {
		beast, ok := beastMap[ba.BeastID]
		if !ok {
			return fmt.Errorf("%w: %d", ErrBeastNotFound, ba.BeastID)
		}

		switch ba.Action.Type {
		case MoveToRoom, Retreat:
			fromRoom, ok := rooms[ba.PreviousRoomID]
			if !ok {
				return fmt.Errorf("%w: from room %d", ErrRoomNotFound, ba.PreviousRoomID)
			}
			toRoom, ok := rooms[ba.Action.TargetRoomID]
			if !ok {
				return fmt.Errorf("%w: to room %d", ErrRoomNotFound, ba.Action.TargetRoomID)
			}
			toRoomType, err := roomTypeReg.Get(toRoom.TypeID)
			if err != nil {
				return fmt.Errorf("apply action: %w", err)
			}
			if err := MoveBeast(beast, fromRoom, toRoom, toRoomType); err != nil {
				return fmt.Errorf("apply action: %w", err)
			}
		}

		beast.State = ba.ResultingState
	}

	return nil
}

// checkFleeTransition checks if a beast should transition to Flee behavior
// based on its HP threshold.
func (be *BehaviorEngine) checkFleeTransition(beast *Beast) {
	b := be.behaviors[beast.ID]
	if b == nil {
		return
	}
	// Don't re-assign if already fleeing.
	if b.Type() == Flee {
		return
	}
	if ShouldFlee(beast, be.params.FleeHPThreshold) {
		roomTypeIDs := be.buildRoomTypeMap()
		be.behaviors[beast.ID] = NewFleeBehavior(be.params.FleeHPThreshold, roomTypeIDs)
	}
}

// buildRoomTypeMap creates a mapping from room ID to room type ID for all rooms in the cave.
func (be *BehaviorEngine) buildRoomTypeMap() map[int]string {
	m := make(map[int]string)
	for _, r := range be.cave.Rooms {
		m[r.ID] = r.TypeID
	}
	return m
}
