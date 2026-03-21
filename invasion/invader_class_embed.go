package invasion

import _ "embed"

//go:embed invader_class_data.json
var defaultInvaderClassData []byte

// LoadDefaultInvaderClasses loads the built-in invader class definitions
// embedded from invader_class_data.json.
func LoadDefaultInvaderClasses() (*InvaderClassRegistry, error) {
	return LoadInvaderClassesJSON(defaultInvaderClassData)
}
