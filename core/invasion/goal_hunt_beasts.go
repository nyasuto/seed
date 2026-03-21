package invasion

import (
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// DefaultHuntBeastsKillCount is the number of beasts an invader must defeat
// to achieve the HuntBeasts goal.
const DefaultHuntBeastsKillCount = 2

// HuntBeastsGoal directs an invader to seek out rooms containing beasts
// and defeat a certain number of them.
type HuntBeastsGoal struct {
	// RequiredKills is how many beasts the invader must defeat.
	RequiredKills int
	// Kills tracks how many beasts have been defeated so far.
	Kills int
}

// NewHuntBeastsGoal creates a HuntBeastsGoal with the default kill count.
func NewHuntBeastsGoal() *HuntBeastsGoal {
	return &HuntBeastsGoal{
		RequiredKills: DefaultHuntBeastsKillCount,
	}
}

// Type returns HuntBeasts.
func (g *HuntBeastsGoal) Type() GoalType {
	return HuntBeasts
}

// TargetRoomID returns the ID of the nearest known room containing beasts.
// If no beast rooms are known via memory, all rooms in the cave are scanned
// for rooms that currently have beasts. Returns 0 if no valid target is found.
func (g *HuntBeastsGoal) TargetRoomID(cave *world.Cave, invader *Invader, memory *ExplorationMemory) int {
	// First check memory for known beast rooms.
	if memory != nil && len(memory.KnownBeastRooms) > 0 {
		bestID := 0
		bestDist := -1
		for roomID, hasBeast := range memory.KnownBeastRooms {
			if !hasBeast {
				continue
			}
			dist := roomDistance(cave, invader.CurrentRoomID, roomID)
			if bestID == 0 || dist < bestDist {
				bestID = roomID
				bestDist = dist
			}
		}
		if bestID != 0 {
			return bestID
		}
	}

	// Fallback: scan cave rooms for any room with beasts.
	bestID := 0
	bestDist := -1
	for _, room := range cave.Rooms {
		if room.BeastCount() > 0 && room.ID != invader.CurrentRoomID {
			dist := roomDistance(cave, invader.CurrentRoomID, room.ID)
			if bestID == 0 || dist < bestDist {
				bestID = room.ID
				bestDist = dist
			}
		}
	}
	return bestID
}

// IsAchieved returns true when the invader has defeated enough beasts.
func (g *HuntBeastsGoal) IsAchieved(_ *world.Cave, _ *Invader) bool {
	return g.Kills >= g.RequiredKills
}

// roomDistance returns a simple distance metric between two rooms in the cave.
// It uses the Manhattan distance between room center positions. If either room
// is not found, it returns a large value.
func roomDistance(cave *world.Cave, fromID, toID int) int {
	var fromRoom, toRoom *world.Room
	for _, r := range cave.Rooms {
		if r.ID == fromID {
			fromRoom = r
		}
		if r.ID == toID {
			toRoom = r
		}
	}
	if fromRoom == nil || toRoom == nil {
		return 999999
	}
	fromCenter := types.Pos{X: fromRoom.Pos.X + fromRoom.Width/2, Y: fromRoom.Pos.Y + fromRoom.Height/2}
	toCenter := types.Pos{X: toRoom.Pos.X + toRoom.Width/2, Y: toRoom.Pos.Y + toRoom.Height/2}
	return fromCenter.Distance(toCenter)
}
