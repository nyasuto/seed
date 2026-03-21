package economy

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed deficit_params_data.json
var deficitParamsJSON []byte

// DeficitParams holds the thresholds and penalty values for deficit processing.
type DeficitParams struct {
	// MildThreshold is the shortage ratio below which the deficit is considered mild.
	MildThreshold float64 `json:"mild_threshold"`
	// ModerateThreshold is the shortage ratio below which the deficit is considered moderate.
	ModerateThreshold float64 `json:"moderate_threshold"`
	// MildGrowthPenalty is the growth multiplier applied during mild deficit (e.g. 0.5 = half speed).
	MildGrowthPenalty float64 `json:"mild_growth_penalty"`
	// ModerateCapacityPenalty is the capacity multiplier applied during moderate deficit.
	ModerateCapacityPenalty float64 `json:"moderate_capacity_penalty"`
	// SevereHPDrain is the HP reduction applied to beasts during severe deficit.
	SevereHPDrain int `json:"severe_hp_drain"`
	// SevereTrapDisable controls whether traps are disabled during severe deficit.
	SevereTrapDisable bool `json:"severe_trap_disable"`
}

// DefaultDeficitParams returns the default DeficitParams loaded from the embedded JSON.
func DefaultDeficitParams() (*DeficitParams, error) {
	return LoadDeficitParams(deficitParamsJSON)
}

// LoadDeficitParams decodes DeficitParams from JSON data.
func LoadDeficitParams(data []byte) (*DeficitParams, error) {
	var p DeficitParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal deficit params: %w", err)
	}
	return &p, nil
}
