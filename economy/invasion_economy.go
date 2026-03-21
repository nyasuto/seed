package economy

import (
	"fmt"

	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/types"
)

// InvasionEconomySummary records the economic impact of invasion events
// processed in a single call to ProcessInvasionEvents.
type InvasionEconomySummary struct {
	// RewardChi is the total chi gained from defeating invaders.
	RewardChi float64
	// StolenChi is the total chi lost to escaping invaders.
	StolenChi float64
	// NetChi is RewardChi - StolenChi.
	NetChi float64
	// BeastsLost is the number of beasts defeated during the invasion.
	BeastsLost int
}

// InvasionEconomyProcessor converts invasion events into ChiPool transactions.
type InvasionEconomyProcessor struct{}

// NewInvasionEconomyProcessor creates a new InvasionEconomyProcessor.
func NewInvasionEconomyProcessor() *InvasionEconomyProcessor {
	return &InvasionEconomyProcessor{}
}

// ProcessInvasionEvents processes a slice of invasion events and applies
// the economic effects to the given ChiPool.
//
// Processing rules:
//  1. InvaderDefeated events: sum RewardChi and deposit to ChiPool (TransactionType: Reward)
//  2. InvaderEscaped events: sum StolenChi and withdraw from ChiPool (TransactionType: Theft)
//  3. BeastDefeated events: count for summary (revival cost is a future extension)
func (p *InvasionEconomyProcessor) ProcessInvasionEvents(events []invasion.InvasionEvent, chiPool *ChiPool, tick types.Tick) InvasionEconomySummary {
	var summary InvasionEconomySummary

	for _, ev := range events {
		switch ev.Type {
		case invasion.InvaderDefeated:
			summary.RewardChi += ev.RewardChi
		case invasion.InvaderEscaped:
			summary.StolenChi += ev.StolenChi
		case invasion.BeastDefeated:
			summary.BeastsLost++
		}
	}

	summary.NetChi = summary.RewardChi - summary.StolenChi

	// Deposit rewards.
	if summary.RewardChi > 0 {
		chiPool.Deposit(summary.RewardChi, Reward, fmt.Sprintf("invasion reward at tick %d", tick), tick)
	}

	// Withdraw stolen chi.
	if summary.StolenChi > 0 {
		chiPool.Withdraw(summary.StolenChi, Theft, fmt.Sprintf("invasion theft at tick %d", tick), tick)
	}

	return summary
}
