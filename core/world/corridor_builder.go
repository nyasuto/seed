package world

import (
	"errors"
	"fmt"

	"github.com/nyasuto/seed/core/types"
)

// ErrNoPath is returned when no valid path can be found between two positions.
var ErrNoPath = errors.New("no path found between positions")

// BuildCorridor finds the shortest path between fromPos and toPos using BFS,
// carving through rock as needed. Existing corridors and room floors are
// traversable, but cells belonging to rooms other than the source/destination
// are avoided. The resulting path cells are set to Corridor type on the grid.
func BuildCorridor(grid *Grid, fromPos, toPos types.Pos, corridorID, fromRoomID, toRoomID int) (Corridor, error) {
	if !grid.InBounds(fromPos) {
		return Corridor{}, fmt.Errorf("from position (%d,%d): %w", fromPos.X, fromPos.Y, ErrOutOfBounds)
	}
	if !grid.InBounds(toPos) {
		return Corridor{}, fmt.Errorf("to position (%d,%d): %w", toPos.X, toPos.Y, ErrOutOfBounds)
	}

	// BFS
	type node struct {
		pos types.Pos
	}

	visited := make(map[types.Pos]bool)
	parent := make(map[types.Pos]types.Pos)
	queue := []node{{pos: fromPos}}
	visited[fromPos] = true

	found := false
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.pos == toPos {
			found = true
			break
		}

		for _, neighbor := range current.pos.Neighbors() {
			if !grid.InBounds(neighbor) || visited[neighbor] {
				continue
			}
			if !canTraverse(grid, neighbor, fromRoomID, toRoomID) {
				continue
			}
			visited[neighbor] = true
			parent[neighbor] = current.pos
			queue = append(queue, node{pos: neighbor})
		}
	}

	if !found {
		return Corridor{}, fmt.Errorf("from (%d,%d) to (%d,%d): %w", fromPos.X, fromPos.Y, toPos.X, toPos.Y, ErrNoPath)
	}

	// Reconstruct path from toPos back to fromPos
	path := []types.Pos{}
	for pos := toPos; pos != fromPos; pos = parent[pos] {
		path = append(path, pos)
	}
	path = append(path, fromPos)

	// Reverse to get fromPos → toPos order
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	// Carve corridor cells on the grid (skip cells that are already room floor or entrance)
	for _, pos := range path {
		cell, _ := grid.At(pos)
		if cell.Type == Rock {
			_ = grid.Set(pos, Cell{Type: CorridorFloor, RoomID: 0})
		}
	}

	return Corridor{
		ID:         corridorID,
		FromRoomID: fromRoomID,
		ToRoomID:   toRoomID,
		Path:       path,
	}, nil
}

// canTraverse reports whether the given position can be included in a corridor path.
// Rock, Corridor, and Entrance cells are always traversable.
// RoomFloor cells are only traversable if they belong to the source or destination room.
// HardRock and Water cells are impassable and cannot be traversed.
func canTraverse(grid *Grid, pos types.Pos, fromRoomID, toRoomID int) bool {
	cell, err := grid.At(pos)
	if err != nil {
		return false
	}
	if cell.Type.IsImpassable() {
		return false
	}
	switch cell.Type {
	case Rock, CorridorFloor, Entrance:
		return true
	case RoomFloor:
		return cell.RoomID == fromRoomID || cell.RoomID == toRoomID
	default:
		return false
	}
}
