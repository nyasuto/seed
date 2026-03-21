package economy

import "github.com/ponpoko/chaosseed-core/world"

// MaintenanceBreakdown holds the per-tick maintenance cost broken down by category.
type MaintenanceBreakdown struct {
	// RoomCost is the total maintenance cost for all rooms.
	RoomCost float64
	// BeastCost is the total maintenance cost for all beasts.
	BeastCost float64
	// TrapCost is the total maintenance cost for all traps.
	TrapCost float64
	// Total is the sum of all maintenance costs.
	Total float64
}

// MaintenanceCalculator computes per-tick maintenance costs.
type MaintenanceCalculator struct {
	params *CostParams
}

// NewMaintenanceCalculator creates a MaintenanceCalculator with the given parameters.
func NewMaintenanceCalculator(params *CostParams) *MaintenanceCalculator {
	return &MaintenanceCalculator{params: params}
}

// CalcTickMaintenance computes the total maintenance cost for a single tick.
// rooms is the list of rooms in the cave, beastCount is the total number of
// beasts, and trapCount is the total number of active traps.
func (mc *MaintenanceCalculator) CalcTickMaintenance(rooms []world.Room, beastCount int, trapCount int) MaintenanceBreakdown {
	var roomCost float64
	for _, r := range rooms {
		if cost, ok := mc.params.RoomMaintenancePerTick[r.TypeID]; ok {
			roomCost += cost
		}
	}

	beastCost := float64(beastCount) * mc.params.BeastMaintenancePerTick
	trapCost := float64(trapCount) * mc.params.TrapMaintenancePerTick

	return MaintenanceBreakdown{
		RoomCost:  roomCost,
		BeastCost: beastCost,
		TrapCost:  trapCost,
		Total:     roomCost + beastCost + trapCost,
	}
}

// storageRoomTypeID is the room type ID for chi storage rooms.
const storageRoomTypeID = "chi_storage"

// CalcChiPoolCap computes the chi pool capacity based on storage rooms.
// It sums the base capacity plus per-storage-room and per-storage-level bonuses.
func CalcChiPoolCap(rooms []world.Room, params *CostParams) float64 {
	cap := params.ChiPoolBaseCap
	for _, r := range rooms {
		if r.TypeID == storageRoomTypeID {
			cap += params.ChiPoolCapPerStorageRoom
			// Level bonus: additional capacity for levels above 1.
			if r.Level > 1 {
				cap += float64(r.Level-1) * params.ChiPoolCapPerStorageLevel
			}
		}
	}
	return cap
}
