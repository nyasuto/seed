// Package invasion implements the invader system for the chaosseed-core engine.
//
// This package handles invader species definitions, goal-oriented AI,
// pathfinding, combat resolution, and invasion wave management.
// Invaders are external enemies that enter the cave seeking to destroy the
// dragon vein core, hunt beasts, or steal treasures.
//
// Key concepts:
//   - InvaderClass: defines base stats, element, preferred goal, and retreat behavior
//   - Invader: an individual instance of a class with exploration memory and state
//   - Goal: target-driven AI (destroy core, hunt beasts, steal treasure)
//   - Pathfinder: A*-based room-to-room navigation with fog-of-war exploration
//   - Combat: turn-based resolution between beasts/traps and invaders
//   - Wave: scheduled groups of invaders with escalating difficulty
package invasion
