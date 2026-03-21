package senju

import "encoding/json"

// DefeatParams holds the parameters that control beast defeat and revival behavior.
type DefeatParams struct {
	// StunnedDuration is the number of ticks a beast remains stunned after defeat.
	StunnedDuration int `json:"stunned_duration"`
	// RevivalHPRatio is the fraction of MaxHP restored upon revival (0.0–1.0).
	RevivalHPRatio float64 `json:"revival_hp_ratio"`
	// LevelPenalty is the number of levels lost upon defeat (minimum level is 1).
	LevelPenalty int `json:"level_penalty"`
}

// DefaultDefeatParams returns the default defeat parameters.
func DefaultDefeatParams() *DefeatParams {
	return &DefeatParams{
		StunnedDuration: 20,
		RevivalHPRatio:  0.3,
		LevelPenalty:    1,
	}
}

// LoadDefeatParams parses defeat parameters from JSON data.
func LoadDefeatParams(data []byte) (*DefeatParams, error) {
	var p DefeatParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
