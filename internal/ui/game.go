package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/vibloteket/eit2/internal/lobby"
)

const (
	logicalWidth  = 1280
	logicalHeight = 720
)

var (
	background = color.RGBA{R: 8, G: 12, B: 20, A: 255}
	panel      = color.RGBA{R: 18, G: 28, B: 42, A: 255}
	accent     = color.RGBA{R: 76, G: 230, B: 166, A: 255}
	muted      = color.RGBA{R: 148, G: 163, B: 184, A: 255}
)

type Game struct {
	Lobby      lobby.Lobby
	gamepadIDs []ebiten.GamepadID
}

func NewGame() *Game { return &Game{} }

func (g *Game) Update() error {
	g.gamepadIDs = ebiten.AppendGamepadIDs(g.gamepadIDs[:0])
	for _, id := range g.gamepadIDs {
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) {
			g.Lobby.Join(lobby.Device{Kind: lobby.DeviceGamepad, ID: int(id), Name: ebiten.GamepadName(id)})
		}
	}
	// Keyboard is a single input device and useful for desktop/web development.
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.Lobby.Join(lobby.Device{Kind: lobby.DeviceKeyboard, Name: "Keyboard"})
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(background)
	ebitenutil.DebugPrintAt(screen, "EIT 2", 30, 24)
	ebitenutil.DebugPrintAt(screen, "Couch multiplayer prototype", 30, 46)

	const gap = 20
	const margin = 40
	width := (logicalWidth - margin*2 - gap*3) / lobby.MaxPlayers
	for i := 0; i < lobby.MaxPlayers; i++ {
		x := margin + i*(width+gap)
		ebitenutil.DrawRect(screen, float64(x), 130, float64(width), 430, panel)
		ebitenutil.DrawRect(screen, float64(x), 130, float64(width), 3, accent)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("PLAYER %d", i+1), x+18, 160)
		if i < len(g.Lobby.Slots) {
			device := g.Lobby.Slots[i].Device
			name := device.Name
			if name == "" {
				name = string(device.Kind)
			}
			ebitenutil.DebugPrintAt(screen, name, x+18, 220)
			ebitenutil.DebugPrintAt(screen, "READY", x+18, 255)
		} else {
			ebitenutil.DebugPrintAt(screen, "PRESS A TO JOIN", x+18, 220)
		}
	}

	message := "Connect gamepads and press A. Enter joins one keyboard player."
	if g.Lobby.CanStart() {
		message = "READY TO START - two or more players joined"
	}
	ebitenutil.DebugPrintAt(screen, message, 40, 640)
	ebitenutil.DrawRect(screen, 40, 670, 1200, 2, muted)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return logicalWidth, logicalHeight
}
