package fengshui

import (
	"errors"
	"fmt"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// ErrUnreachable is returned when no passable path can be found from the source position.
var ErrUnreachable = errors.New("no reachable path from source")

// BuildDragonVein creates a new DragonVein by computing its path from sourcePos
// using BFS. The path follows corridors, room floors, and entrances — rock cells
// are impassable. The resulting path covers all reachable non-rock cells from
// the source position.
func BuildDragonVein(cave *world.Cave, sourcePos types.Pos, element types.Element, flowRate float64) (*DragonVein, error) {
	if !cave.Grid.InBounds(sourcePos) {
		return nil, fmt.Errorf("source position (%d,%d): out of bounds", sourcePos.X, sourcePos.Y)
	}

	cell, err := cave.Grid.At(sourcePos)
	if err != nil {
		return nil, fmt.Errorf("reading source cell: %w", err)
	}
	if cell.Type == world.Rock {
		return nil, fmt.Errorf("source position (%d,%d): %w", sourcePos.X, sourcePos.Y, ErrUnreachable)
	}

	path := bfsPath(cave, sourcePos)
	if len(path) == 0 {
		return nil, fmt.Errorf("source position (%d,%d): %w", sourcePos.X, sourcePos.Y, ErrUnreachable)
	}

	return &DragonVein{
		SourcePos: sourcePos,
		Element:   element,
		FlowRate:  flowRate,
		Path:      path,
	}, nil
}

// RebuildDragonVein recomputes the path of an existing dragon vein based on
// the current state of the cave. The source position, element, and flow rate
// are preserved from the original vein.
func RebuildDragonVein(cave *world.Cave, existingVein *DragonVein) (*DragonVein, error) {
	rebuilt, err := BuildDragonVein(cave, existingVein.SourcePos, existingVein.Element, existingVein.FlowRate)
	if err != nil {
		return nil, err
	}
	rebuilt.ID = existingVein.ID
	return rebuilt, nil
}

// bfsPath performs a BFS from sourcePos and returns all reachable positions
// in BFS visit order. Only non-rock cells (corridor, room floor, entrance)
// are traversable.
func bfsPath(cave *world.Cave, sourcePos types.Pos) []types.Pos {
	visited := make(map[types.Pos]bool)
	queue := []types.Pos{sourcePos}
	visited[sourcePos] = true

	var path []types.Pos

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		path = append(path, current)

		for _, neighbor := range current.Neighbors() {
			if !cave.Grid.InBounds(neighbor) || visited[neighbor] {
				continue
			}
			cell, err := cave.Grid.At(neighbor)
			if err != nil || cell.Type == world.Rock {
				continue
			}
			visited[neighbor] = true
			queue = append(queue, neighbor)
		}
	}

	return path
}
