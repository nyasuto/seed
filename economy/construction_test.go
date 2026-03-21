package economy

import "testing"

func TestConstructionCost_DefaultLoad(t *testing.T) {
	c := DefaultConstructionCost()
	if c == nil {
		t.Fatal("DefaultConstructionCost returned nil")
	}
	if len(c.RoomCost) == 0 {
		t.Error("RoomCost should not be empty")
	}
}

func TestConstructionCost_CalcRoomCost(t *testing.T) {
	c := DefaultConstructionCost()

	tests := []struct {
		roomTypeID string
		want       float64
	}{
		{"dragon_den", 50.0},
		{"chi_storage", 20.0},
		{"beast_room", 15.0},
		{"trap_room", 25.0},
		{"recovery_room", 20.0},
		{"warehouse", 10.0},
		{"unknown_type", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.roomTypeID, func(t *testing.T) {
			got := c.CalcRoomCost(tt.roomTypeID)
			if got != tt.want {
				t.Errorf("CalcRoomCost(%q) = %v, want %v", tt.roomTypeID, got, tt.want)
			}
		})
	}
}

func TestConstructionCost_CalcCorridorCost(t *testing.T) {
	c := DefaultConstructionCost()

	tests := []struct {
		name       string
		pathLength int
		want       float64
	}{
		{"zero length", 0, 0.0},
		{"single cell", 1, 2.0},
		{"five cells", 5, 10.0},
		{"ten cells", 10, 20.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.CalcCorridorCost(tt.pathLength)
			if got != tt.want {
				t.Errorf("CalcCorridorCost(%d) = %v, want %v", tt.pathLength, got, tt.want)
			}
		})
	}
}

func TestConstructionCost_CalcUpgradeCost(t *testing.T) {
	c := DefaultConstructionCost()

	tests := []struct {
		name         string
		roomTypeID   string
		currentLevel int
		want         float64
	}{
		{"dragon_den level 1", "dragon_den", 1, 60.0},  // 50 + 1*10
		{"dragon_den level 3", "dragon_den", 3, 80.0},  // 50 + 3*10
		{"warehouse level 1", "warehouse", 1, 20.0},    // 10 + 1*10
		{"warehouse level 5", "warehouse", 5, 60.0},    // 10 + 5*10
		{"chi_storage level 0", "chi_storage", 0, 20.0}, // 20 + 0*10
		{"unknown type level 2", "unknown", 2, 20.0},   // 0 + 2*10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.CalcUpgradeCost(tt.roomTypeID, tt.currentLevel)
			if got != tt.want {
				t.Errorf("CalcUpgradeCost(%q, %d) = %v, want %v", tt.roomTypeID, tt.currentLevel, got, tt.want)
			}
		})
	}
}

func TestConstructionCost_LoadInvalidJSON(t *testing.T) {
	_, err := LoadConstructionCost([]byte("not json"))
	if err == nil {
		t.Error("LoadConstructionCost should return error for invalid JSON")
	}
}
