package input

import (
	"testing"
)

// stubMapView implements CellConverter for testing.
type stubMapView struct {
	offsetX int
	offsetY int
}

func (s *stubMapView) ScreenToCell(px, py, gridWidth, gridHeight int) (cx, cy int, ok bool) {
	cx = (px - s.offsetX) / 32
	cy = (py - s.offsetY) / 32
	if px < s.offsetX || py < s.offsetY || cx < 0 || cy < 0 || cx >= gridWidth || cy >= gridHeight {
		return 0, 0, false
	}
	return cx, cy, true
}

func TestMouseTracker_CursorCell_DefaultInvalid(t *testing.T) {
	mv := &stubMapView{offsetX: 32, offsetY: 32}
	mt := NewMouseTracker(mv, 24, 20)

	// Before any Update call, cursor cell should be invalid
	// because default (0,0) screen position maps to before the offset.
	cx, cy, ok := mt.CursorCell()
	if ok {
		t.Errorf("expected invalid before Update, got (%d,%d)", cx, cy)
	}
	_ = cx
	_ = cy
}

func TestMouseTracker_ScreenToCellConversion(t *testing.T) {
	// This tests the coordinate conversion that MouseTracker relies on.
	// MouseTracker delegates to CellConverter.ScreenToCell, so we verify the
	// expected mapping: screen(320+offset, 160+offset) → cell(10, 5).
	mv := &stubMapView{offsetX: 0, offsetY: 0}

	tests := []struct {
		name    string
		px, py  int
		wantCX  int
		wantCY  int
		wantOK  bool
		gridW   int
		gridH   int
	}{
		{
			name:   "exact cell boundary",
			px:     320, py: 160,
			wantCX: 10, wantCY: 5, wantOK: true,
			gridW: 24, gridH: 20,
		},
		{
			name:   "within tile",
			px:     335, py: 175,
			wantCX: 10, wantCY: 5, wantOK: true,
			gridW: 24, gridH: 20,
		},
		{
			name:   "origin cell",
			px:     0, py: 0,
			wantCX: 0, wantCY: 0, wantOK: true,
			gridW: 24, gridH: 20,
		},
		{
			name:   "beyond grid width",
			px:     24 * 32, py: 0,
			wantCX: 0, wantCY: 0, wantOK: false,
			gridW: 24, gridH: 20,
		},
		{
			name:   "beyond grid height",
			px:     0, py: 20 * 32,
			wantCX: 0, wantCY: 0, wantOK: false,
			gridW: 24, gridH: 20,
		},
		{
			name:   "negative coordinates",
			px:     -1, py: -1,
			wantCX: 0, wantCY: 0, wantOK: false,
			gridW: 24, gridH: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cx, cy, ok := mv.ScreenToCell(tt.px, tt.py, tt.gridW, tt.gridH)
			if ok != tt.wantOK {
				t.Errorf("ScreenToCell(%d,%d) ok=%v, want %v", tt.px, tt.py, ok, tt.wantOK)
			}
			if ok && (cx != tt.wantCX || cy != tt.wantCY) {
				t.Errorf("ScreenToCell(%d,%d) = (%d,%d), want (%d,%d)", tt.px, tt.py, cx, cy, tt.wantCX, tt.wantCY)
			}
		})
	}
}

func TestMouseTracker_WithOffset(t *testing.T) {
	// With offset (32, 32), screen pixel (352, 192) should map to cell (10, 5).
	mv := &stubMapView{offsetX: 32, offsetY: 32}

	cx, cy, ok := mv.ScreenToCell(352, 192, 24, 20)
	if !ok {
		t.Fatal("ScreenToCell(352,192) returned ok=false")
	}
	if cx != 10 || cy != 5 {
		t.Errorf("ScreenToCell(352,192) = (%d,%d), want (10,5)", cx, cy)
	}
}

func TestMouseTracker_OutsideMap_BeforeOffset(t *testing.T) {
	mv := &stubMapView{offsetX: 32, offsetY: 32}

	// Pixels before the offset are outside the map.
	_, _, ok := mv.ScreenToCell(16, 16, 24, 20)
	if ok {
		t.Error("ScreenToCell before offset should return ok=false")
	}
}
