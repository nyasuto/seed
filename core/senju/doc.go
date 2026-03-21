// Package senju provides the beast (仙獣) system for the chaosseed-core engine.
//
// This package handles species definitions, beast placement into rooms,
// growth and leveling mechanics, and behavioral AI patterns.
// Beasts are elemental creatures that inhabit cave rooms, grow by consuming
// chi energy, and defend the cave against invaders.
//
// Key concepts:
//   - Species: defines base stats, element, and growth characteristics
//   - Beast: an individual instance of a species placed in a room
//   - Growth: beasts gain EXP by consuming chi, with element affinity bonuses
//   - Combat: effective stats are modified by room element compatibility
package senju
