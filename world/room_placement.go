package world

import (
	"errors"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

// ErrRoomOutOfBounds is returned when a room extends beyond the grid boundaries.
var ErrRoomOutOfBounds = errors.New("room extends out of grid bounds")

// ErrRoomOverlap is returned when a room overlaps with an existing non-rock cell.
var ErrRoomOverlap = errors.New("room overlaps with existing structure")

// ErrRoomOnImpassable is returned when a room would be placed on an impassable cell (HardRock or Water).
var ErrRoomOnImpassable = errors.New("room placement blocked by impassable terrain")

// CanPlaceRoom reports whether the given room can be placed on the grid.
// It checks that all cells within the room bounds are inside the grid,
// and that every cell under the room is Rock (unexcavated).
// Impassable cells (HardRock, Water) also block placement.
func CanPlaceRoom(grid *Grid, room *Room) bool {
	for y := room.Pos.Y; y < room.Pos.Y+room.Height; y++ {
		for x := room.Pos.X; x < room.Pos.X+room.Width; x++ {
			pos := types.Pos{X: x, Y: y}
			if !grid.InBounds(pos) {
				return false
			}
			cell, _ := grid.At(pos)
			if cell.Type != Rock {
				return false
			}
		}
	}
	return true
}

// PlaceRoom places the given room on the grid by setting all cells within
// the room bounds to RoomFloor with the room's ID. Entrance cells are set
// to Entrance type.
// It returns an error if the room cannot be placed.
func PlaceRoom(grid *Grid, room *Room) error {
	// Validate bounds
	for y := room.Pos.Y; y < room.Pos.Y+room.Height; y++ {
		for x := room.Pos.X; x < room.Pos.X+room.Width; x++ {
			pos := types.Pos{X: x, Y: y}
			if !grid.InBounds(pos) {
				return fmt.Errorf("room at (%d,%d) size %dx%d: %w", room.Pos.X, room.Pos.Y, room.Width, room.Height, ErrRoomOutOfBounds)
			}
			cell, _ := grid.At(pos)
			if cell.Type.IsImpassable() {
				return fmt.Errorf("room at (%d,%d) size %dx%d blocked by %s at (%d,%d): %w", room.Pos.X, room.Pos.Y, room.Width, room.Height, cell.Type, x, y, ErrRoomOnImpassable)
			}
			if cell.Type != Rock {
				return fmt.Errorf("room at (%d,%d) size %dx%d conflicts at (%d,%d): %w", room.Pos.X, room.Pos.Y, room.Width, room.Height, x, y, ErrRoomOverlap)
			}
		}
	}

	// Place room floor cells
	for y := room.Pos.Y; y < room.Pos.Y+room.Height; y++ {
		for x := room.Pos.X; x < room.Pos.X+room.Width; x++ {
			pos := types.Pos{X: x, Y: y}
			_ = grid.Set(pos, Cell{Type: RoomFloor, RoomID: room.ID})
		}
	}

	// Place entrance cells
	for _, ent := range room.Entrances {
		if grid.InBounds(ent.Pos) {
			_ = grid.Set(ent.Pos, Cell{Type: Entrance, RoomID: room.ID})
		}
	}

	return nil
}
