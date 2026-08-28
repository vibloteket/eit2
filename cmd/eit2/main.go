package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibloteket/eit2/internal/ui"
	"github.com/vibloteket/eit2/internal/version"
)

func main() {
	flags := flag.NewFlagSet("eit2", flag.ExitOnError)
	fullscreen := flags.Bool("fullscreen", false, "start in fullscreen mode")
	windowed := flags.Bool("windowed", false, "force windowed mode")
	showVersion := flags.Bool("version", false, "print version and exit")
	_ = flags.Parse(os.Args[1:])

	if *showVersion {
		fmt.Printf("Eit 2 v%s\n", version.Value)
		return
	}
	if *fullscreen && *windowed {
		log.Fatal("--fullscreen and --windowed cannot be used together")
	}

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Eit 2")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if *fullscreen {
		ebiten.SetFullscreen(true)
	}
	if err := ebiten.RunGame(ui.NewGame()); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
