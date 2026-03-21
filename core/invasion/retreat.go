package invasion

import (
	"slices"
	"sort"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// RetreatReason represents why an invader decided to retreat.
type RetreatReason int

const (
	// ReasonLowHP means the invader's HP fell below its retreat threshold.
	ReasonLowHP RetreatReason = iota
	// ReasonMoraleBroken means half or more of the wave companions are defeated.
	ReasonMoraleBroken
	// ReasonGoalComplete means the invader achieved its objective.
	ReasonGoalComplete
)

// String returns the name of the retreat reason.
func (r RetreatReason) String() string {
	switch r {
	case ReasonLowHP:
		return "LowHP"
	case ReasonMoraleBroken:
		return "MoraleBroken"
	case ReasonGoalComplete:
		return "GoalComplete"
	default:
		return "Unknown"
	}
}

// RetreatEvaluator determines whether an invader should retreat from the cave.
type RetreatEvaluator struct {
	classRegistry *InvaderClassRegistry
}

// NewRetreatEvaluator creates a new RetreatEvaluator with the given class registry.
func NewRetreatEvaluator(registry *InvaderClassRegistry) *RetreatEvaluator {
	return &RetreatEvaluator{classRegistry: registry}
}

// ShouldRetreat evaluates whether the given invader should retreat.
// It checks three conditions in order:
//  1. Goal achieved → retreat to carry spoils home
//  2. HP ≤ MaxHP × RetreatThreshold → retreat due to low health
//  3. Half or more of wave companions defeated → retreat due to morale break
//
// The waveInvaders parameter includes all invaders in the same wave (including the subject).
// Returns true and the reason if the invader should retreat.
func (re *RetreatEvaluator) ShouldRetreat(invader *Invader, waveInvaders []*Invader) (bool, RetreatReason) {
	// Already retreating or defeated — no re-evaluation needed.
	if invader.State == Retreating || invader.State == Defeated {
		return false, 0
	}

	// 3. Goal achieved → retreat with spoils.
	if invader.State == GoalAchieved {
		return true, ReasonGoalComplete
	}

	// 1. HP ≤ MaxHP × RetreatThreshold → retreat.
	class, err := re.classRegistry.Get(invader.ClassID)
	if err == nil && invader.MaxHP > 0 {
		threshold := float64(invader.MaxHP) * class.RetreatThreshold
		if float64(invader.HP) <= threshold {
			return true, ReasonLowHP
		}
	}

	// 2. Half or more of wave companions defeated → morale break.
	if len(waveInvaders) > 1 {
		defeatedCount := 0
		for _, inv := range waveInvaders {
			if inv.State == Defeated {
				defeatedCount++
			}
		}
		if defeatedCount*2 >= len(waveInvaders) {
			return true, ReasonMoraleBroken
		}
	}

	return false, 0
}

// RetreatResult holds the outcome of a completed retreat.
type RetreatResult struct {
	// InvaderID is the retreating invader's ID.
	InvaderID int
	// Reason is why the invader retreated.
	Reason RetreatReason
	// StolenChi is the amount of chi stolen during retreat.
	// Only non-zero when a thief retreats with GoalComplete.
	StolenChi float64
}

// RetreatPathfinder computes retreat paths for invaders heading back to the entrance.
// It uses the invader's ExplorationMemory to retrace visited rooms in reverse
// chronological order, falling back to BFS shortest path if needed.
type RetreatPathfinder struct {
	cave           *world.Cave
	adjacencyGraph world.AdjacencyGraph
}

// NewRetreatPathfinder creates a new RetreatPathfinder.
func NewRetreatPathfinder(cave *world.Cave, graph world.AdjacencyGraph) *RetreatPathfinder {
	return &RetreatPathfinder{
		cave:           cave,
		adjacencyGraph: graph,
	}
}

// FindRetreatPath returns the sequence of room IDs the invader should follow
// to retreat from their current position to the entry room.
// The path starts with the current room and ends with the entry room.
//
// Strategy: retrace visited rooms in reverse visit-tick order, filtering to
// rooms that are reachable (adjacent in the graph) along the way.
// If the memory-based path cannot reach the entry room, falls back to
// BFS shortest path.
func (rp *RetreatPathfinder) FindRetreatPath(invader *Invader) []int {
	currentRoom := invader.CurrentRoomID
	entryRoom := rp.findEntryRoom(invader.Memory)

	if currentRoom == entryRoom {
		return []int{currentRoom}
	}

	// Build reverse-chronological list of visited rooms (excluding current).
	path := rp.buildMemoryPath(currentRoom, entryRoom, invader.Memory)
	if path != nil {
		return path
	}

	// Fallback: BFS shortest path.
	return rp.bfsPath(currentRoom, entryRoom)
}

// findEntryRoom returns the room with the earliest visit tick in the invader's memory.
// This is the room where the invader entered the cave.
func (rp *RetreatPathfinder) findEntryRoom(memory *ExplorationMemory) int {
	var entryRoom int
	var earliest types.Tick
	first := true
	for roomID, tick := range memory.VisitedRooms {
		if first || tick < earliest {
			earliest = tick
			entryRoom = roomID
			first = false
		}
	}
	return entryRoom
}

// buildMemoryPath tries to construct a retreat path by retracing visited rooms
// in reverse visit-tick order. Only includes rooms connected by the adjacency graph.
// Returns nil if the path cannot reach the entry room.
func (rp *RetreatPathfinder) buildMemoryPath(from, to int, memory *ExplorationMemory) []int {
	// Collect visited rooms sorted by visit tick descending (most recent first),
	// excluding the current room.
	type visitEntry struct {
		roomID int
		tick   types.Tick
	}
	var entries []visitEntry
	for roomID, tick := range memory.VisitedRooms {
		if roomID != from {
			entries = append(entries, visitEntry{roomID, tick})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].tick != entries[j].tick {
			return entries[i].tick > entries[j].tick
		}
		return entries[i].roomID < entries[j].roomID // deterministic tie-break
	})

	path := []int{from}
	current := from
	used := map[int]bool{from: true}

	for _, entry := range entries {
		if !rp.isAdjacent(current, entry.roomID) {
			continue
		}
		path = append(path, entry.roomID)
		used[entry.roomID] = true
		current = entry.roomID
		if current == to {
			return path
		}
	}

	// Memory-based path didn't reach the entry room.
	return nil
}

// isAdjacent checks whether two rooms are directly connected in the adjacency graph.
func (rp *RetreatPathfinder) isAdjacent(a, b int) bool {
	return slices.Contains(rp.adjacencyGraph.Neighbors(a), b)
}

// bfsPath finds the shortest path from 'from' to 'to' using BFS.
func (rp *RetreatPathfinder) bfsPath(from, to int) []int {
	if from == to {
		return []int{from}
	}
	visited := map[int]bool{from: true}
	parent := map[int]int{}
	queue := []int{from}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		neighbors := rp.adjacencyGraph.Neighbors(current)
		sort.Ints(neighbors)
		for _, next := range neighbors {
			if visited[next] {
				continue
			}
			visited[next] = true
			parent[next] = current
			if next == to {
				return reconstructRetreatPath(parent, from, to)
			}
			queue = append(queue, next)
		}
	}
	return nil
}

// reconstructRetreatPath traces back from 'to' to 'from' using the parent map.
func reconstructRetreatPath(parent map[int]int, from, to int) []int {
	path := []int{to}
	current := to
	for current != from {
		current = parent[current]
		path = append(path, current)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
