package economy

import (
	"errors"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
)

// TransactionType represents the category of a chi transaction.
type TransactionType int

const (
	// Supply is chi deposited from dragon vein supply.
	Supply TransactionType = iota
	// RoomMaintenance is chi spent on room upkeep.
	RoomMaintenance
	// BeastMaintenance is chi spent on beast upkeep.
	BeastMaintenance
	// TrapMaintenance is chi spent on trap upkeep.
	TrapMaintenance
	// Reward is chi gained as a reward.
	Reward
	// Theft is chi lost to theft by invaders.
	Theft
	// Construction is chi spent on building rooms or corridors.
	Construction
	// BeastSummon is chi spent on summoning a beast.
	BeastSummon
	// RoomUpgrade is chi spent on upgrading a room.
	RoomUpgrade
	// Deficit is a forced withdrawal due to deficit processing.
	Deficit
)

// ChiTransaction records a single chi pool transaction.
type ChiTransaction struct {
	Tick         types.Tick
	Amount       float64
	Type         TransactionType
	Reason       string
	BalanceAfter float64
}

// ChiPool represents the player's spendable chi resource.
// Current is clamped to [0, Cap].
type ChiPool struct {
	Current float64
	Cap     float64
	History []ChiTransaction
}

// NewChiPool creates a ChiPool with the given capacity and zero balance.
func NewChiPool(cap float64) *ChiPool {
	return &ChiPool{
		Cap: cap,
	}
}

// Balance returns the current chi balance.
func (p *ChiPool) Balance() float64 {
	return p.Current
}

// CanAfford returns true if the pool has at least the given amount.
func (p *ChiPool) CanAfford(amount float64) bool {
	return p.Current >= amount
}

// Deposit adds chi to the pool. The result is clamped to Cap.
func (p *ChiPool) Deposit(amount float64, txType TransactionType, reason string, tick types.Tick) error {
	if amount < 0 {
		return errors.New("deposit amount must be non-negative")
	}
	p.Current += amount
	if p.Current > p.Cap {
		p.Current = p.Cap
	}
	p.History = append(p.History, ChiTransaction{
		Tick:         tick,
		Amount:       amount,
		Type:         txType,
		Reason:       reason,
		BalanceAfter: p.Current,
	})
	return nil
}

// ErrInsufficientChi is returned when a withdrawal exceeds the available balance.
// The withdrawal still proceeds for the available amount (partial withdrawal).
var ErrInsufficientChi = errors.New("insufficient chi")

// Withdraw removes chi from the pool. If the requested amount exceeds the
// current balance, only the available amount is withdrawn and the balance
// is set to zero. In this case, ErrInsufficientChi is returned with the
// shortage amount in the error message.
func (p *ChiPool) Withdraw(amount float64, txType TransactionType, reason string, tick types.Tick) error {
	if amount < 0 {
		return errors.New("withdraw amount must be non-negative")
	}
	var err error
	actual := amount
	if amount > p.Current {
		shortage := amount - p.Current
		actual = p.Current
		err = fmt.Errorf("%w: shortage %.6f", ErrInsufficientChi, shortage)
	}
	p.Current -= actual
	p.History = append(p.History, ChiTransaction{
		Tick:         tick,
		Amount:       actual,
		Type:         txType,
		Reason:       reason,
		BalanceAfter: p.Current,
	})
	return err
}
