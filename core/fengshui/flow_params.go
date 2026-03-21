package fengshui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed flow_params_data.json
var defaultFlowParamsJSON []byte

// FlowParams holds the parameters that govern chi flow mechanics:
// element interaction multipliers and base decay rate.
type FlowParams struct {
	// GeneratesMultiplier is the flow rate multiplier when the dragon vein's
	// element generates the room's element (productive cycle).
	GeneratesMultiplier float64 `json:"generates_multiplier"`
	// OvercomesMultiplier is the flow rate multiplier when the dragon vein's
	// element overcomes the room's element (destructive cycle).
	OvercomesMultiplier float64 `json:"overcomes_multiplier"`
	// SameElementMultiplier is the flow rate multiplier when the dragon vein's
	// element matches the room's element.
	SameElementMultiplier float64 `json:"same_element_multiplier"`
	// NeutralMultiplier is the flow rate multiplier when there is no special
	// elemental relationship.
	NeutralMultiplier float64 `json:"neutral_multiplier"`
	// BaseDecayRate is the fraction of chi lost per tick due to natural decay.
	BaseDecayRate float64 `json:"base_decay_rate"`
}

// DefaultFlowParams returns the default chi flow parameters loaded from
// the embedded JSON data.
func DefaultFlowParams() *FlowParams {
	var p FlowParams
	// Embedded data is guaranteed valid; ignore error.
	_ = json.Unmarshal(defaultFlowParamsJSON, &p)
	return &p
}

// LoadFlowParams reads chi flow parameters from a JSON file at the given path.
func LoadFlowParams(path string) (*FlowParams, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading flow params: %w", err)
	}
	var p FlowParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing flow params: %w", err)
	}
	return &p, nil
}
