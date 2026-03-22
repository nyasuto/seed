// Package main is the entry point for the ChaosForge GUI client.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1088
	screenHeight = 728
)

// Game implements ebiten.Game interface.
type Game struct{}

// Update proceeds the game state.
func (g *Game) Update() error {
	return nil
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
}

// Layout returns the game's logical screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("ChaosForge")

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
