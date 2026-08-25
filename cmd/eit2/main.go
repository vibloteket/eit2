package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibloteket/eit2/internal/ui"
)

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Eit 2")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(ui.NewGame()); err != nil {
		log.Fatal(err)
	}
}
