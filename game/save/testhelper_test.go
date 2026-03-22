package save

import (
	_ "embed"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
)

//go:embed testdata/tutorial.json
var tutorialJSON []byte

func loadTestScenario(t *testing.T) []byte {
	t.Helper()
	return tutorialJSON
}

func loadScenario(t *testing.T, data []byte) *scenario.Scenario {
	t.Helper()
	sc, err := scenario.LoadScenario(data)
	if err != nil {
		t.Fatalf("LoadScenario: %v", err)
	}
	return sc
}
