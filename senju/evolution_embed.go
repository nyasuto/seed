package senju

import _ "embed"

//go:embed evolution_data.json
var defaultEvolutionData []byte

// LoadDefaultEvolution loads the built-in evolution path definitions.
func LoadDefaultEvolution() (*EvolutionRegistry, error) {
	return LoadEvolutionJSON(defaultEvolutionData)
}
