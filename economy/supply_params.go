package economy

import (
	_ "embed"
	"encoding/json"
)

//go:embed supply_params_data.json
var defaultSupplyParamsJSON []byte

// SupplyParams holds the parameters that control chi supply calculation.
type SupplyParams struct {
	// BaseSupplyPerVein is the base chi supply per dragon vein per tick.
	BaseSupplyPerVein float64 `json:"base_supply_per_vein"`
	// FengShuiMinMultiplier is the minimum multiplier applied based on feng shui score.
	FengShuiMinMultiplier float64 `json:"feng_shui_min_multiplier"`
	// FengShuiMaxMultiplier is the maximum multiplier applied based on feng shui score.
	FengShuiMaxMultiplier float64 `json:"feng_shui_max_multiplier"`
	// ChiRatioSupplyWeight is the weight of average chi fill ratio on supply bonus.
	ChiRatioSupplyWeight float64 `json:"chi_ratio_supply_weight"`
}

// DefaultSupplyParams returns the default supply parameters loaded from embedded JSON.
func DefaultSupplyParams() (*SupplyParams, error) {
	return LoadSupplyParams(defaultSupplyParamsJSON)
}

// LoadSupplyParams parses supply parameters from JSON data.
func LoadSupplyParams(data []byte) (*SupplyParams, error) {
	var p SupplyParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
