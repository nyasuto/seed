package batch

import (
	"testing"

	"github.com/nyasuto/seed/sim/server"
)

func TestParseSweepParam(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		wantKey  string
		wantVals []string
		wantErr  bool
	}{
		{
			name:     "simple float values",
			spec:     "initial_state.starting_chi=100.0,200.0,300.0",
			wantKey:  "initial_state.starting_chi",
			wantVals: []string{"100.0", "200.0", "300.0"},
		},
		{
			name:     "single value",
			spec:     "constraints.max_ticks=500",
			wantKey:  "constraints.max_ticks",
			wantVals: []string{"500"},
		},
		{
			name:     "with spaces",
			spec:     "  key = a , b , c  ",
			wantKey:  "key",
			wantVals: []string{"a", "b", "c"},
		},
		{
			name:    "missing equals",
			spec:    "no_equals_sign",
			wantErr: true,
		},
		{
			name:    "empty key",
			spec:    "=1,2,3",
			wantErr: true,
		},
		{
			name:    "no values",
			spec:    "key=",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := ParseSweepParam(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSweepParam(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if param.Key != tt.wantKey {
				t.Errorf("Key = %q, want %q", param.Key, tt.wantKey)
			}
			if len(param.Values) != len(tt.wantVals) {
				t.Fatalf("Values count = %d, want %d", len(param.Values), len(tt.wantVals))
			}
			for i, v := range param.Values {
				if v != tt.wantVals[i] {
					t.Errorf("Values[%d] = %q, want %q", i, v, tt.wantVals[i])
				}
			}
		})
	}
}

func TestSetJSONPath(t *testing.T) {
	input := []byte(`{"a":{"b":1},"c":2}`)

	result, err := setJSONPath(input, "a.b", "42")
	if err != nil {
		t.Fatalf("setJSONPath: %v", err)
	}

	// Verify modified value.
	expected := `"b":42`
	if !contains(string(result), expected) {
		t.Errorf("result %s should contain %s", result, expected)
	}
}

func TestSetJSONPath_NewKey(t *testing.T) {
	input := []byte(`{"a":{"b":1}}`)

	result, err := setJSONPath(input, "a.new_key", "3.14")
	if err != nil {
		t.Fatalf("setJSONPath: %v", err)
	}

	if !contains(string(result), `"new_key":3.14`) {
		t.Errorf("result %s should contain new_key:3.14", result)
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		input string
		want  any
	}{
		{"42", int64(42)},
		{"3.14", 3.14},
		{"true", true},
		{"false", false},
		{"hello", "hello"},
		{"0.5", 0.5},
	}

	for _, tt := range tests {
		got := parseValue(tt.input)
		switch w := tt.want.(type) {
		case int64:
			g, ok := got.(int64)
			if !ok || g != w {
				t.Errorf("parseValue(%q) = %v (%T), want %v", tt.input, got, got, w)
			}
		case float64:
			g, ok := got.(float64)
			if !ok || g != w {
				t.Errorf("parseValue(%q) = %v (%T), want %v", tt.input, got, got, w)
			}
		case bool:
			g, ok := got.(bool)
			if !ok || g != w {
				t.Errorf("parseValue(%q) = %v (%T), want %v", tt.input, got, got, w)
			}
		case string:
			g, ok := got.(string)
			if !ok || g != w {
				t.Errorf("parseValue(%q) = %v (%T), want %v", tt.input, got, got, w)
			}
		}
	}
}

func TestRunSweep_StartingChi(t *testing.T) {
	scenarioJSON, err := server.LoadBuiltinScenarioJSON("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenarioJSON: %v", err)
	}

	param := SweepParam{
		Key:    "initial_state.starting_chi",
		Values: []string{"100.0", "200.0", "500.0"},
	}

	baseConfig := BatchConfig{
		Games:    10,
		BaseSeed: 42,
		AI:       AINoop,
		Parallel: 2,
	}

	results, err := RunSweep(scenarioJSON, param, baseConfig)
	if err != nil {
		t.Fatalf("RunSweep: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("RunSweep returned %d results, want 3", len(results))
	}

	for i, r := range results {
		if r.ParamKey != "initial_state.starting_chi" {
			t.Errorf("result[%d].ParamKey = %q, want initial_state.starting_chi", i, r.ParamKey)
		}
		if r.ParamValue != param.Values[i] {
			t.Errorf("result[%d].ParamValue = %q, want %q", i, r.ParamValue, param.Values[i])
		}
		if r.Result == nil {
			t.Errorf("result[%d].Result is nil", i)
			continue
		}
		if len(r.Result.Summaries) != 10 {
			t.Errorf("result[%d].Summaries count = %d, want 10", i, len(r.Result.Summaries))
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
