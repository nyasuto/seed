package balance

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ApplyParameter modifies a parameter in a scenario JSON file and creates a backup.
// scenarioPath is the path to the scenario JSON file.
// key is the dotted path into the JSON (e.g. "initial_state.starting_chi").
// value is the string representation of the new value.
// Returns the backup file path on success.
func ApplyParameter(scenarioPath, key, value string) (string, error) {
	data, err := os.ReadFile(scenarioPath)
	if err != nil {
		return "", fmt.Errorf("reading scenario file: %w", err)
	}

	// Create backup before modifying.
	backupPath := scenarioPath + ".bak"
	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return "", fmt.Errorf("creating backup: %w", err)
	}

	modified, err := setJSONPath(data, key, value)
	if err != nil {
		return backupPath, fmt.Errorf("setting parameter %s=%s: %w", key, value, err)
	}

	if err := os.WriteFile(scenarioPath, modified, 0o644); err != nil {
		return backupPath, fmt.Errorf("writing modified scenario: %w", err)
	}

	return backupPath, nil
}

// setJSONPath modifies a value at a dotted path in raw JSON.
// The value string is auto-detected as number, boolean, or string.
func setJSONPath(data []byte, dottedKey string, value string) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("unmarshaling scenario JSON: %w", err)
	}

	parts := strings.Split(dottedKey, ".")
	if err := setNestedValue(root, parts, parseValue(value)); err != nil {
		return nil, err
	}

	return json.MarshalIndent(root, "", "  ")
}

// setNestedValue sets a value at a nested path in a map.
func setNestedValue(m map[string]any, path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("empty path")
	}

	if len(path) == 1 {
		m[path[0]] = value
		return nil
	}

	next, ok := m[path[0]]
	if !ok {
		child := make(map[string]any)
		m[path[0]] = child
		return setNestedValue(child, path[1:], value)
	}

	child, ok := next.(map[string]any)
	if !ok {
		return fmt.Errorf("path element %q is not an object", path[0])
	}

	return setNestedValue(child, path[1:], value)
}

// parseValue attempts to convert a string to a float64, int, or bool.
func parseValue(s string) any {
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		if !strings.Contains(s, ".") {
			return i
		}
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	return s
}
