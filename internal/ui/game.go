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
	matchcore "github.com/vibloteket/eit2/internal/match"
	"github.com/vibloteket/eit2/internal/version"
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
	actionAnti  = controls.UseAntidote
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
	padHeld     map[int]map[action]int
	fontSource  *text.GoTextFaceSource
	view        view
	players     []*core.Game
	match       *matchcore.Match
	paused      bool
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
		padHeld:     make(map[int]map[action]int),
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
	g.match = matchcore.New(len(g.Lobby.Slots))
	g.players = g.match.Players
	g.paused = false
	clear(g.heldActions)
	clear(g.padHeld)
	g.view = viewPlay
}

func (g *Game) restart() {
	g.start()
}

func (g *Game) backToLobby() {
	g.paused = false
	clear(g.heldActions)
	g.view = viewLobby
}

func (g *Game) updatePlay() {
	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.backToLobby()
		return
	}
	if len(g.players) == 0 {
		return
	}
	player := g.players[0]
	g.updateGamepads()
	soloGameOver := len(g.players) == 1 && player.GameOver
	matchOver := g.match != nil && g.match.Over
	for _, id := range g.pressedIDs {
		x, y := ebiten.TouchPosition(id)
		if g.handlePlayMenuPointer(x, y, soloGameOver || matchOver) {
			return
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if g.handlePlayMenuPointer(x, y, soloGameOver || matchOver) {
			return
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.paused = !g.paused
		clear(g.heldActions)
	}
	if g.paused || soloGameOver || matchOver {
		return
	}
	if keyboardPlayer := g.playerForDevice(lobby.DeviceKeyboard); keyboardPlayer != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			keyboardPlayer.MoveInput(-1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			keyboardPlayer.MoveInput(1)
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
			if ebiten.Tick()%2 == 0 {
				keyboardPlayer.StepDown()
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			keyboardPlayer.RotateInput(-1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			keyboardPlayer.RotateInput(1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			keyboardPlayer.HardDrop()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyE) {
			keyboardPlayer.UseAntidote()
		}
	}
	touchPlayer := g.playerForDevice(lobby.DeviceTouch)
	if touchPlayer == nil {
		touchPlayer = player
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
			apply(touchPlayer, action)
		}
		g.heldActions[action] = ticks + 1
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		for _, control := range touchButtons() {
			if control.Rect.contains(x, y) {
				apply(touchPlayer, control.Do)
			}
		}
	}
	g.match.Tick()
}

func (g *Game) playerForDevice(kind lobby.DeviceKind) *core.Game {
	for i, slot := range g.Lobby.Slots {
		if slot.Device.Kind == kind && i < len(g.players) {
			return g.players[i]
		}
	}
	return nil
}

func (g *Game) updateGamepads() {
	for playerIndex, slot := range g.Lobby.Slots {
		if slot.Device.Kind != lobby.DeviceGamepad || playerIndex >= len(g.players) {
			continue
		}
		id := ebiten.GamepadID(slot.Device.ID)
		active := map[action]bool{
			actionLeft:  ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftLeft),
			actionRight: ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftRight),
			actionDown:  ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftBottom),
		}
		held := g.padHeld[playerIndex]
		if held == nil {
			held = make(map[action]int)
			g.padHeld[playerIndex] = held
		}
		for action := range held {
			if !active[action] {
				delete(held, action)
			}
		}
		for action := range active {
			if !active[action] {
				continue
			}
			ticks := held[action]
			if controls.ShouldRepeat(action, ticks) {
				apply(g.players[playerIndex], action)
			}
			held[action] = ticks + 1
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightLeft) {
			g.players[playerIndex].RotateInput(-1)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) {
			g.players[playerIndex].RotateInput(1)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) {
			g.players[playerIndex].HardDrop()
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonFrontTopRight) {
			g.match.CycleTarget(playerIndex)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonFrontTopLeft) {
			g.players[playerIndex].UseAntidote()
		}
	}
}

func (g *Game) handlePlayMenuPointer(x, y int, gameOver bool) bool {
	if g.paused || gameOver {
		if g.paused && resumeButton().contains(x, y) {
			g.paused = false
			clear(g.heldActions)
			return true
		}
		restart, back := menuButtons(gameOver)
		if restart.contains(x, y) {
			g.restart()
			return true
		}
		if back.contains(x, y) {
			g.backToLobby()
			return true
		}
		return true // The modal menu consumes all pointer input.
	}
	if pauseButton().contains(x, y) {
		g.paused = true
		clear(g.heldActions)
		return true
	}
	return false
}

func apply(game *core.Game, action action) {
	switch action {
	case actionLeft:
		game.MoveInput(-1)
	case actionRight:
		game.MoveInput(1)
	case actionDown:
		game.StepDown()
	case actionCCW:
		game.RotateInput(-1)
	case actionCW:
		game.RotateInput(1)
	case actionDrop:
		game.HardDrop()
	case actionAnti:
		game.UseAntidote()
	}
}

func startButton() imageRect  { return imageRect{X: 490, Y: 590, W: 300, H: 85} }
func pauseButton() imageRect  { return imageRect{X: 45, Y: 205, W: 160, H: 62} }
func resumeButton() imageRect { return imageRect{X: 375, Y: 340, W: 160, H: 72} }

func menuButtons(gameOver bool) (restart, back imageRect) {
	if gameOver {
		return imageRect{X: 455, Y: 340, W: 180, H: 72}, imageRect{X: 650, Y: 340, W: 180, H: 72}
	}
	return imageRect{X: 560, Y: 340, W: 160, H: 72}, imageRect{X: 745, Y: 340, W: 160, H: 72}
}

func touchButtons() []button {
	return []button{
		{Rect: imageRect{X: 45, Y: 420, W: 145, H: 120}, Label: "LEFT", Do: actionLeft},
		{Rect: imageRect{X: 205, Y: 420, W: 145, H: 120}, Label: "RIGHT", Do: actionRight},
		{Rect: imageRect{X: 125, Y: 570, W: 145, H: 120}, Label: "DOWN", Do: actionDown},
		{Rect: imageRect{X: 930, Y: 420, W: 145, H: 120}, Label: "CCW", Do: actionCCW},
		{Rect: imageRect{X: 1090, Y: 420, W: 145, H: 120}, Label: "CW", Do: actionCW},
		{Rect: imageRect{X: 970, Y: 570, W: 150, H: 120}, Label: "DROP", Do: actionDrop},
		{Rect: imageRect{X: 1135, Y: 570, W: 100, H: 120}, Label: "ANTI", Do: actionAnti},
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
	drawText(screen, "v"+version.Value, g.face(20), 1135, 35, muted)
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
	if len(g.players) > 1 {
		g.drawCouch(screen)
		g.drawMatchOverlay(screen)
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
				drawSpecial(screen, boardX+x*cell, boardY+y*cell, cell, game.Specials[y][x], g.face(float64(cell-5)))
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
	if game.Blind {
		drawText(screen, "?", g.face(54), 1040, 76, white)
	} else {
		next := core.Piece{Kind: game.NextKind}
		for _, point := range core.PieceCells(next) {
			drawCell(screen, 1020+point.X*24, 82+point.Y*24, 24, game.NextKind+1)
		}
	}
	drawText(screen, "TARGET", g.face(20), 1020, 150, muted)
	drawText(screen, "SELF", g.face(22), 1020, 178, white)
	drawText(screen, "STORED", g.face(20), 1020, 215, muted)
	stored := "—"
	if game.Antidotes > 0 {
		stored = fmt.Sprintf("Antidote × %d", game.Antidotes)
	}
	drawText(screen, stored, g.face(22), 1020, 243, white)
	drawText(screen, "EFFECTS", g.face(20), 1020, 280, muted)
	effects := effectLabel(game)
	drawText(screen, effects, g.face(22), 1020, 307, white)

	pause := pauseButton()
	ebitenutil.DrawRect(screen, float64(pause.X), float64(pause.Y), float64(pause.W), float64(pause.H), panel)
	ebitenutil.DrawRect(screen, float64(pause.X), float64(pause.Y), float64(pause.W), 4, accent)
	drawCenteredText(screen, "PAUSE", g.face(20), float64(pause.X+pause.W/2), float64(pause.Y+18), white)

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

	g.drawMatchOverlay(screen)
}

func (g *Game) drawCouch(screen *ebiten.Image) {
	count := len(g.players)
	const gap = 12
	areaWidth := (logicalWidth - 40 - gap*(count-1)) / count
	cell := (logicalHeight - 125) / core.BoardHeight
	if areaWidth*2/3/core.BoardWidth < cell {
		cell = areaWidth * 2 / 3 / core.BoardWidth
	}
	for i, game := range g.players {
		x := 20 + i*(areaWidth+gap)
		boardW, boardH := core.BoardWidth*cell, core.BoardHeight*cell
		boardX, boardY := x+8, 72
		drawText(screen, fmt.Sprintf("P%d", i+1), g.face(22), float64(x+8), 20, white)
		drawText(screen, fmt.Sprintf("%d pts · L%d", game.Score, game.Lines/5), g.face(16), float64(x+48), 25, muted)
		ebitenutil.DrawRect(screen, float64(boardX-3), float64(boardY-3), float64(boardW+6), float64(boardH+6), muted)
		ebitenutil.DrawRect(screen, float64(boardX), float64(boardY), float64(boardW), float64(boardH), panel)
		for y, row := range game.Board {
			for bx, value := range row {
				if value != 0 {
					drawCell(screen, boardX+bx*cell, boardY+y*cell, cell, value)
					drawSpecial(screen, boardX+bx*cell, boardY+y*cell, cell, game.Specials[y][bx], g.face(float64(cell-4)))
				}
			}
		}
		for _, point := range game.Cells(game.Active) {
			if point.Y >= 0 {
				drawCell(screen, boardX+point.X*cell, boardY+point.Y*cell, cell, game.Active.Kind+1)
			}
		}
		hudX := boardX + boardW + 12
		drawText(screen, "NEXT", g.face(14), float64(hudX), 80, muted)
		if game.Blind {
			drawText(screen, "?", g.face(28), float64(hudX+10), 105, white)
		} else {
			for _, point := range core.PieceCells(core.Piece{Kind: game.NextKind}) {
				drawCell(screen, hudX+point.X*14, 102+point.Y*14, 14, game.NextKind+1)
			}
		}
		drawText(screen, "TARGET", g.face(14), float64(hudX), 185, muted)
		target := "—"
		if targetIndex := g.match.Target(i); targetIndex >= 0 {
			target = fmt.Sprintf("P%d", targetIndex+1)
		}
		drawText(screen, target, g.face(20), float64(hudX), 210, white)
		drawText(screen, "EFFECTS", g.face(13), float64(hudX), 245, muted)
		drawText(screen, effectLabel(game), g.face(13), float64(hudX), 264, white)
		drawText(screen, "STORED", g.face(14), float64(hudX), 300, muted)
		stored := "—"
		if game.Antidotes > 0 {
			stored = fmt.Sprintf("A × %d", game.Antidotes)
		}
		drawText(screen, stored, g.face(17), float64(hudX), 323, white)
		if game.GameOver {
			drawCenteredText(screen, "OUT", g.face(24), float64(boardX+boardW/2), float64(boardY+boardH/2), white)
		}
		for attacker := range g.players {
			if g.match.Target(attacker) == i {
				ebitenutil.DrawRect(screen, float64(boardX-6), float64(boardY-6), float64(boardW+12), 4, pieceColors[attacker%7+1])
			}
		}
	}
	pause := pauseButton()
	ebitenutil.DrawRect(screen, float64(pause.X), float64(pause.Y), float64(pause.W), float64(pause.H), panel)
	drawCenteredText(screen, "PAUSE", g.face(18), float64(pause.X+pause.W/2), float64(pause.Y+18), white)
}

func (g *Game) drawMatchOverlay(screen *ebiten.Image) {
	gameOver := len(g.players) == 1 && g.players[0].GameOver
	matchOver := g.match != nil && g.match.Over
	if !g.paused && !gameOver && !matchOver {
		return
	}
	ebitenutil.DrawRect(screen, 385, 235, 510, 230, color.RGBA{R: 8, G: 12, B: 20, A: 238})
	title := "PAUSED"
	if gameOver {
		title = "GAME OVER"
	} else if matchOver && g.match.Winner >= 0 {
		title = fmt.Sprintf("PLAYER %d WINS", g.match.Winner+1)
	}
	drawCenteredText(screen, title, g.face(48), logicalWidth/2, 260, white)
	restart, back := menuButtons(gameOver || matchOver)
	if g.paused {
		resume := resumeButton()
		ebitenutil.DrawRect(screen, float64(resume.X), float64(resume.Y), float64(resume.W), float64(resume.H), accent)
		drawCenteredText(screen, "RESUME", g.face(19), float64(resume.X+resume.W/2), float64(resume.Y+22), background)
	}
	ebitenutil.DrawRect(screen, float64(restart.X), float64(restart.Y), float64(restart.W), float64(restart.H), panel)
	ebitenutil.DrawRect(screen, float64(back.X), float64(back.Y), float64(back.W), float64(back.H), panel)
	drawCenteredText(screen, "RESTART", g.face(19), float64(restart.X+restart.W/2), float64(restart.Y+22), white)
	drawCenteredText(screen, "LOBBY", g.face(19), float64(back.X+back.W/2), float64(back.Y+22), white)
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
	case actionAnti:
		drawCenteredText(screen, "ANTI", labelFace, cx, cy-14, white)
	}
}

func effectLabel(game *core.Game) string {
	label := ""
	appendEffect := func(effect string) {
		if label != "" {
			label += " · "
		}
		label += effect
	}
	if game.Blind {
		appendEffect("Blind")
	}
	if game.Inverse {
		appendEffect("Inverse")
	}
	if game.FasterStacks > 0 {
		appendEffect("Faster")
	}
	if game.SlowerBonus > 0 {
		appendEffect("Slower")
	}
	if label == "" {
		return "None"
	}
	return label
}

func drawSpecial(screen *ebiten.Image, x, y, size int, special core.Special, face *text.GoTextFace) {
	label := ""
	switch special {
	case core.SpecialAntidote:
		label = "A"
	case core.SpecialClear:
		label = "C"
	case core.SpecialBlind:
		label = "B"
	case core.SpecialInverse:
		label = "I"
	case core.SpecialFaster:
		label = "F"
	case core.SpecialSlower:
		label = "S"
	case core.SpecialBridge:
		label = "+2"
	case core.SpecialQuestion:
		label = "?"
	}
	if label != "" {
		drawCenteredText(screen, label, face, float64(x+size/2), float64(y+1), background)
	}
}

func drawCell(screen *ebiten.Image, x, y, size, value int) {
	colour := pieceColors[value]
	ebitenutil.DrawRect(screen, float64(x+1), float64(y+1), float64(size-2), float64(size-2), colour)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return logicalWidth, logicalHeight
}
