package senju

import _ "embed"

//go:embed growth_params_data.json
var defaultGrowthParamsData []byte

// LoadDefaultGrowthParams loads the built-in growth parameters.
func LoadDefaultGrowthParams() (*GrowthParams, error) {
	return LoadGrowthParams(defaultGrowthParamsData)
}
