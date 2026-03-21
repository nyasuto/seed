package world

// AdjacencyGraph represents rooms as nodes and corridors as edges,
// providing graph operations on the cave topology.
type AdjacencyGraph struct {
	// edges maps each room ID to its set of neighboring room IDs.
	edges map[int]map[int]bool
}

// BuildAdjacencyGraph constructs an AdjacencyGraph from the cave's corridors.
// Each corridor creates a bidirectional edge between its two rooms.
func (c *Cave) BuildAdjacencyGraph() AdjacencyGraph {
	g := AdjacencyGraph{
		edges: make(map[int]map[int]bool),
	}

	// Register all rooms as nodes (even if they have no corridors).
	for _, r := range c.Rooms {
		if g.edges[r.ID] == nil {
			g.edges[r.ID] = make(map[int]bool)
		}
	}

	// Add bidirectional edges for each corridor.
	for _, cor := range c.Corridors {
		if g.edges[cor.FromRoomID] == nil {
			g.edges[cor.FromRoomID] = make(map[int]bool)
		}
		if g.edges[cor.ToRoomID] == nil {
			g.edges[cor.ToRoomID] = make(map[int]bool)
		}
		g.edges[cor.FromRoomID][cor.ToRoomID] = true
		g.edges[cor.ToRoomID][cor.FromRoomID] = true
	}

	return g
}

// Neighbors returns the IDs of rooms directly connected to the given room.
// Returns nil if the room ID is not in the graph.
func (g *AdjacencyGraph) Neighbors(roomID int) []int {
	neighbors, ok := g.edges[roomID]
	if !ok {
		return nil
	}
	result := make([]int, 0, len(neighbors))
	for id := range neighbors {
		result = append(result, id)
	}
	return result
}

// PathExists reports whether there is a path between the two rooms
// using breadth-first search.
func (g *AdjacencyGraph) PathExists(from, to int) bool {
	if _, ok := g.edges[from]; !ok {
		return false
	}
	if _, ok := g.edges[to]; !ok {
		return false
	}
	if from == to {
		return true
	}

	visited := make(map[int]bool)
	queue := []int{from}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for neighbor := range g.edges[current] {
			if neighbor == to {
				return true
			}
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return false
}
