package batch

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/nyasuto/seed/core/scenario"
)

// SweepParam represents a parsed parameter sweep specification.
// Format: "dotted.key=v1,v2,v3"
type SweepParam struct {
	// Key is the dotted path into the scenario JSON (e.g. "initial_state.starting_chi").
	Key string
	// Values is the list of string values to try.
	Values []string
}

// SweepResult holds the batch result for one parameter value.
type SweepResult struct {
	// ParamKey is the parameter path that was modified.
	ParamKey string
	// ParamValue is the string representation of the value used.
	ParamValue string
	// Result is the batch execution result.
	Result *BatchResult
}

// ParseSweepParam parses a sweep specification string in the format "key=v1,v2,v3".
func ParseSweepParam(spec string) (SweepParam, error) {
	eqIdx := strings.Index(spec, "=")
	if eqIdx < 0 {
		return SweepParam{}, fmt.Errorf("invalid sweep format %q: missing '='", spec)
	}

	key := strings.TrimSpace(spec[:eqIdx])
	if key == "" {
		return SweepParam{}, fmt.Errorf("invalid sweep format %q: empty key", spec)
	}

	valPart := strings.TrimSpace(spec[eqIdx+1:])
	if valPart == "" {
		return SweepParam{}, fmt.Errorf("invalid sweep format %q: no values", spec)
	}

	values := strings.Split(valPart, ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}

	return SweepParam{Key: key, Values: values}, nil
}

// RunSweep executes a batch for each value in the sweep parameter.
// scenarioJSON is the original scenario JSON bytes. For each sweep value,
// the specified key is modified in the JSON, the scenario is re-loaded,
// and a batch is executed with the given base config.
func RunSweep(scenarioJSON []byte, param SweepParam, baseConfig BatchConfig) ([]SweepResult, error) {
	results := make([]SweepResult, 0, len(param.Values))

	for _, valStr := range param.Values {
		modified, err := setJSONPath(scenarioJSON, param.Key, valStr)
		if err != nil {
			return nil, fmt.Errorf("setting %s=%s: %w", param.Key, valStr, err)
		}

		sc, err := scenario.LoadScenario(modified)
		if err != nil {
			return nil, fmt.Errorf("loading scenario with %s=%s: %w", param.Key, valStr, err)
		}

		config := baseConfig
		config.Scenario = sc

		runner, err := NewBatchRunner(config)
		if err != nil {
			return nil, fmt.Errorf("creating runner for %s=%s: %w", param.Key, valStr, err)
		}

		batchResult, err := runner.Run()
		if err != nil {
			return nil, fmt.Errorf("running batch for %s=%s: %w", param.Key, valStr, err)
		}

		results = append(results, SweepResult{
			ParamKey:   param.Key,
			ParamValue: valStr,
			Result:     batchResult,
		})
	}

	return results, nil
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

	return json.Marshal(root)
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
		// Create intermediate map if it doesn't exist.
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
// Falls back to string if no conversion matches.
func parseValue(s string) any {
	// Try bool.
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Try integer.
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		// Only use int if there's no decimal point.
		if !strings.Contains(s, ".") {
			return i
		}
	}

	// Try float.
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	return s
}
