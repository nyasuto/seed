// Package simulation integrates all chaosseed-core subsystems into a unified
// tick-based game loop.
//
// It orchestrates the world, fengshui, senju, invasion, economy, and scenario
// packages to run a complete game session. The simulation is fully deterministic:
// given the same RNG seed and player actions, the outcome is identical.
//
// Key concepts:
//
//   - GameState: holds every subsystem engine, the active scenario, progress
//     tracker, economy state, and the deterministic RNG. This is the single
//     source of truth for a running game.
//   - GameStatus / GameResult: Running, Won, or Lost — with the final tick
//     and reason when the game ends.
//   - PlayerAction: the set of actions a player (or AI) can take each tick,
//     such as digging rooms, connecting corridors, placing beasts, or upgrading.
//   - SimulationEngine: the main loop driver. Each Step executes one tick:
//     validate and apply player actions, update chi flow, grow beasts, process
//     invasions, settle economy, fire scripted events, and evaluate win/lose
//     conditions.
//   - Checkpoint / Replay: save and restore game state for undo or deterministic
//     replay of recorded sessions.
//   - AIPlayer: pluggable AI interface for automated play and balance testing.
//   - SimulationRunner: high-level entry point for CLI and batch execution.
package simulation
