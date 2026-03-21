package invasion

import (
	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// InvasionEventType classifies the kind of event that occurred during an invasion tick.
type InvasionEventType int

const (
	// WaveStarted means a pending wave was activated.
	WaveStarted InvasionEventType = iota
	// WaveCompleted means all invaders in a wave were defeated (defense success).
	WaveCompleted
	// WaveFailed means invaders escaped or achieved goals (defense failure).
	WaveFailed
	// InvaderMoved means an invader moved to a new room.
	InvaderMoved
	// InvaderDefeated means an invader's HP reached zero.
	InvaderDefeated
	// InvaderRetreating means an invader began or continued retreating.
	InvaderRetreating
	// InvaderEscaped means a retreating invader reached the entrance and left.
	InvaderEscaped
	// CombatOccurred means a combat round was resolved between a beast and an invader.
	CombatOccurred
	// BeastDefeated means a beast's HP reached zero.
	BeastDefeated
	// TrapTriggered means a trap activated against an invader.
	TrapTriggered
	// GoalAchievedEvent means an invader completed its objective.
	GoalAchievedEvent
)

// String returns the name of the event type.
func (t InvasionEventType) String() string {
	switch t {
	case WaveStarted:
		return "WaveStarted"
	case WaveCompleted:
		return "WaveCompleted"
	case WaveFailed:
		return "WaveFailed"
	case InvaderMoved:
		return "InvaderMoved"
	case InvaderDefeated:
		return "InvaderDefeated"
	case InvaderRetreating:
		return "InvaderRetreating"
	case InvaderEscaped:
		return "InvaderEscaped"
	case CombatOccurred:
		return "CombatOccurred"
	case BeastDefeated:
		return "BeastDefeated"
	case TrapTriggered:
		return "TrapTriggered"
	case GoalAchievedEvent:
		return "GoalAchievedEvent"
	default:
		return "Unknown"
	}
}

// InvasionEvent records a single event that occurred during an invasion tick.
type InvasionEvent struct {
	// Type is the classification of this event.
	Type InvasionEventType
	// Tick is the game tick when this event occurred.
	Tick types.Tick
	// WaveID is the wave this event is associated with (0 if not applicable).
	WaveID int
	// InvaderID is the invader involved (0 if not applicable).
	InvaderID int
	// BeastID is the beast involved (0 if not applicable).
	BeastID int
	// RoomID is the room where the event occurred (0 if not applicable).
	RoomID int
	// Damage is the amount of damage dealt (0 if not applicable).
	Damage int
	// RewardChi is the chi reward for defeating an invader.
	RewardChi float64
	// StolenChi is the chi stolen by an escaping invader.
	StolenChi float64
	// Details provides additional human-readable information.
	Details string
}

// InvasionEngine orchestrates the invasion tick loop, combining combat,
// pathfinding, retreat evaluation, and trap effects into a cohesive system.
type InvasionEngine struct {
	combatEngine      *CombatEngine
	pathfinder        *Pathfinder
	retreatEvaluator  *RetreatEvaluator
	retreatPathfinder *RetreatPathfinder
	trapEffects       []TrapEffect
	cave              *world.Cave
	adjacencyGraph    world.AdjacencyGraph
	rng               types.RNG
	classRegistry     *InvaderClassRegistry
}

// NewInvasionEngine creates a new InvasionEngine with all required subsystems.
func NewInvasionEngine(
	cave *world.Cave,
	adjacencyGraph world.AdjacencyGraph,
	combatParams CombatParams,
	rng types.RNG,
	classRegistry *InvaderClassRegistry,
	trapEffects []TrapEffect,
) *InvasionEngine {
	return &InvasionEngine{
		combatEngine:      NewCombatEngine(combatParams, rng),
		pathfinder:        NewPathfinder(cave, adjacencyGraph),
		retreatEvaluator:  NewRetreatEvaluator(classRegistry),
		retreatPathfinder: NewRetreatPathfinder(cave, adjacencyGraph),
		trapEffects:       trapEffects,
		cave:              cave,
		adjacencyGraph:    adjacencyGraph,
		rng:               rng,
		classRegistry:     classRegistry,
	}
}

// Tick processes one tick of the invasion system. It returns all events
// that occurred during this tick.
//
// Processing order:
//  1. Activate pending waves whose TriggerTick has been reached
//  2. Update exploration memory for each invader
//  3. Check retreat conditions
//  4. Move advancing invaders toward their goals
//  5. Move retreating invaders toward the entrance
//  6. Apply trap effects to advancing invaders in trap rooms
//  7. Resolve combat in rooms with both beasts and invaders
//  8. Check for defeated invaders
//  9. Check for defeated beasts
//  10. Check goal achievement
//  11. Evaluate wave completion
func (e *InvasionEngine) Tick(
	currentTick types.Tick,
	waves []*InvasionWave,
	beasts []*senju.Beast,
	rooms []*world.Room,
	roomTypes *world.RoomTypeRegistry,
	roomChi map[int]*fengshui.RoomChi,
) []InvasionEvent {
	var events []InvasionEvent

	// 1. Activate pending waves.
	for _, wave := range waves {
		if wave.State == Pending && currentTick >= wave.TriggerTick {
			wave.State = Active
			events = append(events, InvasionEvent{
				Type:   WaveStarted,
				Tick:   currentTick,
				WaveID: wave.ID,
			})
		}
	}

	// Process only active waves.
	for _, wave := range waves {
		if wave.State != Active {
			continue
		}

		for _, inv := range wave.Invaders {
			// Skip finished invaders.
			if inv.State == Defeated {
				continue
			}

			// 2. Update exploration memory.
			inv.Memory.Visit(inv.CurrentRoomID, currentTick, e.cave, rooms)

			// 3. Check retreat conditions.
			if inv.State == Advancing || inv.State == Fighting {
				// Check goal achievement first.
				if inv.Goal.IsAchieved(e.cave, inv) && inv.State != GoalAchieved {
					inv.State = GoalAchieved
					events = append(events, InvasionEvent{
						Type:      GoalAchievedEvent,
						Tick:      currentTick,
						WaveID:    wave.ID,
						InvaderID: inv.ID,
						RoomID:    inv.CurrentRoomID,
					})
				}

				shouldRetreat, reason := e.retreatEvaluator.ShouldRetreat(inv, wave.Invaders)
				if shouldRetreat {
					inv.State = Retreating
					stolenChi := e.calcStolenChi(inv, reason, roomChi)
					events = append(events, InvasionEvent{
						Type:      InvaderRetreating,
						Tick:      currentTick,
						WaveID:    wave.ID,
						InvaderID: inv.ID,
						RoomID:    inv.CurrentRoomID,
						StolenChi: stolenChi,
						Details:   reason.String(),
					})
					continue
				}
			}

			// 6. Handle SlowTicks — slowed invaders skip movement.
			if inv.SlowTicks > 0 {
				inv.SlowTicks--
				continue
			}

			// 4. Move advancing invaders.
			if inv.State == Advancing {
				nextRoom := e.pathfinder.FindNextRoom(inv, e.rng)
				if nextRoom != inv.CurrentRoomID {
					inv.CurrentRoomID = nextRoom
					inv.StayTicks = 0
					events = append(events, InvasionEvent{
						Type:      InvaderMoved,
						Tick:      currentTick,
						WaveID:    wave.ID,
						InvaderID: inv.ID,
						RoomID:    nextRoom,
					})
				} else {
					inv.StayTicks++
				}
			}

			// 5. Move retreating invaders.
			if inv.State == Retreating {
				retreatPath := e.retreatPathfinder.FindRetreatPath(inv)
				if len(retreatPath) >= 2 {
					inv.CurrentRoomID = retreatPath[1]
					inv.StayTicks = 0
					events = append(events, InvasionEvent{
						Type:      InvaderRetreating,
						Tick:      currentTick,
						WaveID:    wave.ID,
						InvaderID: inv.ID,
						RoomID:    retreatPath[1],
					})
				}
				// Check if at entry room (earliest visited).
				if e.isAtEntryRoom(inv) {
					stolenChi := e.calcEscapeStolenChi(inv, roomChi)
					events = append(events, InvasionEvent{
						Type:      InvaderEscaped,
						Tick:      currentTick,
						WaveID:    wave.ID,
						InvaderID: inv.ID,
						RoomID:    inv.CurrentRoomID,
						StolenChi: stolenChi,
					})
					inv.State = Defeated // Remove from active pool.
				}
			}
		}
	}

	// 7. Apply trap effects to advancing invaders in trap rooms.
	trapMap := e.buildTrapMap()
	for _, wave := range waves {
		if wave.State != Active {
			continue
		}
		for _, inv := range wave.Invaders {
			if inv.State != Advancing {
				continue
			}
			trap, hasTrap := trapMap[inv.CurrentRoomID]
			if !hasTrap {
				continue
			}
			result := ApplyTrap(inv, trap, e.combatEngine.params)
			events = append(events, InvasionEvent{
				Type:      TrapTriggered,
				Tick:      currentTick,
				WaveID:    wave.ID,
				InvaderID: inv.ID,
				RoomID:    inv.CurrentRoomID,
				Damage:    result.Damage,
			})
		}
	}

	// 8. Combat: match beasts and invaders in the same room.
	beastsByRoom := e.buildBeastsByRoom(beasts)
	invadersByRoom := e.buildInvadersByRoom(waves)

	for roomID, roomBeasts := range beastsByRoom {
		roomInvaders, ok := invadersByRoom[roomID]
		if !ok || len(roomInvaders) == 0 {
			continue
		}

		// Filter to alive beasts and advancing/fighting invaders.
		var aliveBeasts []*senju.Beast
		for _, b := range roomBeasts {
			if b.HP > 0 {
				aliveBeasts = append(aliveBeasts, b)
			}
		}
		var fightingInvaders []*Invader
		for _, inv := range roomInvaders {
			if inv.State == Advancing || inv.State == Fighting {
				fightingInvaders = append(fightingInvaders, inv)
			}
		}
		if len(aliveBeasts) == 0 || len(fightingInvaders) == 0 {
			continue
		}

		// Mark invaders as fighting.
		for _, inv := range fightingInvaders {
			inv.State = Fighting
		}

		chi := roomChi[roomID]
		results := e.combatEngine.ResolveRoomCombat(aliveBeasts, fightingInvaders, chi)
		for i, r := range results {
			var beastID, invaderID int
			if i < len(aliveBeasts) {
				beastID = aliveBeasts[i].ID
			}
			if i < len(fightingInvaders) {
				invaderID = fightingInvaders[i].ID
			}
			events = append(events, InvasionEvent{
				Type:      CombatOccurred,
				Tick:      currentTick,
				BeastID:   beastID,
				InvaderID: invaderID,
				RoomID:    roomID,
				Damage:    r.BeastDamageTaken + r.InvaderDamageTaken,
				Details:   r.FirstAttacker + " attacked first",
			})
		}
	}

	// 9. Check for defeated invaders.
	for _, wave := range waves {
		if wave.State != Active {
			continue
		}
		for _, inv := range wave.Invaders {
			if inv.HP <= 0 && inv.State != Defeated {
				inv.State = Defeated
				class, err := e.classRegistry.Get(inv.ClassID)
				rewardChi := 0.0
				if err == nil {
					rewardChi = class.RewardChi
				}
				events = append(events, InvasionEvent{
					Type:      InvaderDefeated,
					Tick:      currentTick,
					WaveID:    wave.ID,
					InvaderID: inv.ID,
					RoomID:    inv.CurrentRoomID,
					RewardChi: rewardChi,
				})
			}
		}
	}

	// 10. Check for defeated beasts.
	for _, b := range beasts {
		if b.HP <= 0 && b.State != senju.Recovering {
			b.State = senju.Recovering
			events = append(events, InvasionEvent{
				Type:    BeastDefeated,
				Tick:    currentTick,
				BeastID: b.ID,
				RoomID:  b.RoomID,
			})
		}
	}

	// 11. Evaluate wave completion.
	for _, wave := range waves {
		if wave.State != Active {
			continue
		}
		allDone := true
		anyGoalAchieved := false
		anyEscaped := false
		allDefeated := true
		for _, inv := range wave.Invaders {
			if inv.State != Defeated {
				allDefeated = false
			}
			if inv.State == Advancing || inv.State == Fighting {
				allDone = false
			}
			if inv.State == GoalAchieved {
				anyGoalAchieved = true
				allDone = false // still needs to retreat
			}
			if inv.State == Retreating {
				allDone = false
			}
		}

		if !allDone {
			continue
		}

		if allDefeated {
			wave.State = Completed
			events = append(events, InvasionEvent{
				Type:   WaveCompleted,
				Tick:   currentTick,
				WaveID: wave.ID,
			})
		} else if anyGoalAchieved || anyEscaped {
			wave.State = Failed
			events = append(events, InvasionEvent{
				Type:   WaveFailed,
				Tick:   currentTick,
				WaveID: wave.ID,
			})
		} else {
			// All remaining invaders are Defeated — this is a defense success.
			wave.State = Completed
			events = append(events, InvasionEvent{
				Type:   WaveCompleted,
				Tick:   currentTick,
				WaveID: wave.ID,
			})
		}
	}

	return events
}

// BuildInvaderPositions returns a map of room ID to invader IDs for all
// active Advancing or Fighting invaders. This is used by senju.BehaviorEngine
// to determine beast responses.
func (e *InvasionEngine) BuildInvaderPositions(waves []*InvasionWave) map[int][]int {
	positions := make(map[int][]int)
	for _, wave := range waves {
		if wave.State != Active {
			continue
		}
		for _, inv := range wave.Invaders {
			if inv.State == Advancing || inv.State == Fighting {
				positions[inv.CurrentRoomID] = append(positions[inv.CurrentRoomID], inv.ID)
			}
		}
	}
	return positions
}

// CollectRewards sums up the RewardChi from all InvaderDefeated events.
func (e *InvasionEngine) CollectRewards(events []InvasionEvent) float64 {
	total := 0.0
	for _, ev := range events {
		if ev.Type == InvaderDefeated {
			total += ev.RewardChi
		}
	}
	return total
}

// CollectStolenChi sums up the StolenChi from all InvaderEscaped events.
func (e *InvasionEngine) CollectStolenChi(events []InvasionEvent) float64 {
	total := 0.0
	for _, ev := range events {
		if ev.Type == InvaderEscaped {
			total += ev.StolenChi
		}
	}
	return total
}

// buildTrapMap creates a lookup from room ID to TrapEffect.
func (e *InvasionEngine) buildTrapMap() map[int]TrapEffect {
	m := make(map[int]TrapEffect, len(e.trapEffects))
	for _, t := range e.trapEffects {
		m[t.RoomID] = t
	}
	return m
}

// buildBeastsByRoom groups living beasts by their current room ID.
func (e *InvasionEngine) buildBeastsByRoom(beasts []*senju.Beast) map[int][]*senju.Beast {
	m := make(map[int][]*senju.Beast)
	for _, b := range beasts {
		if b.HP > 0 {
			m[b.RoomID] = append(m[b.RoomID], b)
		}
	}
	return m
}

// buildInvadersByRoom groups active invaders by their current room ID.
func (e *InvasionEngine) buildInvadersByRoom(waves []*InvasionWave) map[int][]*Invader {
	m := make(map[int][]*Invader)
	for _, wave := range waves {
		if wave.State != Active {
			continue
		}
		for _, inv := range wave.Invaders {
			if inv.State == Advancing || inv.State == Fighting {
				m[inv.CurrentRoomID] = append(m[inv.CurrentRoomID], inv)
			}
		}
	}
	return m
}

// isAtEntryRoom checks if the invader is at its entry room (the earliest visited room).
func (e *InvasionEngine) isAtEntryRoom(inv *Invader) bool {
	if inv.Memory == nil || len(inv.Memory.VisitedRooms) == 0 {
		return true
	}
	var entryRoom int
	var earliest types.Tick
	first := true
	for roomID, tick := range inv.Memory.VisitedRooms {
		if first || tick < earliest {
			earliest = tick
			entryRoom = roomID
			first = false
		}
	}
	return inv.CurrentRoomID == entryRoom
}

// calcStolenChi computes the chi stolen when an invader begins retreating.
// Only StealTreasure invaders with GoalComplete steal chi.
func (e *InvasionEngine) calcStolenChi(inv *Invader, reason RetreatReason, roomChi map[int]*fengshui.RoomChi) float64 {
	if reason != ReasonGoalComplete || inv.Goal.Type() != StealTreasure {
		return 0
	}
	chi, ok := roomChi[inv.CurrentRoomID]
	if !ok {
		return 0
	}
	// Steal half the current chi in the room.
	stolen := chi.Current * 0.5
	chi.Current -= stolen
	if chi.Current < 0 {
		chi.Current = 0
	}
	return stolen
}

// calcEscapeStolenChi computes the chi carried away when an invader escapes.
// This is for invaders that already started retreating with stolen goods.
func (e *InvasionEngine) calcEscapeStolenChi(inv *Invader, roomChi map[int]*fengshui.RoomChi) float64 {
	if inv.Goal.Type() != StealTreasure {
		return 0
	}
	// The stolen chi was already calculated at retreat start; record a nominal amount.
	return 0
}
