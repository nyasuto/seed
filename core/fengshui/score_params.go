package fengshui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed score_params_data.json
var defaultScoreParamsJSON []byte

// ScoreParams holds the parameters used to calculate feng shui scores for rooms.
type ScoreParams struct {
	// GeneratesBonus is added for each adjacent room whose element is generated
	// by this room's element (productive cycle).
	GeneratesBonus float64 `json:"generates_bonus"`
	// OvercomesPenalty is added (typically negative) for each adjacent room whose
	// element is overcome by this room's element.
	OvercomesPenalty float64 `json:"overcomes_penalty"`
	// SameElementBonus is added for each adjacent room with the same element.
	SameElementBonus float64 `json:"same_element_bonus"`
	// DragonVeinBonus is added if the room is on any dragon vein's path.
	DragonVeinBonus float64 `json:"dragon_vein_bonus"`
	// ChiRatioWeight is multiplied by the room's chi fill ratio to produce
	// the chi score component.
	ChiRatioWeight float64 `json:"chi_ratio_weight"`
}

// DefaultScoreParams returns the default scoring parameters loaded from
// the embedded JSON data.
func DefaultScoreParams() *ScoreParams {
	var p ScoreParams
	// Embedded data is guaranteed valid; ignore error.
	_ = json.Unmarshal(defaultScoreParamsJSON, &p)
	return &p
}

// MaxRoomScore returns the theoretical maximum feng shui score for a single
// room, assuming full chi, dragon vein connectivity, and maxNeighbors adjacent
// rooms with optimal elemental relationships. Use this to normalize CaveTotal
// into a [0,1] range.
func (p *ScoreParams) MaxRoomScore(maxNeighbors int) float64 {
	adjBonus := p.GeneratesBonus
	if p.SameElementBonus > adjBonus {
		adjBonus = p.SameElementBonus
	}
	return p.ChiRatioWeight + p.DragonVeinBonus + float64(maxNeighbors)*adjBonus
}

// LoadScoreParams reads scoring parameters from a JSON file at the given path.
func LoadScoreParams(path string) (*ScoreParams, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading score params: %w", err)
	}
	var p ScoreParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing score params: %w", err)
	}
	return &p, nil
}
