package ui

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"

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
	white      = color.RGBA{R: 235, G: 241, B: 247, A: 255}
	muted      = color.RGBA{R: 148, G: 163, B: 184, A: 255}
)

type Game struct {
	Lobby      lobby.Lobby
	gamepadIDs []ebiten.GamepadID
	fontSource *text.GoTextFaceSource
}

func NewGame() *Game {
	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(fmt.Sprintf("load embedded font: %v", err))
	}
	return &Game{fontSource: fontSource}
}

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

func (g *Game) face(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: g.fontSource, Size: size}
}

func drawText(screen *ebiten.Image, value string, face *text.GoTextFace, x, y float64, colour color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(colour)
	text.Draw(screen, value, face, op)
}

func drawCenteredText(screen *ebiten.Image, value string, face *text.GoTextFace, centerX, y float64, colour color.Color) {
	width, _ := text.Measure(value, face, 0)
	drawText(screen, value, face, centerX-width/2, y, colour)
}

func fitText(value string, face *text.GoTextFace, maxWidth float64) string {
	width, _ := text.Measure(value, face, 0)
	if width <= maxWidth {
		return value
	}
	for len(value) > 1 {
		value = value[:len(value)-1]
		candidate := value + "…"
		width, _ = text.Measure(candidate, face, 0)
		if width <= maxWidth {
			return candidate
		}
	}
	return "…"
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(background)
	drawText(screen, "EIT 2", g.face(64), 40, 24, accent)
	drawText(screen, "Couch multiplayer prototype", g.face(27), 42, 92, white)

	const gap = 20
	const margin = 40
	width := (logicalWidth - margin*2 - gap*3) / lobby.MaxPlayers
	for i := 0; i < lobby.MaxPlayers; i++ {
		x := margin + i*(width+gap)
		centerX := float64(x + width/2)
		ebitenutil.DrawRect(screen, float64(x), 145, float64(width), 415, panel)
		ebitenutil.DrawRect(screen, float64(x), 145, float64(width), 5, accent)
		drawCenteredText(screen, fmt.Sprintf("PLAYER %d", i+1), g.face(28), centerX, 175, white)
		if i < len(g.Lobby.Slots) {
			device := g.Lobby.Slots[i].Device
			name := device.Name
			if name == "" {
				name = string(device.Kind)
			}
			name = fitText(name, g.face(22), float64(width-30))
			drawCenteredText(screen, name, g.face(22), centerX, 270, muted)
			drawCenteredText(screen, "READY", g.face(34), centerX, 325, accent)
		} else {
			drawCenteredText(screen, "PRESS A", g.face(30), centerX, 270, white)
			drawCenteredText(screen, "TO JOIN", g.face(30), centerX, 312, white)
		}
	}

	message := "Connect gamepads and press A  •  Enter joins one keyboard player"
	colour := muted
	if g.Lobby.CanStart() {
		message = "READY TO START — two or more players joined"
		colour = accent
	}
	drawCenteredText(screen, message, g.face(25), logicalWidth/2, 625, colour)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return logicalWidth, logicalHeight
}
