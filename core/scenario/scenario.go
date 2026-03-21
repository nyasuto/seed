package scenario

// Scenario defines a complete game session configuration.
// It specifies the initial cave layout, win/lose conditions, invasion wave
// schedules, scripted events, and gameplay constraints. Execution of the
// scenario is handled by the simulation package.
type Scenario struct {
	ID             string
	Name           string
	Description    string
	Difficulty     string
	InitialState   InitialState
	WinConditions  []ConditionDef
	LoseConditions []ConditionDef
	WaveSchedule   []WaveScheduleEntry
	Events         []EventDef
	Constraints    GameConstraints
}
