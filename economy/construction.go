package economy

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed construction_data.json
var defaultConstructionCostJSON []byte

// ConstructionCost holds the cost parameters for building rooms, corridors, and upgrades.
type ConstructionCost struct {
	// RoomCost maps room type IDs to their construction cost in chi.
	RoomCost map[string]float64 `json:"room_cost"`
	// CorridorCostPerCell is the chi cost per cell of corridor.
	CorridorCostPerCell float64 `json:"corridor_cost_per_cell"`
	// RoomUpgradeCostBase maps room type IDs to their base upgrade cost.
	RoomUpgradeCostBase map[string]float64 `json:"room_upgrade_cost_base"`
	// RoomUpgradeCostPerLevel is the additional cost per existing level when upgrading.
	RoomUpgradeCostPerLevel float64 `json:"room_upgrade_cost_per_level"`
}

// DefaultConstructionCost returns the default construction cost loaded from embedded JSON.
func DefaultConstructionCost() *ConstructionCost {
	c, err := LoadConstructionCost(defaultConstructionCostJSON)
	if err != nil {
		panic(fmt.Sprintf("failed to load embedded construction cost: %v", err))
	}
	return c
}

// LoadConstructionCost parses construction cost parameters from JSON data.
func LoadConstructionCost(data []byte) (*ConstructionCost, error) {
	var c ConstructionCost
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("unmarshal construction cost: %w", err)
	}
	return &c, nil
}

// CalcRoomCost returns the construction cost for a room of the given type.
// Returns 0 if the room type is not found.
func (c *ConstructionCost) CalcRoomCost(roomTypeID string) float64 {
	return c.RoomCost[roomTypeID]
}

// CalcCorridorCost returns the construction cost for a corridor of the given path length.
func (c *ConstructionCost) CalcCorridorCost(pathLength int) float64 {
	return c.CorridorCostPerCell * float64(pathLength)
}

// CalcUpgradeCost returns the cost to upgrade a room from the given current level.
// The formula is: base cost + (current level * cost per level).
func (c *ConstructionCost) CalcUpgradeCost(roomTypeID string, currentLevel int) float64 {
	base := c.RoomUpgradeCostBase[roomTypeID]
	return base + float64(currentLevel)*c.RoomUpgradeCostPerLevel
}
