package senju

import (
	"errors"
	"fmt"

	"github.com/ponpoko/chaosseed-core/world"
)

// Placement errors.
var (
	ErrRoomFull          = errors.New("room is at beast capacity")
	ErrRoomNotAllowed    = errors.New("room type does not allow beasts")
	ErrBeastNotInRoom    = errors.New("beast is not in the specified room")
	ErrBeastAlreadyInRoom = errors.New("beast is already placed in a room")
)

// PlaceBeast places a beast into a room, validating capacity and room type constraints.
// The beast must not already be assigned to a room (RoomID == 0).
func PlaceBeast(beast *Beast, room *world.Room, roomType world.RoomType) error {
	if beast.RoomID != 0 {
		return fmt.Errorf("%w: beast %d is in room %d", ErrBeastAlreadyInRoom, beast.ID, beast.RoomID)
	}
	if roomType.MaxBeasts <= 0 {
		return fmt.Errorf("%w: %s", ErrRoomNotAllowed, roomType.ID)
	}
	if !room.HasBeastCapacity(roomType) {
		return fmt.Errorf("%w: room %d (max %d)", ErrRoomFull, room.ID, roomType.MaxBeasts)
	}

	beast.RoomID = room.ID
	room.BeastIDs = append(room.BeastIDs, beast.ID)
	return nil
}

// RemoveBeast removes a beast from its current room.
func RemoveBeast(beast *Beast, room *world.Room) error {
	if beast.RoomID != room.ID {
		return fmt.Errorf("%w: beast %d has roomID %d, not %d", ErrBeastNotInRoom, beast.ID, beast.RoomID, room.ID)
	}

	// Remove beast ID from room's list.
	for i, id := range room.BeastIDs {
		if id == beast.ID {
			room.BeastIDs = append(room.BeastIDs[:i], room.BeastIDs[i+1:]...)
			break
		}
	}
	beast.RoomID = 0
	return nil
}

// MoveBeast moves a beast from one room to another.
func MoveBeast(beast *Beast, fromRoom *world.Room, toRoom *world.Room, toRoomType world.RoomType) error {
	if err := RemoveBeast(beast, fromRoom); err != nil {
		return fmt.Errorf("move beast: remove: %w", err)
	}
	if err := PlaceBeast(beast, toRoom, toRoomType); err != nil {
		// Rollback: put beast back in the original room.
		beast.RoomID = fromRoom.ID
		fromRoom.BeastIDs = append(fromRoom.BeastIDs, beast.ID)
		return fmt.Errorf("move beast: place: %w", err)
	}
	return nil
}
