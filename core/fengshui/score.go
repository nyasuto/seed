package fengshui

// FengShuiScore holds the breakdown of a room's feng shui evaluation.
type FengShuiScore struct {
	// RoomID identifies the room being scored.
	RoomID int
	// ChiScore is derived from the room's chi fill ratio times ChiRatioWeight.
	ChiScore float64
	// AdjacencyScore is the sum of elemental bonuses/penalties from adjacent rooms.
	AdjacencyScore float64
	// DragonVeinScore is the bonus for being on a dragon vein's path.
	DragonVeinScore float64
	// Total is the sum of all score components.
	Total float64
}
