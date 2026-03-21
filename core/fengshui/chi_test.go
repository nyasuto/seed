package fengshui

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func TestRoomChi_IsFull(t *testing.T) {
	tests := []struct {
		name     string
		chi      RoomChi
		expected bool
	}{
		{
			name:     "full when current equals capacity",
			chi:      RoomChi{RoomID: 1, Current: 100, Capacity: 100, Element: types.Wood},
			expected: true,
		},
		{
			name:     "full when current exceeds capacity",
			chi:      RoomChi{RoomID: 1, Current: 150, Capacity: 100, Element: types.Wood},
			expected: true,
		},
		{
			name:     "not full when current below capacity",
			chi:      RoomChi{RoomID: 1, Current: 50, Capacity: 100, Element: types.Wood},
			expected: false,
		},
		{
			name:     "full when capacity is zero",
			chi:      RoomChi{RoomID: 1, Current: 0, Capacity: 0, Element: types.Wood},
			expected: true,
		},
		{
			name:     "full when capacity is negative",
			chi:      RoomChi{RoomID: 1, Current: 0, Capacity: -10, Element: types.Wood},
			expected: true,
		},
		{
			name:     "not full when empty",
			chi:      RoomChi{RoomID: 1, Current: 0, Capacity: 100, Element: types.Wood},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.chi.IsFull(); got != tt.expected {
				t.Errorf("IsFull() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRoomChi_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		chi      RoomChi
		expected bool
	}{
		{
			name:     "empty when current is zero",
			chi:      RoomChi{RoomID: 1, Current: 0, Capacity: 100, Element: types.Fire},
			expected: true,
		},
		{
			name:     "empty when current is negative",
			chi:      RoomChi{RoomID: 1, Current: -5, Capacity: 100, Element: types.Fire},
			expected: true,
		},
		{
			name:     "not empty when current is positive",
			chi:      RoomChi{RoomID: 1, Current: 1, Capacity: 100, Element: types.Fire},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.chi.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRoomChi_Ratio(t *testing.T) {
	tests := []struct {
		name     string
		chi      RoomChi
		expected float64
	}{
		{
			name:     "half full",
			chi:      RoomChi{RoomID: 1, Current: 50, Capacity: 100, Element: types.Earth},
			expected: 0.5,
		},
		{
			name:     "completely full",
			chi:      RoomChi{RoomID: 1, Current: 100, Capacity: 100, Element: types.Earth},
			expected: 1.0,
		},
		{
			name:     "empty",
			chi:      RoomChi{RoomID: 1, Current: 0, Capacity: 100, Element: types.Earth},
			expected: 0.0,
		},
		{
			name:     "over capacity clamps to 1",
			chi:      RoomChi{RoomID: 1, Current: 150, Capacity: 100, Element: types.Earth},
			expected: 1.0,
		},
		{
			name:     "negative current clamps to 0",
			chi:      RoomChi{RoomID: 1, Current: -10, Capacity: 100, Element: types.Earth},
			expected: 0.0,
		},
		{
			name:     "zero capacity returns 0",
			chi:      RoomChi{RoomID: 1, Current: 50, Capacity: 0, Element: types.Earth},
			expected: 0.0,
		},
		{
			name:     "negative capacity returns 0",
			chi:      RoomChi{RoomID: 1, Current: 50, Capacity: -10, Element: types.Earth},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.chi.Ratio()
			if got != tt.expected {
				t.Errorf("Ratio() = %v, want %v", got, tt.expected)
			}
		})
	}
}
