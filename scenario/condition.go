package scenario

// ConditionDef defines a win or lose condition in data-driven form.
// Type identifies the kind of condition (e.g. "survive_until",
// "defeat_all_waves", "core_destroyed"), and Params holds type-specific
// parameters as key-value pairs loaded from JSON scenario data.
type ConditionDef struct {
	// Type is the condition identifier used by the factory to instantiate
	// the corresponding ConditionEvaluator.
	Type string
	// Params holds condition-specific parameters. For example a
	// "survive_until" condition might contain {"ticks": 3000}.
	Params map[string]any
}
