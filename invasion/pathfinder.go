package invasion

import (
	"sort"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// Pathfinder provides goal-oriented pathfinding for invaders navigating a cave.
// It uses the cave's adjacency graph and the invader's exploration memory
// to determine movement decisions.
type Pathfinder struct {
	cave           *world.Cave
	adjacencyGraph world.AdjacencyGraph
}

// NewPathfinder creates a new Pathfinder for the given cave.
func NewPathfinder(cave *world.Cave, graph world.AdjacencyGraph) *Pathfinder {
	return &Pathfinder{
		cave:           cave,
		adjacencyGraph: graph,
	}
}

// FindPath returns the shortest path (as a list of room IDs) from 'from' to 'to'
// using BFS on the adjacency graph. The result includes both endpoints.
// Returns nil if no path exists.
func (p *Pathfinder) FindPath(from, to int) []int {
	if from == to {
		return []int{from}
	}

	// BFS with parent tracking.
	visited := map[int]bool{from: true}
	parent := map[int]int{}
	queue := []int{from}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		neighbors := p.adjacencyGraph.Neighbors(current)
		sort.Ints(neighbors) // deterministic order
		for _, next := range neighbors {
			if visited[next] {
				continue
			}
			visited[next] = true
			parent[next] = current
			if next == to {
				// Reconstruct path.
				return reconstructPath(parent, from, to)
			}
			queue = append(queue, next)
		}
	}

	return nil // no path found
}

// reconstructPath traces back from 'to' to 'from' using the parent map.
func reconstructPath(parent map[int]int, from, to int) []int {
	path := []int{to}
	current := to
	for current != from {
		current = parent[current]
		path = append(path, current)
	}
	// Reverse.
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// FindNextRoom determines the next room an invader should move to.
// Decision logic:
//  1. If the goal's target room is known, follow the shortest path toward it.
//  2. If the target is unknown, prefer unvisited adjacent rooms (exploration).
//  3. If all adjacent rooms are visited, backtrack toward the nearest unvisited room.
//  4. If fully explored, move randomly (via RNG).
func (p *Pathfinder) FindNextRoom(invader *Invader, rng types.RNG) int {
	currentRoom := invader.CurrentRoomID
	memory := invader.Memory

	// 1. If goal target is known, move along shortest path.
	targetRoom := invader.Goal.TargetRoomID(p.cave, invader, memory)
	if targetRoom != 0 && targetRoom != currentRoom {
		path := p.FindPath(currentRoom, targetRoom)
		if path != nil && len(path) >= 2 {
			return path[1]
		}
	}

	neighbors := p.adjacencyGraph.Neighbors(currentRoom)
	if len(neighbors) == 0 {
		return currentRoom // isolated room, stay
	}
	sort.Ints(neighbors) // deterministic order

	// 2. Prefer unvisited adjacent rooms.
	var unvisited []int
	for _, n := range neighbors {
		if !memory.HasVisited(n) {
			unvisited = append(unvisited, n)
		}
	}
	if len(unvisited) > 0 {
		return unvisited[rng.Intn(len(unvisited))]
	}

	// 3. Backtrack toward the nearest unvisited room (BFS from current).
	nextStep := p.findStepTowardUnvisited(currentRoom, memory)
	if nextStep != 0 {
		return nextStep
	}

	// 4. Fully explored — random move.
	return neighbors[rng.Intn(len(neighbors))]
}

// findStepTowardUnvisited does a BFS from currentRoom, looking for any unvisited
// room. Returns the first step (adjacent room) on the path toward it,
// or 0 if all rooms are visited.
func (p *Pathfinder) findStepTowardUnvisited(currentRoom int, memory *ExplorationMemory) int {
	type bfsEntry struct {
		roomID    int
		firstStep int // the first room in the path from currentRoom
	}

	visited := map[int]bool{currentRoom: true}
	var queue []bfsEntry

	neighbors := p.adjacencyGraph.Neighbors(currentRoom)
	sort.Ints(neighbors)
	for _, n := range neighbors {
		visited[n] = true
		queue = append(queue, bfsEntry{roomID: n, firstStep: n})
	}

	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]

		if !memory.HasVisited(entry.roomID) {
			return entry.firstStep
		}

		nextNeighbors := p.adjacencyGraph.Neighbors(entry.roomID)
		sort.Ints(nextNeighbors)
		for _, next := range nextNeighbors {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, bfsEntry{roomID: next, firstStep: entry.firstStep})
			}
		}
	}

	return 0 // all rooms visited
}
