package world

import (
	"errors"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

// ErrRoomNotFound is returned when a room ID does not exist in the cave.
var ErrRoomNotFound = errors.New("room not found")

// ErrNoEntrance is returned when a room has no entrances to connect.
var ErrNoEntrance = errors.New("room has no entrances")

// AddRoom validates, places, and registers a new room in the cave.
// It assigns an auto-incremented ID to the room and returns the placed room.
func (c *Cave) AddRoom(typeID string, pos types.Pos, w, h int, entrances []RoomEntrance) (*Room, error) {
	room := &Room{
		ID:        c.nextRoomID,
		TypeID:    typeID,
		Pos:       pos,
		Width:     w,
		Height:    h,
		Level:     1,
		Entrances: entrances,
	}

	if !CanPlaceRoom(c.Grid, room) {
		return nil, fmt.Errorf("cannot place room at (%d,%d) size %dx%d: %w", pos.X, pos.Y, w, h, ErrRoomOverlap)
	}

	if err := PlaceRoom(c.Grid, room); err != nil {
		return nil, fmt.Errorf("placing room: %w", err)
	}

	c.Rooms = append(c.Rooms, room)
	c.nextRoomID++
	return room, nil
}

// ConnectRooms connects two rooms by building a corridor between their closest
// entrance pairs. It returns the created corridor.
func (c *Cave) ConnectRooms(roomID1, roomID2 int) (Corridor, error) {
	room1 := c.RoomByID(roomID1)
	if room1 == nil {
		return Corridor{}, fmt.Errorf("room %d: %w", roomID1, ErrRoomNotFound)
	}
	room2 := c.RoomByID(roomID2)
	if room2 == nil {
		return Corridor{}, fmt.Errorf("room %d: %w", roomID2, ErrRoomNotFound)
	}

	if len(room1.Entrances) == 0 {
		return Corridor{}, fmt.Errorf("room %d: %w", roomID1, ErrNoEntrance)
	}
	if len(room2.Entrances) == 0 {
		return Corridor{}, fmt.Errorf("room %d: %w", roomID2, ErrNoEntrance)
	}

	// Find the closest pair of entrances
	var bestFrom, bestTo types.Pos
	bestDist := -1
	for _, e1 := range room1.Entrances {
		// Start from the cell just outside the entrance
		from := e1.Pos.Add(e1.Dir.Delta())
		for _, e2 := range room2.Entrances {
			to := e2.Pos.Add(e2.Dir.Delta())
			dist := from.Distance(to)
			if bestDist < 0 || dist < bestDist {
				bestDist = dist
				bestFrom = from
				bestTo = to
			}
		}
	}

	corridorID := c.nextCorridorID
	corridor, err := BuildCorridor(c.Grid, bestFrom, bestTo, corridorID, roomID1, roomID2)
	if err != nil {
		return Corridor{}, fmt.Errorf("connecting rooms %d and %d: %w", roomID1, roomID2, err)
	}

	c.Corridors = append(c.Corridors, corridor)
	c.nextCorridorID++
	return corridor, nil
}
