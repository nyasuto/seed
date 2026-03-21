package senju

import _ "embed"

//go:embed species_data.json
var defaultSpeciesData []byte

// LoadDefaultSpecies loads the built-in species definitions.
func LoadDefaultSpecies() (*SpeciesRegistry, error) {
	return LoadSpeciesJSON(defaultSpeciesData)
}
