package view

import (
	"testing"

	"github.com/nyasuto/seed/game/asset"
)

func TestCellToScreen_Origin(t *testing.T) {
	mv := NewMapView(0, 0)
	px, py := mv.CellToScreen(0, 0)
	if px != 0 || py != 0 {
		t.Errorf("CellToScreen(0,0) = (%d,%d), want (0,0)", px, py)
	}
}

func TestCellToScreen_WithOffset(t *testing.T) {
	mv := NewMapView(10, 20)
	px, py := mv.CellToScreen(3, 5)
	wantX := 3*asset.TileSize + 10
	wantY := 5*asset.TileSize + 20
	if px != wantX || py != wantY {
		t.Errorf("CellToScreen(3,5) = (%d,%d), want (%d,%d)", px, py, wantX, wantY)
	}
}

func TestCellToScreen_NoOffset(t *testing.T) {
	mv := NewMapView(0, 0)
	px, py := mv.CellToScreen(10, 5)
	wantX := 10 * asset.TileSize // 320
	wantY := 5 * asset.TileSize  // 160
	if px != wantX || py != wantY {
		t.Errorf("CellToScreen(10,5) = (%d,%d), want (%d,%d)", px, py, wantX, wantY)
	}
}

func TestScreenToCell_Basic(t *testing.T) {
	mv := NewMapView(0, 0)
	cx, cy, ok := mv.ScreenToCell(320, 160, 16, 16)
	if !ok {
		t.Fatal("ScreenToCell(320,160) returned ok=false")
	}
	if cx != 10 || cy != 5 {
		t.Errorf("ScreenToCell(320,160) = (%d,%d), want (10,5)", cx, cy)
	}
}

func TestScreenToCell_WithOffset(t *testing.T) {
	mv := NewMapView(10, 20)
	// Screen pixel (10,20) should map to cell (0,0).
	cx, cy, ok := mv.ScreenToCell(10, 20, 16, 16)
	if !ok {
		t.Fatal("ScreenToCell returned ok=false")
	}
	if cx != 0 || cy != 0 {
		t.Errorf("ScreenToCell(10,20) = (%d,%d), want (0,0)", cx, cy)
	}
}

func TestScreenToCell_WithinTile(t *testing.T) {
	mv := NewMapView(0, 0)
	// Clicking anywhere within a tile should return the same cell.
	cx, cy, ok := mv.ScreenToCell(asset.TileSize+15, asset.TileSize*2+31, 16, 16)
	if !ok {
		t.Fatal("ScreenToCell returned ok=false")
	}
	if cx != 1 || cy != 2 {
		t.Errorf("got (%d,%d), want (1,2)", cx, cy)
	}
}

func TestScreenToCell_OutOfBounds_Negative(t *testing.T) {
	mv := NewMapView(10, 20)
	// Pixel before offset → invalid.
	_, _, ok := mv.ScreenToCell(5, 10, 16, 16)
	if ok {
		t.Error("ScreenToCell before offset should return ok=false")
	}
}

func TestScreenToCell_OutOfBounds_BeyondGrid(t *testing.T) {
	mv := NewMapView(0, 0)
	// Grid is 16x16, so cell (16,0) is out of bounds.
	_, _, ok := mv.ScreenToCell(16*asset.TileSize, 0, 16, 16)
	if ok {
		t.Error("ScreenToCell beyond grid should return ok=false")
	}
}

func TestScreenToCell_BoundaryLastCell(t *testing.T) {
	mv := NewMapView(0, 0)
	// Last valid cell (15,15) in a 16x16 grid.
	cx, cy, ok := mv.ScreenToCell(15*asset.TileSize, 15*asset.TileSize, 16, 16)
	if !ok {
		t.Fatal("ScreenToCell at last cell returned ok=false")
	}
	if cx != 15 || cy != 15 {
		t.Errorf("got (%d,%d), want (15,15)", cx, cy)
	}
}

func TestRoundTrip_CellToScreenToCell(t *testing.T) {
	mv := NewMapView(16, 32)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			px, py := mv.CellToScreen(x, y)
			cx, cy, ok := mv.ScreenToCell(px, py, 16, 16)
			if !ok {
				t.Fatalf("round trip failed for cell (%d,%d): ok=false", x, y)
			}
			if cx != x || cy != y {
				t.Errorf("round trip (%d,%d) → screen → (%d,%d)", x, y, cx, cy)
			}
		}
	}
}
