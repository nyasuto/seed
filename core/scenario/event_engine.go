package scenario

import "fmt"

// EventEngine evaluates event conditions against the current game state
// and returns commands for events whose conditions are met.
// It tracks fired one-shot events to prevent re-firing.
type EventEngine struct {
	// FiredEvents tracks IDs of one-shot events that have already fired.
	FiredEvents map[string]bool
}

// NewEventEngine creates a new EventEngine with empty fired-events tracking.
func NewEventEngine() *EventEngine {
	return &EventEngine{
		FiredEvents: make(map[string]bool),
	}
}

// Tick evaluates all event definitions against the given game snapshot.
// For each event whose condition is met (and has not already fired if OneShot),
// the corresponding commands are created and returned.
// One-shot events are recorded so they do not fire again.
func (e *EventEngine) Tick(snapshot GameSnapshot, events []EventDef) ([]EventCommand, error) {
	var commands []EventCommand

	for _, ev := range events {
		// Skip already-fired one-shot events.
		if ev.OneShot && e.FiredEvents[ev.ID] {
			continue
		}

		// Build the condition evaluator.
		cond, err := NewCondition(ev.Condition)
		if err != nil {
			return nil, fmt.Errorf("event %q condition: %w", ev.ID, err)
		}

		// Check if the condition is met.
		if !cond.Evaluate(snapshot) {
			continue
		}

		// Build commands for this event.
		for i, cmdDef := range ev.Commands {
			cmd, err := NewCommand(cmdDef)
			if err != nil {
				return nil, fmt.Errorf("event %q command[%d]: %w", ev.ID, i, err)
			}
			commands = append(commands, cmd)
		}

		// Record one-shot events.
		if ev.OneShot {
			e.FiredEvents[ev.ID] = true
		}
	}

	return commands, nil
}
