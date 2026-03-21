package economy

import (
	_ "embed"
	"encoding/json"
)

//go:embed cost_params_data.json
var defaultCostParamsJSON []byte

// CostParams holds the parameters that control maintenance costs and chi pool capacity.
type CostParams struct {
	// RoomMaintenancePerTick maps room type IDs to their per-tick maintenance cost.
	RoomMaintenancePerTick map[string]float64 `json:"room_maintenance_per_tick"`
	// BeastMaintenancePerTick is the per-tick maintenance cost per beast.
	BeastMaintenancePerTick float64 `json:"beast_maintenance_per_tick"`
	// TrapMaintenancePerTick is the per-tick maintenance cost per trap.
	TrapMaintenancePerTick float64 `json:"trap_maintenance_per_tick"`
	// ChiPoolBaseCap is the base chi pool capacity without any storage rooms.
	ChiPoolBaseCap float64 `json:"chi_pool_base_cap"`
	// ChiPoolCapPerStorageRoom is the additional capacity per chi storage room.
	ChiPoolCapPerStorageRoom float64 `json:"chi_pool_cap_per_storage_room"`
	// ChiPoolCapPerStorageLevel is the additional capacity per storage room level.
	ChiPoolCapPerStorageLevel float64 `json:"chi_pool_cap_per_storage_level"`
}

// DefaultCostParams returns the default cost parameters loaded from embedded JSON.
func DefaultCostParams() *CostParams {
	p, _ := LoadCostParams(defaultCostParamsJSON)
	return p
}

// LoadCostParams parses cost parameters from JSON data.
func LoadCostParams(data []byte) (*CostParams, error) {
	var p CostParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
