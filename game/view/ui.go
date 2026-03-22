package view

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/game/asset"
)

// TopBarData holds the values needed to render the top status bar.
type TopBarData struct {
	ChiPool    int
	MaxChiPool int
	CoreHP     int
	MaxCoreHP  int
	Tick       int
}

// FormatChiPool returns the ChiPool display string (e.g., "Chi: 150/500").
func FormatChiPool(d TopBarData) string {
	return fmt.Sprintf("Chi: %d/%d", d.ChiPool, d.MaxChiPool)
}

// FormatCoreHP returns the CoreHP display string (e.g., "CoreHP: 80/100").
func FormatCoreHP(d TopBarData) string {
	return fmt.Sprintf("CoreHP: %d/%d", d.CoreHP, d.MaxCoreHP)
}

// FormatTick returns the tick display string (e.g., "Tick: 42").
func FormatTick(d TopBarData) string {
	return fmt.Sprintf("Tick: %d", d.Tick)
}

// BarWidth calculates the pixel width of a progress bar given current/max
// values and the total bar width in pixels.
func BarWidth(current, max, totalWidth int) int {
	if max <= 0 || current <= 0 {
		return 0
	}
	if current >= max {
		return totalWidth
	}
	return current * totalWidth / max
}

const (
	topBarHeight   = 28
	topBarPaddingX = 8
	topBarPaddingY = 6
	hpBarWidth     = 100
	hpBarHeight    = 12
	sectionGap     = 24
)

// TopBar renders the top status bar showing ChiPool, CoreHP, and Tick count.
type TopBar struct{}

// Draw renders the top bar onto the screen with the given data.
func (tb *TopBar) Draw(screen *ebiten.Image, data TopBarData) {
	sw := screen.Bounds().Dx()

	// Draw background.
	bg := ebiten.NewImage(sw, topBarHeight)
	bg.Fill(asset.ColorUIBackground)
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(bg, op)

	// Draw bottom border.
	border := ebiten.NewImage(sw, 1)
	border.Fill(asset.ColorUIBorder)
	bop := &ebiten.DrawImageOptions{}
	bop.GeoM.Translate(0, float64(topBarHeight-1))
	screen.DrawImage(border, bop)

	x := topBarPaddingX
	y := topBarPaddingY

	// Chi Pool text.
	chiText := FormatChiPool(data)
	DrawText(screen, chiText, x, y)
	x += TextWidth(chiText) + sectionGap

	// CoreHP label + bar.
	hpLabel := FormatCoreHP(data)
	DrawText(screen, hpLabel, x, y)
	x += TextWidth(hpLabel) + 8

	// HP bar background (dark).
	barY := y + 2
	barBg := ebiten.NewImage(hpBarWidth, hpBarHeight)
	barBg.Fill(color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF})
	bop2 := &ebiten.DrawImageOptions{}
	bop2.GeoM.Translate(float64(x), float64(barY))
	screen.DrawImage(barBg, bop2)

	// HP bar foreground (colored by health percentage).
	fillW := BarWidth(data.CoreHP, data.MaxCoreHP, hpBarWidth)
	if fillW > 0 {
		barFg := ebiten.NewImage(fillW, hpBarHeight)
		barFg.Fill(hpBarColor(data.CoreHP, data.MaxCoreHP))
		fop := &ebiten.DrawImageOptions{}
		fop.GeoM.Translate(float64(x), float64(barY))
		screen.DrawImage(barFg, fop)
	}

	x += hpBarWidth + sectionGap

	// Tick count.
	tickText := FormatTick(data)
	DrawText(screen, tickText, x, y)
}

// hpBarColor returns a color based on the HP ratio:
// green when healthy, yellow when moderate, red when critical.
func hpBarColor(current, max int) color.RGBA {
	if max <= 0 {
		return color.RGBA{R: 0x66, G: 0x66, B: 0x66, A: 0xFF}
	}
	ratio := float64(current) / float64(max)
	switch {
	case ratio > 0.5:
		return color.RGBA{R: 0x4C, G: 0xAF, B: 0x50, A: 0xFF} // green
	case ratio > 0.25:
		return color.RGBA{R: 0xFF, G: 0xB3, B: 0x00, A: 0xFF} // yellow
	default:
		return color.RGBA{R: 0xE5, G: 0x39, B: 0x35, A: 0xFF} // red
	}
}
