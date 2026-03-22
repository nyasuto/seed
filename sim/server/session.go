package server

import (
	"embed"
	"fmt"
	"os"

	"github.com/nyasuto/seed/core/scenario"
)

//go:embed scenarios/tutorial.json scenarios/standard.json
var builtinScenarios embed.FS

// builtinNames lists the recognized built-in scenario names.
var builtinNames = map[string]string{
	"tutorial": "scenarios/tutorial.json",
	"standard": "scenarios/standard.json",
}

// BuiltinScenarioNames returns the list of available built-in scenario names.
func BuiltinScenarioNames() []string {
	names := make([]string, 0, len(builtinNames))
	for name := range builtinNames {
		names = append(names, name)
	}
	return names
}

// LoadScenarioFromFile loads a scenario from a JSON file at the given path.
func LoadScenarioFromFile(path string) (*scenario.Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario file: %w", err)
	}
	sc, err := scenario.LoadScenario(data)
	if err != nil {
		return nil, fmt.Errorf("parse scenario %s: %w", path, err)
	}
	return sc, nil
}

// LoadBuiltinScenarioJSON returns the raw JSON bytes for a built-in scenario.
func LoadBuiltinScenarioJSON(name string) ([]byte, error) {
	path, ok := builtinNames[name]
	if !ok {
		return nil, fmt.Errorf("unknown builtin scenario: %q", name)
	}
	data, err := builtinScenarios.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read embedded scenario %s: %w", name, err)
	}
	return data, nil
}

// LoadBuiltinScenario loads a built-in scenario by name (e.g. "tutorial", "standard").
func LoadBuiltinScenario(name string) (*scenario.Scenario, error) {
	data, err := LoadBuiltinScenarioJSON(name)
	if err != nil {
		return nil, err
	}
	sc, err := scenario.LoadScenario(data)
	if err != nil {
		return nil, fmt.Errorf("parse builtin scenario %s: %w", name, err)
	}
	return sc, nil
}

// LoadScenario loads a scenario by name or file path.
// If name matches a built-in scenario, it is loaded from embedded data.
// Otherwise, it is treated as a file path.
func LoadScenario(nameOrPath string) (*scenario.Scenario, error) {
	if _, ok := builtinNames[nameOrPath]; ok {
		return LoadBuiltinScenario(nameOrPath)
	}
	return LoadScenarioFromFile(nameOrPath)
}
