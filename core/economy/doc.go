// Package economy implements the chi economy layer for the chaosseed-core engine.
//
// This package manages chi as a player-level economic resource (ChiPool),
// which is distinct from the room-level physical chi simulated by
// the fengshui package (ChiFlowEngine).
//
//   - ChiFlowEngine (fengshui package): simulates physical chi flow through
//     dragon veins, rooms, and corridors. Each room accumulates and transfers
//     chi based on elemental compatibility and adjacency.
//
//   - ChiPool (this package): represents the player's spendable chi resource,
//     used as currency for construction, beast summoning, upgrades, and other
//     player actions. Supply is derived from the overall state of the chi flow
//     system (dragon vein count, room fill ratios, feng shui score).
//
// The SupplyCalculator reads the physical chi state each tick and converts it
// into ChiPool deposits. The MaintenanceCalculator computes per-tick upkeep
// costs for rooms, beasts, and traps, which are withdrawn from the ChiPool.
// The DeficitProcessor handles cases where maintenance exceeds the available
// balance, applying graduated penalties.
package economy
