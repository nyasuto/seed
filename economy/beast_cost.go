package economy

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

//go:embed beast_cost_data.json
var defaultBeastCostJSON []byte

// beastCostJSON is the JSON representation used for marshaling/unmarshaling.
// Element keys are stored as their String() names (e.g. "Wood", "Fire").
type beastCostJSON struct {
	SummonCostByElement map[string]float64 `json:"summon_cost_by_element"`
}

// BeastCost holds the cost parameters for summoning beasts.
type BeastCost struct {
	// SummonCostByElement maps elements to their summoning cost in chi.
	SummonCostByElement map[types.Element]float64
}

// DefaultBeastCost returns the default beast cost loaded from embedded JSON.
func DefaultBeastCost() *BeastCost {
	bc, err := LoadBeastCost(defaultBeastCostJSON)
	if err != nil {
		panic(fmt.Sprintf("failed to load embedded beast cost: %v", err))
	}
	return bc
}

// LoadBeastCost parses beast cost parameters from JSON data.
func LoadBeastCost(data []byte) (*BeastCost, error) {
	var raw beastCostJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal beast cost: %w", err)
	}
	bc := &BeastCost{
		SummonCostByElement: make(map[types.Element]float64, len(raw.SummonCostByElement)),
	}
	for name, cost := range raw.SummonCostByElement {
		elem, err := parseElement(name)
		if err != nil {
			return nil, fmt.Errorf("beast cost: %w", err)
		}
		bc.SummonCostByElement[elem] = cost
	}
	return bc, nil
}

// CalcSummonCost returns the summoning cost for a beast of the given element.
// Returns 0 if the element is not found.
func (bc *BeastCost) CalcSummonCost(element types.Element) float64 {
	return bc.SummonCostByElement[element]
}

// parseElement converts an element name string to a types.Element.
func parseElement(name string) (types.Element, error) {
	for e := range types.Element(types.ElementCount) {
		if e.String() == name {
			return e, nil
		}
	}
	return 0, fmt.Errorf("unknown element: %q", name)
}
