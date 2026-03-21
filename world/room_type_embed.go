package world

import _ "embed"

//go:embed room_type_data.json
var defaultRoomTypeData []byte

// LoadDefaultRoomTypes loads the built-in room type definitions
// embedded from room_type_data.json.
func LoadDefaultRoomTypes() (*RoomTypeRegistry, error) {
	return LoadRoomTypesJSON(defaultRoomTypeData)
}
