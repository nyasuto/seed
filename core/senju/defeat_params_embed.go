package senju

import _ "embed"

//go:embed defeat_params_data.json
var defaultDefeatParamsData []byte

// LoadDefaultDefeatParams loads the built-in defeat parameters.
func LoadDefaultDefeatParams() (*DefeatParams, error) {
	return LoadDefeatParams(defaultDefeatParamsData)
}
