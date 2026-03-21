package scenario

import (
	"encoding/json"
	"fmt"

	"github.com/nyasuto/seed/core/types"
)

// jsonWaveResult is the JSON representation of WaveResult.
type jsonWaveResult struct {
	WaveID        int      `json:"wave_id"`
	Result        string   `json:"result"`
	CompletedTick types.Tick `json:"completed_tick"`
}

// jsonScenarioProgress is the JSON representation of ScenarioProgress.
type jsonScenarioProgress struct {
	ScenarioID    string           `json:"scenario_id"`
	CurrentTick   types.Tick       `json:"current_tick"`
	FiredEventIDs []string         `json:"fired_event_ids"`
	WaveResults   []jsonWaveResult `json:"wave_results"`
	CoreHP        int              `json:"core_hp"`
}

// MarshalProgress serializes a ScenarioProgress to JSON bytes.
func MarshalProgress(p *ScenarioProgress) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("cannot marshal nil progress")
	}

	jp := jsonScenarioProgress{
		ScenarioID:    p.ScenarioID,
		CurrentTick:   p.CurrentTick,
		FiredEventIDs: p.FiredEventIDs,
		WaveResults:   make([]jsonWaveResult, len(p.WaveResults)),
		CoreHP:        p.CoreHP,
	}

	for i, wr := range p.WaveResults {
		jp.WaveResults[i] = jsonWaveResult(wr)
	}

	data, err := json.Marshal(jp)
	if err != nil {
		return nil, fmt.Errorf("marshal progress: %w", err)
	}
	return data, nil
}

// UnmarshalProgress deserializes JSON bytes into a ScenarioProgress.
func UnmarshalProgress(data []byte) (*ScenarioProgress, error) {
	var jp jsonScenarioProgress
	if err := json.Unmarshal(data, &jp); err != nil {
		return nil, fmt.Errorf("unmarshal progress: %w", err)
	}

	p := &ScenarioProgress{
		ScenarioID:    jp.ScenarioID,
		CurrentTick:   jp.CurrentTick,
		FiredEventIDs: jp.FiredEventIDs,
		WaveResults:   make([]WaveResult, len(jp.WaveResults)),
		CoreHP:        jp.CoreHP,
	}

	for i, jwr := range jp.WaveResults {
		p.WaveResults[i] = WaveResult(jwr)
	}

	return p, nil
}
