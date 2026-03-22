package asset_test

import (
	"testing"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/asset"
)

// allCellTypes returns every defined CellType value.
func allCellTypes() []world.CellType {
	return []world.CellType{
		world.Rock,
		world.CorridorFloor,
		world.RoomFloor,
		world.Entrance,
		world.HardRock,
		world.Water,
	}
}

// allElements returns every defined Element value.
func allElements() []types.Element {
	return []types.Element{
		types.Wood,
		types.Fire,
		types.Earth,
		types.Metal,
		types.Water,
	}
}

func TestPlaceholderProvider_GetTile_NonNil(t *testing.T) {
	p := asset.NewPlaceholderProvider()
	for _, ct := range allCellTypes() {
		for _, el := range allElements() {
			img := p.GetTile(ct, el)
			if img == nil {
				t.Errorf("GetTile(%v, %v) returned nil", ct, el)
			}
		}
	}
}

func TestPlaceholderProvider_GetTile_Size32x32(t *testing.T) {
	p := asset.NewPlaceholderProvider()
	for _, ct := range allCellTypes() {
		for _, el := range allElements() {
			img := p.GetTile(ct, el)
			w, h := img.Bounds().Dx(), img.Bounds().Dy()
			if w != 32 || h != 32 {
				t.Errorf("GetTile(%v, %v) size = %dx%d, want 32x32", ct, el, w, h)
			}
		}
	}
}

func TestPlaceholderProvider_ImplementsTilesetProvider(t *testing.T) {
	var _ asset.TilesetProvider = asset.NewPlaceholderProvider()
}

func TestPlaceholderProvider_GetBeastSprite_NonNil(t *testing.T) {
	p := asset.NewPlaceholderProvider()
	for _, species := range []string{"kirin", "suzaku", "genbu"} {
		img := p.GetBeastSprite(species, 1)
		if img == nil {
			t.Errorf("GetBeastSprite(%q, 1) returned nil", species)
		}
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		if w != 32 || h != 32 {
			t.Errorf("GetBeastSprite(%q, 1) size = %dx%d, want 32x32", species, w, h)
		}
	}
}

func TestPlaceholderProvider_GetInvaderSprite_NonNil(t *testing.T) {
	p := asset.NewPlaceholderProvider()
	for _, class := range []string{"warrior", "thief", "mage"} {
		img := p.GetInvaderSprite(class)
		if img == nil {
			t.Errorf("GetInvaderSprite(%q) returned nil", class)
		}
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		if w != 32 || h != 32 {
			t.Errorf("GetInvaderSprite(%q) size = %dx%d, want 32x32", class, w, h)
		}
	}
}
