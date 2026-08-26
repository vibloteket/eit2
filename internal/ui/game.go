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

	"github.com/vibloteket/eit2/internal/controls"
	core "github.com/vibloteket/eit2/internal/game"
	"github.com/vibloteket/eit2/internal/lobby"
)

const (
	logicalWidth  = 1280
	logicalHeight = 720
)

var (
	background  = color.RGBA{R: 8, G: 12, B: 20, A: 255}
	panel       = color.RGBA{R: 18, G: 28, B: 42, A: 255}
	accent      = color.RGBA{R: 76, G: 230, B: 166, A: 255}
	white       = color.RGBA{R: 235, G: 241, B: 247, A: 255}
	muted       = color.RGBA{R: 148, G: 163, B: 184, A: 255}
	pieceColors = [...]color.RGBA{
		{}, {R: 97, G: 218, B: 251, A: 255}, {R: 255, G: 209, B: 102, A: 255},
		{R: 139, G: 124, B: 246, A: 255}, {R: 255, G: 159, B: 67, A: 255},
		{R: 71, G: 120, B: 245, A: 255}, {R: 76, G: 230, B: 166, A: 255},
		{R: 255, G: 107, B: 107, A: 255},
	}
)

type view int

const (
	viewLobby view = iota
	viewPlay
)

type action = controls.Action

const (
	actionLeft  = controls.Left
	actionRight = controls.Right
	actionDown  = controls.Down
	actionCCW   = controls.RotateCCW
	actionCW    = controls.RotateCW
	actionDrop  = controls.HardDrop
)

type button struct {
	Rect  imageRect
	Label string
	Do    action
}

type imageRect struct{ X, Y, W, H int }

func (r imageRect) contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

type Game struct {
	Lobby       lobby.Lobby
	gamepadIDs  []ebiten.GamepadID
	touchIDs    []ebiten.TouchID
	pressedIDs  []ebiten.TouchID
	heldActions map[action]int
	fontSource  *text.GoTextFaceSource
	view        view
	players     []*core.Game
	touchDevice lobby.Device
}

func NewGame() *Game {
	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(fmt.Sprintf("load embedded font: %v", err))
	}
	return &Game{
		fontSource:  fontSource,
		heldActions: make(map[action]int),
		touchDevice: lobby.Device{Kind: lobby.DeviceTouch, Name: "Touch controls"},
	}
}

func (g *Game) Update() error {
	g.pressedIDs = inpututil.AppendJustPressedTouchIDs(g.pressedIDs[:0])
	if g.view == viewPlay {
		g.updatePlay()
		return nil
	}
	g.updateLobby()
	return nil
}

func (g *Game) updateLobby() {
	g.gamepadIDs = ebiten.AppendGamepadIDs(g.gamepadIDs[:0])
	for _, id := range g.gamepadIDs {
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) {
			g.Lobby.Join(lobby.Device{Kind: lobby.DeviceGamepad, ID: int(id), Name: ebiten.GamepadName(id)})
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.Lobby.CanStart() {
			g.start()
		} else {
			g.Lobby.Join(lobby.Device{Kind: lobby.DeviceKeyboard, Name: "Keyboard"})
		}
	}
	for _, id := range g.pressedIDs {
		x, y := ebiten.TouchPosition(id)
		if g.Lobby.CanStart() && startButton().contains(x, y) {
			g.start()
		} else {
			g.Lobby.Join(g.touchDevice)
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if g.Lobby.CanStart() && startButton().contains(x, y) {
			g.start()
		} else {
			g.Lobby.Join(g.touchDevice)
		}
	}
}

func (g *Game) start() {
	g.players = make([]*core.Game, len(g.Lobby.Slots))
	for i := range g.players {
		g.players[i] = core.New(uint64(i + 1))
	}
	g.view = viewPlay
}

func (g *Game) updatePlay() {
	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.view = viewLobby
		return
	}
	if len(g.players) == 0 {
		return
	}
	player := g.players[0]
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		player.Move(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		player.Move(1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		if ebiten.Tick()%2 == 0 {
			player.StepDown()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		player.Rotate(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		player.Rotate(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		player.HardDrop()
	}
	activeActions := make(map[action]bool)
	for _, id := range g.touchIDs {
		x, y := ebiten.TouchPosition(id)
		for _, control := range touchButtons() {
			if control.Rect.contains(x, y) {
				activeActions[control.Do] = true
			}
		}
	}
	for action := range g.heldActions {
		if !activeActions[action] {
			delete(g.heldActions, action)
		}
	}
	for action := range activeActions {
		ticks := g.heldActions[action]
		if controls.ShouldRepeat(action, ticks) {
			apply(player, action)
		}
		g.heldActions[action] = ticks + 1
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		for _, control := range touchButtons() {
			if control.Rect.contains(x, y) {
				apply(player, control.Do)
			}
		}
	}
	player.Tick()
}

func apply(game *core.Game, action action) {
	switch action {
	case actionLeft:
		game.Move(-1)
	case actionRight:
		game.Move(1)
	case actionDown:
		game.StepDown()
	case actionCCW:
		game.Rotate(-1)
	case actionCW:
		game.Rotate(1)
	case actionDrop:
		game.HardDrop()
	}
}

func startButton() imageRect { return imageRect{X: 490, Y: 590, W: 300, H: 85} }

func touchButtons() []button {
	return []button{
		{Rect: imageRect{X: 45, Y: 420, W: 145, H: 120}, Label: "LEFT", Do: actionLeft},
		{Rect: imageRect{X: 205, Y: 420, W: 145, H: 120}, Label: "RIGHT", Do: actionRight},
		{Rect: imageRect{X: 125, Y: 570, W: 145, H: 120}, Label: "DOWN", Do: actionDown},
		{Rect: imageRect{X: 930, Y: 420, W: 145, H: 120}, Label: "CCW", Do: actionCCW},
		{Rect: imageRect{X: 1090, Y: 420, W: 145, H: 120}, Label: "CW", Do: actionCW},
		{Rect: imageRect{X: 970, Y: 570, W: 225, H: 120}, Label: "DROP", Do: actionDrop},
	}
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
	if g.view == viewPlay {
		g.drawPlay(screen)
		return
	}
	g.drawLobby(screen)
}

func (g *Game) drawLobby(screen *ebiten.Image) {
	screen.Fill(background)
	drawText(screen, "EIT 2", g.face(64), 40, 24, accent)
	drawText(screen, "Tap anywhere or press A to join", g.face(27), 42, 92, white)
	const gap, margin = 20, 40
	width := (logicalWidth - margin*2 - gap*3) / lobby.MaxPlayers
	for i := 0; i < lobby.MaxPlayers; i++ {
		x := margin + i*(width+gap)
		centerX := float64(x + width/2)
		ebitenutil.DrawRect(screen, float64(x), 145, float64(width), 400, panel)
		ebitenutil.DrawRect(screen, float64(x), 145, float64(width), 5, accent)
		drawCenteredText(screen, fmt.Sprintf("PLAYER %d", i+1), g.face(28), centerX, 175, white)
		if i < len(g.Lobby.Slots) {
			name := g.Lobby.Slots[i].Device.Name
			name = fitText(name, g.face(22), float64(width-30))
			drawCenteredText(screen, name, g.face(22), centerX, 270, muted)
			drawCenteredText(screen, "READY", g.face(34), centerX, 325, accent)
		} else {
			drawCenteredText(screen, "PRESS A", g.face(30), centerX, 270, white)
			drawCenteredText(screen, "TO JOIN", g.face(30), centerX, 312, white)
		}
	}
	if g.Lobby.CanStart() {
		r := startButton()
		ebitenutil.DrawRect(screen, float64(r.X), float64(r.Y), float64(r.W), float64(r.H), accent)
		drawCenteredText(screen, "START", g.face(36), float64(r.X+r.W/2), float64(r.Y+20), background)
	} else {
		drawCenteredText(screen, "1–4 players", g.face(25), logicalWidth/2, 625, muted)
	}
}

func (g *Game) drawPlay(screen *ebiten.Image) {
	screen.Fill(background)
	if len(g.players) == 0 {
		return
	}
	game := g.players[0]
	const cell = 27
	boardW, boardH := core.BoardWidth*cell, core.BoardHeight*cell
	boardX, boardY := (logicalWidth-boardW)/2, 40
	ebitenutil.DrawRect(screen, float64(boardX-5), float64(boardY-5), float64(boardW+10), float64(boardH+10), muted)
	ebitenutil.DrawRect(screen, float64(boardX), float64(boardY), float64(boardW), float64(boardH), panel)
	for y, row := range game.Board {
		for x, value := range row {
			if value != 0 {
				drawCell(screen, boardX+x*cell, boardY+y*cell, cell, value)
			}
		}
	}
	for _, point := range game.Cells(game.Active) {
		if point.Y >= 0 {
			drawCell(screen, boardX+point.X*cell, boardY+point.Y*cell, cell, game.Active.Kind+1)
		}
	}
	drawText(screen, "PLAYER 1", g.face(27), 45, 48, white)
	drawText(screen, fmt.Sprintf("SCORE  %d", game.Score), g.face(25), 45, 88, white)
	drawText(screen, fmt.Sprintf("LINES  %d", game.Lines), g.face(25), 45, 124, white)
	drawText(screen, fmt.Sprintf("LEVEL  %d", game.Lines/5), g.face(25), 45, 160, white)

	drawText(screen, "NEXT", g.face(22), 1020, 48, muted)
	next := core.Piece{Kind: game.NextKind}
	for _, point := range core.PieceCells(next) {
		drawCell(screen, 1020+point.X*24, 82+point.Y*24, 24, game.NextKind+1)
	}
	drawText(screen, "STORED", g.face(20), 1020, 150, muted)
	drawText(screen, "—", g.face(30), 1020, 180, white)
	drawText(screen, "EFFECTS", g.face(20), 1020, 215, muted)
	drawText(screen, "None", g.face(22), 1020, 242, white)
	if game.GameOver {
		drawCenteredText(screen, "GAME OVER", g.face(54), logicalWidth/2, 275, white)
	}
	for _, control := range touchButtons() {
		r := control.Rect
		fill := panel
		if g.heldActions[control.Do] > 0 {
			fill = color.RGBA{R: 35, G: 73, B: 76, A: 255}
		}
		ebitenutil.DrawRect(screen, float64(r.X), float64(r.Y), float64(r.W), float64(r.H), fill)
		ebitenutil.DrawRect(screen, float64(r.X), float64(r.Y), float64(r.W), 4, accent)
		drawControlIcon(screen, control, g.face(22), fill)
	}
}

func drawControlIcon(screen *ebiten.Image, control button, labelFace *text.GoTextFace, fill color.RGBA) {
	r := control.Rect
	cx, cy := float64(r.X+r.W/2), float64(r.Y+r.H/2)
	switch control.Do {
	case actionLeft:
		ebitenutil.DrawLine(screen, cx+28, cy, cx-25, cy, white)
		ebitenutil.DrawLine(screen, cx-25, cy, cx-2, cy-22, white)
		ebitenutil.DrawLine(screen, cx-25, cy, cx-2, cy+22, white)
	case actionRight:
		ebitenutil.DrawLine(screen, cx-28, cy, cx+25, cy, white)
		ebitenutil.DrawLine(screen, cx+25, cy, cx+2, cy-22, white)
		ebitenutil.DrawLine(screen, cx+25, cy, cx+2, cy+22, white)
	case actionDown:
		ebitenutil.DrawLine(screen, cx, cy-25, cx, cy+25, white)
		ebitenutil.DrawLine(screen, cx, cy+25, cx-22, cy+2, white)
		ebitenutil.DrawLine(screen, cx, cy+25, cx+22, cy+2, white)
	case actionCCW, actionCW:
		// A blocky circular arrow avoids relying on a font glyph that might be
		// absent on mobile browsers.
		ebitenutil.DrawCircle(screen, cx, cy, 28, white)
		ebitenutil.DrawCircle(screen, cx, cy, 19, fill)
		if control.Do == actionCCW {
			ebitenutil.DrawLine(screen, cx-35, cy-22, cx-10, cy-28, white)
			ebitenutil.DrawLine(screen, cx-35, cy-22, cx-26, cy+2, white)
		} else {
			ebitenutil.DrawLine(screen, cx+35, cy-22, cx+10, cy-28, white)
			ebitenutil.DrawLine(screen, cx+35, cy-22, cx+26, cy+2, white)
		}
	case actionDrop:
		drawCenteredText(screen, "DROP", labelFace, cx, cy-14, white)
	}
}

func drawCell(screen *ebiten.Image, x, y, size, value int) {
	colour := pieceColors[value]
	ebitenutil.DrawRect(screen, float64(x+1), float64(y+1), float64(size-2), float64(size-2), colour)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return logicalWidth, logicalHeight
}
