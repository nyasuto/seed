package scenario

// EventDef defines a scripted event triggered by game state.
// When the Condition is met during a tick, the Commands are executed.
// If OneShot is true, the event fires at most once per scenario run.
type EventDef struct {
	// ID is a unique identifier for this event.
	ID string
	// Condition defines when this event should fire.
	Condition ConditionDef
	// Commands lists the actions to perform when the event fires.
	Commands []CommandDef
	// OneShot indicates the event should fire at most once.
	OneShot bool
}

// CommandDef defines an event command in data-driven form.
// Type identifies the kind of command, and Params holds
// type-specific parameters loaded from JSON scenario data.
type CommandDef struct {
	// Type is the command identifier used by the factory to instantiate
	// the corresponding EventCommand.
	Type string
	// Params holds command-specific parameters.
	Params map[string]any
}
