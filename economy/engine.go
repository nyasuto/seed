package economy

import (
	"fmt"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// EconomyTickResult holds the result of one tick of economic processing.
type EconomyTickResult struct {
	// Tick is the tick number this result corresponds to.
	Tick types.Tick
	// Supply is the chi supply generated this tick.
	Supply float64
	// Maintenance is the maintenance cost breakdown for this tick.
	Maintenance MaintenanceBreakdown
	// DeficitResult holds the deficit processing outcome.
	DeficitResult DeficitResult
	// Balance is the chi pool balance after this tick.
	Balance float64
	// ChiPoolCap is the chi pool capacity after recalculation.
	ChiPoolCap float64
}

// EconomyEngine orchestrates all economic calculations for the dungeon.
type EconomyEngine struct {
	// ChiPool is the central chi resource pool.
	ChiPool *ChiPool
	// SupplyCalc calculates per-tick chi supply.
	SupplyCalc *SupplyCalculator
	// MaintenanceCalc computes per-tick maintenance costs.
	MaintenanceCalc *MaintenanceCalculator
	// DeficitProc handles deficit penalties when maintenance exceeds balance.
	DeficitProc *DeficitProcessor
	// Construction holds construction cost tables.
	Construction *ConstructionCost
	// Beast holds beast summoning cost tables.
	Beast *BeastCost
	// CostParams holds chi pool cap and maintenance parameters.
	CostParams *CostParams
}

// NewEconomyEngine creates an EconomyEngine with the given components.
func NewEconomyEngine(
	chiPool *ChiPool,
	supplyParams *SupplyParams,
	costParams *CostParams,
	deficitParams *DeficitParams,
	constructionCost *ConstructionCost,
	beastCost *BeastCost,
) *EconomyEngine {
	return &EconomyEngine{
		ChiPool:         chiPool,
		SupplyCalc:      NewSupplyCalculator(supplyParams),
		MaintenanceCalc: NewMaintenanceCalculator(costParams),
		DeficitProc:     NewDeficitProcessor(deficitParams),
		Construction:    constructionCost,
		Beast:           beastCost,
		CostParams:      costParams,
	}
}

// Tick processes one tick of the economy.
//
//  1. Recalculate chi pool cap (storage room changes).
//  2. Calculate chi supply and deposit into pool.
//  3. Calculate maintenance costs.
//  4. Process deficit if maintenance exceeds balance.
//  5. If no deficit, withdraw maintenance from pool.
//  6. Record balance.
func (e *EconomyEngine) Tick(
	tick types.Tick,
	veins []fengshui.DragonVein,
	roomChis map[int]*fengshui.RoomChi,
	caveScore float64,
	rooms []world.Room,
	beastCount int,
	trapCount int,
) EconomyTickResult {
	// 1. Recalculate chi pool cap.
	newCap := CalcChiPoolCap(rooms, e.CostParams)
	e.ChiPool.Cap = newCap

	// 2. Calculate supply and deposit.
	supply := e.SupplyCalc.CalcTickSupply(veins, roomChis, caveScore)
	if supply > 0 {
		_ = e.ChiPool.Deposit(supply, Supply, "tick supply", tick)
	}

	// 3. Calculate maintenance.
	maintenance := e.MaintenanceCalc.CalcTickMaintenance(rooms, beastCount, trapCount)

	// 4-5. Process deficit (handles withdrawal and penalties).
	deficitResult := e.DeficitProc.ProcessDeficit(e.ChiPool, maintenance, tick)

	// 6. Record result.
	return EconomyTickResult{
		Tick:          tick,
		Supply:        supply,
		Maintenance:   maintenance,
		DeficitResult: deficitResult,
		Balance:       e.ChiPool.Balance(),
		ChiPoolCap:    newCap,
	}
}

// TryBuildRoom checks if the chi pool can afford the room construction cost,
// withdraws the cost, and returns the amount spent.
func (e *EconomyEngine) TryBuildRoom(roomTypeID string, tick types.Tick) (float64, error) {
	cost := e.Construction.CalcRoomCost(roomTypeID)
	if cost <= 0 {
		return 0, fmt.Errorf("unknown room type: %s", roomTypeID)
	}
	if !e.ChiPool.CanAfford(cost) {
		return 0, ErrInsufficientChi
	}
	_ = e.ChiPool.Withdraw(cost, Construction, fmt.Sprintf("build room %s", roomTypeID), tick)
	return cost, nil
}

// TrySummonBeast checks if the chi pool can afford the beast summoning cost,
// withdraws the cost, and returns the amount spent.
func (e *EconomyEngine) TrySummonBeast(element types.Element, tick types.Tick) (float64, error) {
	cost := e.Beast.CalcSummonCost(element)
	if cost <= 0 {
		return 0, fmt.Errorf("unknown element for summoning: %v", element)
	}
	if !e.ChiPool.CanAfford(cost) {
		return 0, ErrInsufficientChi
	}
	_ = e.ChiPool.Withdraw(cost, BeastSummon, fmt.Sprintf("summon %s beast", element), tick)
	return cost, nil
}

// TryUpgradeRoom checks if the chi pool can afford the room upgrade cost,
// withdraws the cost, and returns the amount spent.
func (e *EconomyEngine) TryUpgradeRoom(roomTypeID string, currentLevel int, tick types.Tick) (float64, error) {
	cost := e.Construction.CalcUpgradeCost(roomTypeID, currentLevel)
	if cost <= 0 {
		return 0, fmt.Errorf("unknown room type for upgrade: %s", roomTypeID)
	}
	if !e.ChiPool.CanAfford(cost) {
		return 0, ErrInsufficientChi
	}
	_ = e.ChiPool.Withdraw(cost, RoomUpgrade, fmt.Sprintf("upgrade room %s lv%d", roomTypeID, currentLevel+1), tick)
	return cost, nil
}

// TryDigCorridor checks if the chi pool can afford the corridor digging cost,
// withdraws the cost, and returns the amount spent.
func (e *EconomyEngine) TryDigCorridor(pathLength int, tick types.Tick) (float64, error) {
	cost := e.Construction.CalcCorridorCost(pathLength)
	if cost <= 0 {
		return 0, nil
	}
	if !e.ChiPool.CanAfford(cost) {
		return 0, ErrInsufficientChi
	}
	_ = e.ChiPool.Withdraw(cost, Construction, fmt.Sprintf("dig corridor len=%d", pathLength), tick)
	return cost, nil
}
