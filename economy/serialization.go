package economy

import (
	"encoding/json"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

// jsonChiTransaction is the JSON representation of a ChiTransaction.
type jsonChiTransaction struct {
	Tick         uint64  `json:"tick"`
	Amount       float64 `json:"amount"`
	Type         int     `json:"type"`
	Reason       string  `json:"reason"`
	BalanceAfter float64 `json:"balance_after"`
}

// jsonChiPool is the JSON representation of a ChiPool.
type jsonChiPool struct {
	Current float64              `json:"current"`
	Cap     float64              `json:"cap"`
	History []jsonChiTransaction `json:"history"`
}

// jsonEconomyState is the top-level JSON representation for economy state.
type jsonEconomyState struct {
	ChiPool jsonChiPool `json:"chi_pool"`
}

// MarshalEconomyState serializes the EconomyEngine's persistent state to JSON.
// Only the ChiPool (balance, cap, and transaction history) is saved.
// Calculators and cost tables are reconstructed from parameters on unmarshal.
func MarshalEconomyState(engine *EconomyEngine) ([]byte, error) {
	if engine == nil {
		return nil, fmt.Errorf("engine must not be nil")
	}
	if engine.ChiPool == nil {
		return nil, fmt.Errorf("engine chi pool must not be nil")
	}

	history := make([]jsonChiTransaction, len(engine.ChiPool.History))
	for i, tx := range engine.ChiPool.History {
		history[i] = jsonChiTransaction{
			Tick:         uint64(tx.Tick),
			Amount:       tx.Amount,
			Type:         int(tx.Type),
			Reason:       tx.Reason,
			BalanceAfter: tx.BalanceAfter,
		}
	}

	state := jsonEconomyState{
		ChiPool: jsonChiPool{
			Current: engine.ChiPool.Current,
			Cap:     engine.ChiPool.Cap,
			History: history,
		},
	}

	return json.Marshal(state)
}

// UnmarshalEconomyState restores an EconomyEngine from JSON data.
// The parameters must be provided to reconstruct calculators and cost tables.
func UnmarshalEconomyState(
	data []byte,
	supplyParams *SupplyParams,
	costParams *CostParams,
	deficitParams *DeficitParams,
	constructionCost *ConstructionCost,
	beastCost *BeastCost,
) (*EconomyEngine, error) {
	var state jsonEconomyState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshalling economy state: %w", err)
	}

	history := make([]ChiTransaction, len(state.ChiPool.History))
	for i, jtx := range state.ChiPool.History {
		history[i] = ChiTransaction{
			Tick:         types.Tick(jtx.Tick),
			Amount:       jtx.Amount,
			Type:         TransactionType(jtx.Type),
			Reason:       jtx.Reason,
			BalanceAfter: jtx.BalanceAfter,
		}
	}

	chiPool := &ChiPool{
		Current: state.ChiPool.Current,
		Cap:     state.ChiPool.Cap,
		History: history,
	}

	engine := NewEconomyEngine(chiPool, supplyParams, costParams, deficitParams, constructionCost, beastCost)
	return engine, nil
}
