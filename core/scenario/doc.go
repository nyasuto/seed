// Package scenario defines scenario data structures for the chaosseed-core engine.
//
// A scenario specifies "what to do" — initial cave layout, win/lose conditions,
// invasion wave schedules, events, and gameplay constraints — but not "how to
// run it". Execution is handled by the simulation package (Phase 7).
//
// Key concepts:
//
//   - Scenario: top-level definition containing all configuration for a game session.
//   - InitialState: cave dimensions, terrain seed, pre-built rooms, dragon veins,
//     starting chi, and initial beasts.
//   - ConditionDef / ConditionEvaluator: data-driven win/lose conditions evaluated
//     against a read-only GameSnapshot each tick.
//   - WaveScheduleEntry: defines when and what invaders appear.
//   - EventDef: scripted events triggered by game state.
//   - GameConstraints: limits on rooms, beasts, ticks, and forbidden room types.
package scenario
