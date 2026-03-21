package types

// Element represents one of the five Chinese elements (五行).
type Element int

const (
	// Wood represents the wood element (木).
	Wood Element = iota
	// Fire represents the fire element (火).
	Fire
	// Earth represents the earth element (土).
	Earth
	// Metal represents the metal element (金).
	Metal
	// Water represents the water element (水).
	Water
)

// ElementCount is the total number of elements.
const ElementCount = 5

// String returns the English name of the element.
func (e Element) String() string {
	switch e {
	case Wood:
		return "Wood"
	case Fire:
		return "Fire"
	case Earth:
		return "Earth"
	case Metal:
		return "Metal"
	case Water:
		return "Water"
	default:
		return "Unknown"
	}
}
