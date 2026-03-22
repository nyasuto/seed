package view

import (
	"image/color"
	"testing"
)

func TestSelectionPanel_HandleClick_HitsItem(t *testing.T) {
	items := []SelectionItem{
		{Label: "Alpha"},
		{Label: "Beta"},
		{Label: "Gamma"},
	}
	sp := NewSelectionPanel(400, 300, "Pick One", items)

	// Each button should be clickable at its center.
	for i, btn := range sp.buttons {
		cx := (btn.Rect.Min.X + btn.Rect.Max.X) / 2
		cy := (btn.Rect.Min.Y + btn.Rect.Max.Y) / 2
		idx, ok := sp.HandleClick(cx, cy)
		if !ok {
			t.Errorf("HandleClick at button %d center (%d,%d) returned ok=false", i, cx, cy)
			continue
		}
		if idx != i {
			t.Errorf("HandleClick at button %d center returned index %d", i, idx)
		}
	}
}

func TestSelectionPanel_HandleClick_MissesAll(t *testing.T) {
	items := []SelectionItem{{Label: "Only"}}
	sp := NewSelectionPanel(400, 300, "Pick", items)

	// Click far outside the panel.
	idx, ok := sp.HandleClick(0, 0)
	if ok {
		t.Errorf("HandleClick(0,0) returned ok=true, index=%d", idx)
	}
}

func TestSelectionPanel_Contains(t *testing.T) {
	items := []SelectionItem{
		{Label: "A"},
		{Label: "B"},
	}
	sp := NewSelectionPanel(400, 300, "Title", items)
	bounds := sp.Bounds()

	// Center of the panel should be contained.
	cx := (bounds.Min.X + bounds.Max.X) / 2
	cy := (bounds.Min.Y + bounds.Max.Y) / 2
	if !sp.Contains(cx, cy) {
		t.Errorf("Contains(%d,%d) at panel center = false, want true", cx, cy)
	}

	// Outside the panel.
	if sp.Contains(bounds.Min.X-10, bounds.Min.Y-10) {
		t.Error("Contains outside panel = true, want false")
	}
}

func TestSelectionPanel_ItemCount(t *testing.T) {
	items := []SelectionItem{
		{Label: "X"},
		{Label: "Y"},
		{Label: "Z"},
	}
	sp := NewSelectionPanel(200, 200, "Test", items)
	if sp.ItemCount() != 3 {
		t.Errorf("ItemCount() = %d, want 3", sp.ItemCount())
	}
}

func TestSelectionPanel_WithColors(t *testing.T) {
	items := []SelectionItem{
		{Label: "Fire", Color: color.RGBA{R: 0xFF, A: 0xFF}},
		{Label: "Water", Color: color.RGBA{B: 0xFF, A: 0xFF}},
	}
	sp := NewSelectionPanel(300, 300, "Element", items)

	// Should still be clickable.
	btn := sp.buttons[0]
	cx := (btn.Rect.Min.X + btn.Rect.Max.X) / 2
	cy := (btn.Rect.Min.Y + btn.Rect.Max.Y) / 2
	idx, ok := sp.HandleClick(cx, cy)
	if !ok || idx != 0 {
		t.Errorf("colored item click: index=%d, ok=%v, want 0/true", idx, ok)
	}
}

func TestSelectionPanel_ButtonsAreVerticallyStacked(t *testing.T) {
	items := []SelectionItem{
		{Label: "First"},
		{Label: "Second"},
		{Label: "Third"},
	}
	sp := NewSelectionPanel(400, 300, "Stack", items)

	for i := 1; i < len(sp.buttons); i++ {
		prev := sp.buttons[i-1]
		curr := sp.buttons[i]
		if curr.Rect.Min.Y <= prev.Rect.Min.Y {
			t.Errorf("button %d (y=%d) not below button %d (y=%d)",
				i, curr.Rect.Min.Y, i-1, prev.Rect.Min.Y)
		}
	}
}
