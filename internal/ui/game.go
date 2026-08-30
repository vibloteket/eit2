package ui

import (
	"bytes"
	"fmt"
	"image/color"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/vibloteket/eit2/internal/controls"
	core "github.com/vibloteket/eit2/internal/game"
	"github.com/vibloteket/eit2/internal/lobby"
	matchcore "github.com/vibloteket/eit2/internal/match"
	"github.com/vibloteket/eit2/internal/sound"
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
	iceBlock    = color.RGBA{R: 205, G: 231, B: 238, A: 238}
	pieceColors = [...]color.RGBA{
		{}, {R: 97, G: 218, B: 251, A: 255}, {R: 255, G: 209, B: 102, A: 255},
		{R: 139, G: 124, B: 246, A: 255}, {R: 255, G: 159, B: 67, A: 255},
		{R: 71, G: 120, B: 245, A: 255}, {R: 76, G: 230, B: 166, A: 255},
		{R: 255, G: 107, B: 107, A: 255}, {R: 142, G: 151, B: 164, A: 255},
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

type keyboardLayout struct {
	ID                  int
	Name                string
	Left, Right, Down   ebiten.Key
	RotateCCW, RotateCW ebiten.Key
	RotateCWAlt, Drop   ebiten.Key
	Antidote, Target    ebiten.Key
}

const noKey ebiten.Key = -1

var keyboardLayouts = []keyboardLayout{
	{controls.KeyboardLayouts[0].ID, controls.KeyboardLayouts[0].Name, ebiten.KeyA, ebiten.KeyD, ebiten.KeyS, ebiten.KeyQ, ebiten.KeyW, noKey, ebiten.KeyShiftLeft, ebiten.KeyE, ebiten.KeyTab},
	{controls.KeyboardLayouts[1].ID, controls.KeyboardLayouts[1].Name, ebiten.KeyArrowLeft, ebiten.KeyArrowRight, ebiten.KeyArrowDown, ebiten.KeyComma, ebiten.KeyArrowUp, ebiten.KeyPeriod, ebiten.KeyShiftRight, ebiten.KeySlash, ebiten.KeyEnter},
	{controls.KeyboardLayouts[2].ID, controls.KeyboardLayouts[2].Name, ebiten.KeyJ, ebiten.KeyL, ebiten.KeyK, ebiten.KeyU, ebiten.KeyI, noKey, ebiten.KeySpace, ebiten.KeyO, ebiten.KeyP},
}

type imageRect struct{ X, Y, W, H int }

func (r imageRect) contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

type Game struct {
	Lobby               lobby.Lobby
	gamepadIDs          []ebiten.GamepadID
	touchIDs            []ebiten.TouchID
	pressedIDs          []ebiten.TouchID
	heldActions         map[action]int
	padHeld             map[int]map[action]int
	fontSource          *text.GoTextFaceSource
	view                view
	players             []*core.Game
	match               *matchcore.Match
	paused              bool
	debugEnabled        bool
	debugOpen           bool
	debugPlayer         int
	controllerDebugOpen bool
	disconnectedPlayer  int
	lobbyFocus          int
	overlayFocus        int
	debugFocus          int
	stickX              map[int]int
	stickY              map[int]int
	sound               *sound.Manager
	soundError          error
	touchDevice         lobby.Device
}

func NewGame() *Game {
	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(fmt.Sprintf("load embedded font: %v", err))
	}
	soundManager, soundErr := sound.New()
	return &Game{
		fontSource:         fontSource,
		sound:              soundManager,
		soundError:         soundErr,
		heldActions:        make(map[action]int),
		padHeld:            make(map[int]map[action]int),
		stickX:             make(map[int]int),
		stickY:             make(map[int]int),
		disconnectedPlayer: -1,
		touchDevice:        lobby.Device{Kind: lobby.DeviceTouch, Name: "Touch controls"},
	}
}

func (g *Game) Update() error {
	g.pressedIDs = inpututil.AppendJustPressedTouchIDs(g.pressedIDs[:0])
	if g.sound != nil {
		g.sound.Update()
	}
	if g.view == viewPlay {
		g.updatePlay()
		g.playAudioEvents()
		return nil
	}
	return g.updateLobby()
}

func isWeb() bool { return runtime.GOOS == "js" }

func leaveWebFullscreen() {
	if isWeb() && ebiten.IsFullscreen() {
		ebiten.SetFullscreen(false)
	}
}

func (g *Game) updateLobby() error {
	g.gamepadIDs = ebiten.AppendGamepadIDs(g.gamepadIDs[:0])
	navigate := func(direction controls.MenuDirection) {
		g.lobbyFocus = controls.NavigateLobby(g.lobbyFocus, direction, !isWeb())
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		navigate(controls.MenuLeft)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		navigate(controls.MenuRight)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		navigate(controls.MenuUp)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		navigate(controls.MenuDown)
	}
	activate := inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	for _, id := range g.gamepadIDs {
		if g.controllerDebugOpen && inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) {
			g.controllerDebugOpen = false
			continue
		}
		device := lobby.Device{Kind: lobby.DeviceGamepad, ID: int(id), Name: ebiten.GamepadName(id)}
		joinedPlayer := g.Lobby.PlayerForDevice(device)
		joined := joinedPlayer >= 0
		xDirection := controls.AxisDirection(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal), g.stickX[int(id)])
		yDirection := controls.AxisDirection(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical), g.stickY[int(id)])
		if xDirection != g.stickX[int(id)] && xDirection != 0 {
			if xDirection < 0 {
				navigate(controls.MenuLeft)
			} else {
				navigate(controls.MenuRight)
			}
		}
		if yDirection != g.stickY[int(id)] && yDirection != 0 {
			if yDirection < 0 {
				navigate(controls.MenuUp)
			} else {
				navigate(controls.MenuDown)
			}
		}
		g.stickX[int(id)], g.stickY[int(id)] = xDirection, yDirection
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftLeft) {
			navigate(controls.MenuLeft)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftRight) {
			navigate(controls.MenuRight)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftTop) {
			navigate(controls.MenuUp)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftBottom) {
			navigate(controls.MenuDown)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) {
			if joined {
				activate = true
			} else {
				g.Lobby.Join(device)
			}
		}
		if joined && inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) {
			g.Lobby.Leave(device)
			if g.lobbyFocus == 0 && !g.Lobby.CanStart() {
				g.lobbyFocus = 1
			}
		}
		if g.Lobby.CanStart() && inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonCenterRight) {
			g.start()
			return nil
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonCenterLeft) {
			if isWeb() {
				leaveWebFullscreen()
			} else {
				return ebiten.Termination
			}
		}
	}
	if activate {
		if g.activateLobbyMenu(g.lobbyFocus) {
			return ebiten.Termination
		}
	}
	if g.Lobby.CanStart() && g.lobbyFocus == 1 {
		g.lobbyFocus = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if isWeb() {
			leaveWebFullscreen()
		} else {
			return ebiten.Termination
		}
	}
	joinKeys := []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3}
	for index, key := range joinKeys {
		if inpututil.IsKeyJustPressed(key) {
			layout := keyboardLayouts[index]
			g.Lobby.Join(lobby.Device{Kind: lobby.DeviceKeyboard, ID: layout.ID, Name: layout.Name})
		}
	}
	for _, id := range g.pressedIDs {
		x, y := ebiten.TouchPosition(id)
		if !isWeb() && exitButton().contains(x, y) {
			return ebiten.Termination
		} else if muteButton().contains(x, y) && g.sound != nil {
			g.sound.ToggleMute()
		} else if musicButton().contains(x, y) && g.sound != nil {
			g.sound.ToggleMusic()
		} else if controllerDebugButton().contains(x, y) {
			g.controllerDebugOpen = !g.controllerDebugOpen
		} else if g.controllerDebugOpen {
			g.controllerDebugOpen = false
		} else if debugLobbyButton().contains(x, y) {
			g.debugEnabled = !g.debugEnabled
		} else if g.Lobby.CanStart() && startButton().contains(x, y) {
			g.start()
		} else {
			g.Lobby.Join(g.touchDevice)
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if !isWeb() && exitButton().contains(x, y) {
			return ebiten.Termination
		} else if muteButton().contains(x, y) && g.sound != nil {
			g.sound.ToggleMute()
		} else if musicButton().contains(x, y) && g.sound != nil {
			g.sound.ToggleMusic()
		} else if controllerDebugButton().contains(x, y) {
			g.controllerDebugOpen = !g.controllerDebugOpen
		} else if g.controllerDebugOpen {
			g.controllerDebugOpen = false
		} else if debugLobbyButton().contains(x, y) {
			g.debugEnabled = !g.debugEnabled
		} else if g.Lobby.CanStart() && startButton().contains(x, y) {
			g.start()
		}
	}
	return nil
}

func (g *Game) activateLobbyMenu(index int) bool {
	switch index {
	case 0:
		if g.Lobby.CanStart() {
			g.start()
		}
	case 1:
		if g.sound != nil {
			g.sound.ToggleMute()
		}
	case 2:
		if g.sound != nil {
			g.sound.ToggleMusic()
		}
	case 3:
		g.controllerDebugOpen = !g.controllerDebugOpen
	case 4:
		g.debugEnabled = !g.debugEnabled
	case 5:
		return !isWeb()
	}
	return false
}

func (g *Game) start() {
	g.match = matchcore.NewSeeded(len(g.Lobby.Slots), uint64(time.Now().UnixNano()))
	g.players = g.match.Players
	g.paused = false
	g.debugOpen = false
	g.debugPlayer = 0
	g.disconnectedPlayer = -1
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
	g.updateControllerConnections()
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.paused = !g.paused
		g.overlayFocus = 0
		clear(g.heldActions)
		clear(g.padHeld)
		return
	}
	if len(g.players) == 0 {
		return
	}
	player := g.players[0]
	if g.debugOpen {
		g.updateDebugGamepads()
		for _, id := range g.pressedIDs {
			x, y := ebiten.TouchPosition(id)
			g.handleDebugPointer(x, y)
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			g.handleDebugPointer(x, y)
		}
		return
	}
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
	if g.disconnectedPlayer >= 0 {
		return
	}
	if g.paused || soloGameOver || matchOver {
		g.updateOverlayGamepads(soloGameOver || matchOver)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.paused = !g.paused
		g.overlayFocus = 0
		clear(g.heldActions)
	}
	if g.paused || soloGameOver || matchOver {
		return
	}
	g.updateKeyboards()
	g.updateSoloKeyboardAliases()
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

func (g *Game) updateControllerConnections() {
	connected := make(map[int]bool)
	for _, id := range ebiten.AppendGamepadIDs(g.gamepadIDs[:0]) {
		connected[int(id)] = true
		if g.disconnectedPlayer >= 0 && inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) {
			device := lobby.Device{Kind: lobby.DeviceGamepad, ID: int(id), Name: ebiten.GamepadName(id)}
			if g.Lobby.ReplaceDevice(g.disconnectedPlayer, device) {
				g.disconnectedPlayer = -1
			}
		}
	}
	if g.disconnectedPlayer >= 0 {
		return
	}
	for player, slot := range g.Lobby.Slots {
		if slot.Device.Kind == lobby.DeviceGamepad && !connected[slot.Device.ID] {
			g.disconnectedPlayer = player
			g.paused = true
			clear(g.heldActions)
			clear(g.padHeld)
			return
		}
	}
}

func (g *Game) playAudioEvents() {
	if g.sound == nil {
		return
	}
	mapping := map[core.AudioEvent]sound.Effect{
		core.AudioLock: sound.Lock, core.AudioLine: sound.Line,
		core.AudioFourLine: sound.FourLine, core.AudioPickup: sound.Pickup,
		core.AudioAttack: sound.Attack, core.AudioGameOver: sound.GameOver,
	}
	for _, player := range g.players {
		for _, event := range player.ConsumeAudioEvents() {
			if effect, ok := mapping[event]; ok {
				g.sound.Play(effect)
			}
		}
	}
}

func (g *Game) updateSoloKeyboardAliases() {
	if len(g.players) != 1 || len(g.Lobby.Slots) != 1 || g.Lobby.Slots[0].Device.Kind != lobby.DeviceKeyboard || g.Lobby.Slots[0].Device.ID != 1 {
		return
	}
	player := g.players[0]
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		player.MoveInput(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		player.MoveInput(1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) && ebiten.Tick()%2 == 0 {
		player.StepDown()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyComma) {
		player.RotateInput(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
		player.RotateInput(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyShiftRight) {
		player.HardDrop()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySlash) {
		player.UseAntidote()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.match.CycleTarget(0)
	}
}

func (g *Game) updateKeyboards() {
	for playerIndex, slot := range g.Lobby.Slots {
		if slot.Device.Kind != lobby.DeviceKeyboard || playerIndex >= len(g.players) {
			continue
		}
		var layout *keyboardLayout
		for i := range keyboardLayouts {
			if keyboardLayouts[i].ID == slot.Device.ID {
				layout = &keyboardLayouts[i]
				break
			}
		}
		if layout == nil {
			continue
		}
		player := g.players[playerIndex]
		if inpututil.IsKeyJustPressed(layout.Left) {
			player.MoveInput(-1)
		}
		if inpututil.IsKeyJustPressed(layout.Right) {
			player.MoveInput(1)
		}
		if ebiten.IsKeyPressed(layout.Down) && ebiten.Tick()%2 == 0 {
			player.StepDown()
		}
		if inpututil.IsKeyJustPressed(layout.RotateCCW) {
			player.RotateInput(-1)
		}
		if inpututil.IsKeyJustPressed(layout.RotateCW) || (layout.RotateCWAlt != noKey && inpututil.IsKeyJustPressed(layout.RotateCWAlt)) {
			player.RotateInput(1)
		}
		if inpututil.IsKeyJustPressed(layout.Drop) {
			player.HardDrop()
		}
		if inpututil.IsKeyJustPressed(layout.Antidote) {
			player.UseAntidote()
		}
		if inpututil.IsKeyJustPressed(layout.Target) {
			g.match.CycleTarget(playerIndex)
		}
	}
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
		if g.disconnectedPlayer < 0 && inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonCenterRight) {
			g.paused = !g.paused
			g.overlayFocus = 0
			clear(g.heldActions)
			clear(g.padHeld)
		}
		if g.paused || (g.match != nil && g.match.Over) {
			continue
		}
		xAxis := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)
		yAxis := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical)
		active := map[action]bool{
			actionLeft:  ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftLeft) || xAxis < -0.55,
			actionRight: ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftRight) || xAxis > 0.55,
			actionDown:  ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftBottom) || yAxis > 0.55,
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

func (g *Game) updateDebugGamepads() {
	buttons := debugSpecialButtons()
	for _, id := range ebiten.AppendGamepadIDs(g.gamepadIDs[:0]) {
		xDirection := controls.AxisDirection(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal), g.stickX[int(id)])
		yDirection := controls.AxisDirection(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical), g.stickY[int(id)])
		if xDirection != g.stickX[int(id)] && xDirection != 0 {
			if xDirection < 0 {
				g.debugFocus = max(0, g.debugFocus-1)
			} else {
				g.debugFocus = min(len(buttons)-1, g.debugFocus+1)
			}
		}
		if yDirection != g.stickY[int(id)] && yDirection != 0 {
			if yDirection < 0 {
				g.debugFocus = max(0, g.debugFocus-4)
			} else {
				g.debugFocus = min(len(buttons)-1, g.debugFocus+4)
			}
		}
		g.stickX[int(id)], g.stickY[int(id)] = xDirection, yDirection
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) {
			g.debugOpen = false
			return
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonFrontTopLeft) && g.debugPlayer > 0 {
			g.debugPlayer--
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonFrontTopRight) && g.debugPlayer+1 < len(g.players) {
			g.debugPlayer++
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftLeft) {
			g.debugFocus = max(0, g.debugFocus-1)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftRight) {
			g.debugFocus = min(len(buttons)-1, g.debugFocus+1)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftTop) {
			g.debugFocus = max(0, g.debugFocus-4)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftBottom) {
			g.debugFocus = min(len(buttons)-1, g.debugFocus+4)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) && len(buttons) > 0 {
			g.match.DebugCollect(g.debugPlayer, buttons[g.debugFocus].Special)
			g.debugOpen = false
			return
		}
	}
}

func (g *Game) updateOverlayGamepads(gameOver bool) {
	buttonCount := 3
	if gameOver {
		buttonCount = 2
	}
	for _, id := range ebiten.AppendGamepadIDs(g.gamepadIDs[:0]) {
		xDirection := controls.AxisDirection(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal), g.stickX[int(id)])
		if xDirection != g.stickX[int(id)] && xDirection != 0 {
			g.overlayFocus = (g.overlayFocus + xDirection + buttonCount) % buttonCount
		}
		g.stickX[int(id)] = xDirection
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftLeft) {
			g.overlayFocus = (g.overlayFocus - 1 + buttonCount) % buttonCount
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftRight) {
			g.overlayFocus = (g.overlayFocus + 1) % buttonCount
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) {
			if g.paused {
				g.paused = false
			} else {
				g.backToLobby()
			}
			return
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) {
			if !gameOver && g.overlayFocus == 0 {
				g.paused = false
				return
			}
			adjusted := g.overlayFocus
			if !gameOver {
				adjusted--
			}
			if adjusted == 0 {
				g.restart()
			} else {
				g.backToLobby()
			}
			return
		}
	}
}

func (g *Game) handleDebugPointer(x, y int) {
	if debugCloseButton().contains(x, y) {
		g.debugOpen = false
		return
	}
	if debugPrevPlayer().contains(x, y) && g.debugPlayer > 0 {
		g.debugPlayer--
		return
	}
	if debugNextPlayer().contains(x, y) && g.debugPlayer+1 < len(g.players) {
		g.debugPlayer++
		return
	}
	for _, item := range debugSpecialButtons() {
		if item.Rect.contains(x, y) {
			g.match.DebugCollect(g.debugPlayer, item.Special)
			g.debugOpen = false
			return
		}
	}
	for _, item := range debugSoundButtons() {
		if item.Rect.contains(x, y) && g.sound != nil {
			g.sound.Play(item.Effect)
			return
		}
	}
}

func (g *Game) handlePlayMenuPointer(x, y int, gameOver bool) bool {
	if g.disconnectedPlayer >= 0 {
		return true
	}
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
	if g.debugEnabled && debugPlayButton().contains(x, y) {
		g.debugOpen = true
		g.debugFocus = 0
		clear(g.heldActions)
		return true
	}
	if len(g.players) == 1 && touchMenuButton().contains(x, y) {
		g.paused = true
		g.overlayFocus = 0
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

func startButton() imageRect           { return imageRect{X: 490, Y: 550, W: 300, H: 64} }
func controllerDebugButton() imageRect { return imageRect{X: 30, Y: 632, W: 240, H: 60} }
func muteButton() imageRect            { return imageRect{X: 285, Y: 632, W: 135, H: 60} }
func musicButton() imageRect           { return imageRect{X: 435, Y: 632, W: 155, H: 60} }
func debugLobbyButton() imageRect      { return imageRect{X: 605, Y: 632, W: 190, H: 60} }
func exitButton() imageRect            { return imageRect{X: 810, Y: 632, W: 130, H: 60} }

func lobbyMenuButtons() []imageRect {
	buttons := []imageRect{startButton(), muteButton(), musicButton(), controllerDebugButton(), debugLobbyButton()}
	if !isWeb() {
		buttons = append(buttons, exitButton())
	}
	return buttons
}
func debugPlayButton() imageRect { return imageRect{X: 45, Y: 280, W: 160, H: 62} }
func touchMenuButton() imageRect { return imageRect{X: 45, Y: 205, W: 160, H: 62} }
func resumeButton() imageRect    { return imageRect{X: 375, Y: 340, W: 160, H: 72} }

func debugCloseButton() imageRect { return imageRect{X: 1035, Y: 110, W: 150, H: 60} }
func debugPrevPlayer() imageRect  { return imageRect{X: 150, Y: 110, W: 90, H: 60} }
func debugNextPlayer() imageRect  { return imageRect{X: 390, Y: 110, W: 90, H: 60} }

func debugSoundButtons() []struct {
	Rect   imageRect
	Label  string
	Effect sound.Effect
} {
	return []struct {
		Rect   imageRect
		Label  string
		Effect sound.Effect
	}{
		{imageRect{105, 590, 160, 48}, "Lock", sound.Lock},
		{imageRect{275, 590, 160, 48}, "Line", sound.Line},
		{imageRect{445, 590, 160, 48}, "Four-line", sound.FourLine},
		{imageRect{615, 590, 160, 48}, "Pickup", sound.Pickup},
		{imageRect{785, 590, 160, 48}, "Attack", sound.Attack},
		{imageRect{955, 590, 160, 48}, "Game over", sound.GameOver},
	}
}

func debugSpecialButtons() []struct {
	Rect    imageRect
	Special core.Special
} {
	buttons := make([]struct {
		Rect    imageRect
		Special core.Special
	}, 0, len(core.AllSpecials))
	const columns, width, height, gapX, gapY = 4, 255, 52, 18, 8
	for i, special := range core.AllSpecials {
		row, column := i/columns, i%columns
		buttons = append(buttons, struct {
			Rect    imageRect
			Special core.Special
		}{Rect: imageRect{X: 105 + column*(width+gapX), Y: 190 + row*(height+gapY), W: width, H: height}, Special: special})
	}
	return buttons
}

func menuButtons(gameOver bool) (restart, back imageRect) {
	if gameOver {
		return imageRect{X: 455, Y: 340, W: 180, H: 72}, imageRect{X: 650, Y: 340, W: 180, H: 72}
	}
	return imageRect{X: 560, Y: 340, W: 160, H: 72}, imageRect{X: 745, Y: 340, W: 160, H: 72}
}

func touchButtons() []button {
	return []button{
		{Rect: imageRect{X: 15, Y: 420, W: 175, H: 120}, Label: "LEFT", Do: actionLeft},
		{Rect: imageRect{X: 210, Y: 420, W: 175, H: 120}, Label: "RIGHT", Do: actionRight},
		{Rect: imageRect{X: 125, Y: 570, W: 145, H: 120}, Label: "DOWN", Do: actionDown},
		{Rect: imageRect{X: 895, Y: 420, W: 175, H: 120}, Label: "CCW", Do: actionCCW},
		{Rect: imageRect{X: 1090, Y: 420, W: 175, H: 120}, Label: "CW", Do: actionCW},
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
	drawText(screen, "Touch / gamepad A / keys 1, 2, 3 to join · Enter starts", g.face(24), 42, 92, white)
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
	if !isWeb() {
		exit := exitButton()
		ebitenutil.DrawRect(screen, float64(exit.X), float64(exit.Y), float64(exit.W), float64(exit.H), panel)
		drawCenteredText(screen, "EXIT", g.face(17), float64(exit.X+exit.W/2), float64(exit.Y+16), white)
	}
	controllers := controllerDebugButton()
	ebitenutil.DrawRect(screen, float64(controllers.X), float64(controllers.Y), float64(controllers.W), float64(controllers.H), panel)
	drawCenteredText(screen, "CONTROLLER DEBUG", g.face(17), float64(controllers.X+controllers.W/2), float64(controllers.Y+20), white)
	mute := muteButton()
	ebitenutil.DrawRect(screen, float64(mute.X), float64(mute.Y), float64(mute.W), float64(mute.H), panel)
	muteLabel := "SOUND ON"
	if g.sound == nil || g.sound.Muted() {
		muteLabel = "MUTED"
	} else if !g.sound.Ready() {
		muteLabel = "AUDIO WAIT"
	}
	drawCenteredText(screen, muteLabel, g.face(16), float64(mute.X+mute.W/2), float64(mute.Y+20), white)
	music := musicButton()
	ebitenutil.DrawRect(screen, float64(music.X), float64(music.Y), float64(music.W), float64(music.H), panel)
	musicLabel := "MUSIC ON"
	if g.sound == nil || !g.sound.MusicEnabled() {
		musicLabel = "MUSIC OFF"
	}
	drawCenteredText(screen, musicLabel, g.face(16), float64(music.X+music.W/2), float64(music.Y+20), white)
	debug := debugLobbyButton()
	debugFill := panel
	debugLabel := "DEBUG MODE: OFF"
	if g.debugEnabled {
		debugFill = color.RGBA{R: 35, G: 73, B: 76, A: 255}
		debugLabel = "DEBUG MODE: ON"
	}
	ebitenutil.DrawRect(screen, float64(debug.X), float64(debug.Y), float64(debug.W), float64(debug.H), debugFill)
	drawCenteredText(screen, debugLabel, g.face(17), float64(debug.X+debug.W/2), float64(debug.Y+20), white)
	r := startButton()
	startFill, startText := accent, background
	if !g.Lobby.CanStart() {
		startFill, startText = panel, muted
		drawCenteredText(screen, "JOIN WITH TOUCH, 1 / 2 / 3 OR GAMEPAD A", g.face(16), logicalWidth/2, float64(r.Y-27), muted)
	}
	ebitenutil.DrawRect(screen, float64(r.X), float64(r.Y), float64(r.W), float64(r.H), startFill)
	drawCenteredText(screen, "START", g.face(26), float64(r.X+r.W/2), float64(r.Y+17), startText)
	buttons := lobbyMenuButtons()
	if g.lobbyFocus >= 0 && g.lobbyFocus < len(buttons) {
		r := buttons[g.lobbyFocus]
		ebitenutil.DrawRect(screen, float64(r.X-4), float64(r.Y-4), float64(r.W+8), 4, white)
		ebitenutil.DrawRect(screen, float64(r.X-4), float64(r.Y+r.H), float64(r.W+8), 4, white)
		ebitenutil.DrawRect(screen, float64(r.X-4), float64(r.Y-4), 4, float64(r.H+8), white)
		ebitenutil.DrawRect(screen, float64(r.X+r.W), float64(r.Y-4), 4, float64(r.H+8), white)
	}
	g.drawControllerDebug(screen)
}

func standardButtonName(button ebiten.StandardGamepadButton) string {
	names := map[ebiten.StandardGamepadButton]string{
		ebiten.StandardGamepadButtonRightBottom:   "A",
		ebiten.StandardGamepadButtonRightRight:    "B",
		ebiten.StandardGamepadButtonRightLeft:     "X",
		ebiten.StandardGamepadButtonRightTop:      "Y",
		ebiten.StandardGamepadButtonFrontTopLeft:  "LB",
		ebiten.StandardGamepadButtonFrontTopRight: "RB",
		ebiten.StandardGamepadButtonCenterLeft:    "BACK",
		ebiten.StandardGamepadButtonCenterRight:   "START",
		ebiten.StandardGamepadButtonLeftTop:       "UP",
		ebiten.StandardGamepadButtonLeftBottom:    "DOWN",
		ebiten.StandardGamepadButtonLeftLeft:      "LEFT",
		ebiten.StandardGamepadButtonLeftRight:     "RIGHT",
	}
	return names[button]
}

func pressedStandardButtons(id ebiten.GamepadID) string {
	pressed := ""
	for button := ebiten.StandardGamepadButton(0); button <= ebiten.StandardGamepadButtonMax; button++ {
		if !ebiten.IsStandardGamepadButtonPressed(id, button) {
			continue
		}
		name := standardButtonName(button)
		if name == "" {
			name = fmt.Sprintf("B%d", button)
		}
		if pressed != "" {
			pressed += " "
		}
		pressed += name
	}
	if pressed == "" {
		return "—"
	}
	return pressed
}

func (g *Game) drawControllerDebug(screen *ebiten.Image) {
	if !g.controllerDebugOpen {
		return
	}
	ebitenutil.DrawRect(screen, 55, 65, 1170, 570, color.RGBA{R: 8, G: 12, B: 20, A: 250})
	ebitenutil.DrawRect(screen, 55, 65, 1170, 5, accent)
	drawText(screen, "CONTROLLER DEBUG", g.face(32), 90, 92, accent)
	drawText(screen, "Tap anywhere or press B to close · press buttons to see live state", g.face(17), 90, 137, muted)
	ids := ebiten.AppendGamepadIDs(g.gamepadIDs[:0])
	if len(ids) == 0 {
		drawCenteredText(screen, "No gamepads detected", g.face(28), logicalWidth/2, 310, white)
		return
	}
	for index, id := range ids {
		y := 185 + index*100
		device := lobby.Device{Kind: lobby.DeviceGamepad, ID: int(id)}
		player := g.Lobby.PlayerForDevice(device)
		assignment := "not joined"
		if player >= 0 {
			assignment = fmt.Sprintf("PLAYER %d", player+1)
		}
		mapping := "raw mapping"
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			mapping = "standard mapping"
		}
		name := ebiten.GamepadName(id)
		if name == "" {
			name = "Unnamed controller"
		}
		drawText(screen, fmt.Sprintf("ID %d · %s", id, name), g.face(20), 90, float64(y), white)
		drawText(screen, assignment+" · "+mapping+fmt.Sprintf(" · axes %d", ebiten.GamepadAxisCount(id)), g.face(16), 90, float64(y+31), muted)
		drawText(screen, "Pressed: "+pressedStandardButtons(id), g.face(17), 90, float64(y+58), accent)
		sdlID := ebiten.GamepadSDLID(id)
		if sdlID != "" {
			drawText(screen, "SDL: "+sdlID, g.face(13), 650, float64(y+31), muted)
		}
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
		g.drawDebugPanel(screen)
		return
	}
	game := g.players[0]
	const cell = 27
	boardW, boardH := core.BoardWidth*cell, core.BoardHeight*cell
	boardX, boardY := (logicalWidth-boardW)/2, 40
	ebitenutil.DrawRect(screen, float64(boardX-5), float64(boardY-5), float64(boardW+10), float64(boardH+10), muted)
	drawBoardBackground(screen, boardX, boardY, boardW, boardH, game.BackgroundVariant)
	for y, row := range game.Board {
		for x, value := range row {
			if value != 0 {
				hasSpecial := game.Specials[y][x] != core.SpecialNone
				drawSettledCell(screen, boardX+x*cell, boardY+y*cell, cell, value, game.Mini && !hasSpecial, game.Trans)
				drawSpecial(screen, boardX+x*cell, boardY+y*cell, cell, game.Specials[y][x], g.face(float64(cell-5)))
			}
		}
	}
	if !game.Blink || game.BlinkVisible {
		for _, point := range game.Cells(game.Active) {
			if point.Y >= 0 {
				drawCell(screen, boardX+point.X*cell, boardY+point.Y*cell, cell, game.Active.Kind+1)
			}
		}
	}
	drawBlackout(screen, game, boardX, boardY, boardW, boardH, cell)
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

	menu := touchMenuButton()
	ebitenutil.DrawRect(screen, float64(menu.X), float64(menu.Y), float64(menu.W), float64(menu.H), panel)
	ebitenutil.DrawRect(screen, float64(menu.X), float64(menu.Y), float64(menu.W), 4, accent)
	drawCenteredText(screen, "MENU", g.face(20), float64(menu.X+menu.W/2), float64(menu.Y+18), white)
	if g.debugEnabled {
		debug := debugPlayButton()
		ebitenutil.DrawRect(screen, float64(debug.X), float64(debug.Y), float64(debug.W), float64(debug.H), panel)
		ebitenutil.DrawRect(screen, float64(debug.X), float64(debug.Y), float64(debug.W), 4, muted)
		drawCenteredText(screen, "DEBUG", g.face(19), float64(debug.X+debug.W/2), float64(debug.Y+18), white)
	}

	if game.LastEvent != "" {
		drawCenteredText(screen, game.LastEvent, g.face(22), logicalWidth/2, 645, accent)
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

	g.drawMatchOverlay(screen)
	g.drawDebugPanel(screen)
}

func drawBoardBackground(screen *ebiten.Image, x, y, width, height, variant int) {
	variants := [...]color.RGBA{
		panel,
		{R: 35, G: 50, B: 72, A: 255},
		{R: 56, G: 38, B: 66, A: 255},
		{R: 29, G: 63, B: 54, A: 255},
		{R: 71, G: 48, B: 32, A: 255},
		{R: 45, G: 44, B: 76, A: 255},
		{R: 69, G: 35, B: 48, A: 255},
	}
	if variant < 0 || variant >= len(variants) {
		variant = 0
	}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(width), float64(height), variants[variant])
}

func (g *Game) drawCouch(screen *ebiten.Image) {
	count := len(g.players)
	const gap = 18
	areaWidth := 300
	if count == 4 {
		areaWidth = 294
	}
	totalWidth := count*areaWidth + (count-1)*gap
	groupX := (logicalWidth - totalWidth) / 2
	cell := (logicalHeight - 125) / core.BoardHeight
	if areaWidth*2/3/core.BoardWidth < cell {
		cell = areaWidth * 2 / 3 / core.BoardWidth
	}
	for i, game := range g.players {
		x := groupX + i*(areaWidth+gap)
		boardW, boardH := core.BoardWidth*cell, core.BoardHeight*cell
		boardX, boardY := x+8, 72
		drawText(screen, fmt.Sprintf("P%d", i+1), g.face(22), float64(x+8), 20, white)
		drawText(screen, fmt.Sprintf("%d pts · L%d", game.Score, game.Lines/5), g.face(16), float64(x+48), 25, muted)
		ebitenutil.DrawRect(screen, float64(boardX-3), float64(boardY-3), float64(boardW+6), float64(boardH+6), muted)
		drawBoardBackground(screen, boardX, boardY, boardW, boardH, game.BackgroundVariant)
		for y, row := range game.Board {
			for bx, value := range row {
				if value != 0 {
					hasSpecial := game.Specials[y][bx] != core.SpecialNone
					drawSettledCell(screen, boardX+bx*cell, boardY+y*cell, cell, value, game.Mini && !hasSpecial, game.Trans)
					drawSpecial(screen, boardX+bx*cell, boardY+y*cell, cell, game.Specials[y][bx], g.face(float64(cell-4)))
				}
			}
		}
		if !game.Blink || game.BlinkVisible {
			for _, point := range game.Cells(game.Active) {
				if point.Y >= 0 {
					drawCell(screen, boardX+point.X*cell, boardY+point.Y*cell, cell, game.Active.Kind+1)
				}
			}
		}
		drawBlackout(screen, game, boardX, boardY, boardW, boardH, cell)
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
		statusCenter := float64(boardX + boardW/2)
		drawCenteredText(screen, fmt.Sprintf("P%d · %d pts · Level %d", i+1, game.Score, game.Lines/5), g.face(15), statusCenter, 625, white)
		if game.LastEvent != "" {
			event := fitText(game.LastEvent, g.face(13), float64(boardW+80))
			drawCenteredText(screen, event, g.face(13), statusCenter, 650, accent)
		}
		if game.GameOver {
			drawCenteredText(screen, "OUT", g.face(24), float64(boardX+boardW/2), float64(boardY+boardH/2), white)
		}
		for attacker := range g.players {
			if g.match.Target(attacker) == i {
				ebitenutil.DrawRect(screen, float64(boardX-6), float64(boardY-6), float64(boardW+12), 4, pieceColors[attacker%7+1])
			}
		}
	}
}

func (g *Game) drawDebugPanel(screen *ebiten.Image) {
	if !g.debugOpen || g.match == nil || len(g.players) == 0 {
		return
	}
	ebitenutil.DrawRect(screen, 70, 70, 1140, 590, color.RGBA{R: 8, G: 12, B: 20, A: 248})
	ebitenutil.DrawRect(screen, 70, 70, 1140, 5, accent)
	drawText(screen, "DEBUG MODE", g.face(34), 105, 98, accent)
	drawText(screen, "Source player", g.face(18), 105, 151, muted)
	prev, next := debugPrevPlayer(), debugNextPlayer()
	ebitenutil.DrawRect(screen, float64(prev.X), float64(prev.Y), float64(prev.W), float64(prev.H), panel)
	ebitenutil.DrawRect(screen, float64(next.X), float64(next.Y), float64(next.W), float64(next.H), panel)
	drawCenteredText(screen, "<", g.face(28), float64(prev.X+prev.W/2), float64(prev.Y+10), white)
	drawCenteredText(screen, ">", g.face(28), float64(next.X+next.W/2), float64(next.Y+10), white)
	drawCenteredText(screen, fmt.Sprintf("PLAYER %d", g.debugPlayer+1), g.face(24), 315, 122, white)
	close := debugCloseButton()
	ebitenutil.DrawRect(screen, float64(close.X), float64(close.Y), float64(close.W), float64(close.H), panel)
	drawCenteredText(screen, "CLOSE", g.face(18), float64(close.X+close.W/2), float64(close.Y+16), white)

	for index, item := range debugSpecialButtons() {
		ebitenutil.DrawRect(screen, float64(item.Rect.X), float64(item.Rect.Y), float64(item.Rect.W), float64(item.Rect.H), panel)
		border := muted
		if index == g.debugFocus {
			border = accent
		}
		ebitenutil.DrawRect(screen, float64(item.Rect.X), float64(item.Rect.Y), float64(item.Rect.W), 3, border)
		drawCenteredText(screen, item.Special.Name(), g.face(16), float64(item.Rect.X+item.Rect.W/2), float64(item.Rect.Y+17), white)
	}
	for _, item := range debugSoundButtons() {
		ebitenutil.DrawRect(screen, float64(item.Rect.X), float64(item.Rect.Y), float64(item.Rect.W), float64(item.Rect.H), panel)
		drawCenteredText(screen, item.Label, g.face(15), float64(item.Rect.X+item.Rect.W/2), float64(item.Rect.Y+13), white)
	}
	audioState := "Audio ready"
	if g.soundError != nil {
		audioState = "Audio error"
	} else if g.sound == nil || g.sound.Muted() {
		audioState = "Muted"
	} else if !g.sound.Ready() {
		audioState = "Audio waiting for user interaction"
	}
	musicState := "music stopped"
	if g.sound != nil && g.sound.MusicPlaying() {
		musicState = "music playing"
	} else if g.sound != nil && !g.sound.MusicEnabled() {
		musicState = "music disabled"
	}
	drawCenteredText(screen, audioState+" · "+musicState+" · sound test", g.face(15), logicalWidth/2, 565, muted)
}

func (g *Game) drawMatchOverlay(screen *ebiten.Image) {
	gameOver := len(g.players) == 1 && g.players[0].GameOver
	matchOver := g.match != nil && g.match.Over
	if !g.paused && !gameOver && !matchOver && g.disconnectedPlayer < 0 {
		return
	}
	ebitenutil.DrawRect(screen, 385, 235, 510, 230, color.RGBA{R: 8, G: 12, B: 20, A: 238})
	title := "PAUSED"
	if g.disconnectedPlayer >= 0 {
		title = fmt.Sprintf("PLAYER %d CONTROLLER DISCONNECTED", g.disconnectedPlayer+1)
	} else if gameOver {
		title = "GAME OVER"
	} else if matchOver && g.match.Winner >= 0 {
		title = fmt.Sprintf("PLAYER %d WINS", g.match.Winner+1)
	}
	fontSize := 48.0
	if g.disconnectedPlayer >= 0 {
		fontSize = 25
	}
	drawCenteredText(screen, title, g.face(fontSize), logicalWidth/2, 260, white)
	if g.disconnectedPlayer >= 0 {
		drawCenteredText(screen, "Connect a controller and press A", g.face(20), logicalWidth/2, 320, muted)
		return
	}
	restart, back := menuButtons(gameOver || matchOver)
	if g.paused {
		resume := resumeButton()
		fill, textColour := panel, white
		if g.overlayFocus == 0 {
			fill, textColour = accent, background
		}
		ebitenutil.DrawRect(screen, float64(resume.X), float64(resume.Y), float64(resume.W), float64(resume.H), fill)
		drawCenteredText(screen, "RESUME", g.face(19), float64(resume.X+resume.W/2), float64(resume.Y+22), textColour)
	}
	restartFocus, backFocus := 0, 1
	if g.paused {
		restartFocus, backFocus = 1, 2
	}
	restartFill, backFill := panel, panel
	if g.overlayFocus == restartFocus {
		restartFill = accent
	}
	if g.overlayFocus == backFocus {
		backFill = accent
	}
	ebitenutil.DrawRect(screen, float64(restart.X), float64(restart.Y), float64(restart.W), float64(restart.H), restartFill)
	ebitenutil.DrawRect(screen, float64(back.X), float64(back.Y), float64(back.W), float64(back.H), backFill)
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
	if game.PacketTicks > 0 {
		appendEffect("Packet")
	}
	if game.Mini {
		appendEffect("Mini")
	}
	if game.Blink {
		appendEffect("Blink")
	}
	if game.SZ {
		appendEffect("SZ")
	}
	if game.Trans {
		appendEffect("Ice")
	}
	if game.Blackout {
		appendEffect("Blackout")
	}
	if game.RumbleRounds > 0 {
		appendEffect("Rumble")
	}
	if game.BackgroundVariant > 0 {
		appendEffect("Background")
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
	case core.SpecialStair:
		label = "ST"
	case core.SpecialFill:
		label = "FL"
	case core.SpecialFlip:
		label = "FP"
	case core.SpecialSwitch:
		label = "SW"
	case core.SpecialPacket:
		label = "P"
	case core.SpecialRing:
		label = "R"
	case core.SpecialMini:
		label = "M"
	case core.SpecialBlink:
		label = "BK"
	case core.SpecialSZ:
		label = "SZ"
	case core.SpecialTrans:
		label = "IC"
	case core.SpecialCastle:
		label = "CA"
	case core.SpecialColor:
		label = "BO"
	case core.SpecialRumble:
		label = "RU"
	case core.SpecialBackground:
		label = "BG"
	}
	if label != "" {
		drawCenteredText(screen, label, face, float64(x+size/2), float64(y+1), background)
	}
}

func drawBlackout(screen *ebiten.Image, game *core.Game, boardX, boardY, boardW, boardH, cell int) {
	if !game.Blackout {
		return
	}
	mask := ebiten.NewImage(boardW, boardH)
	mask.Fill(color.RGBA{A: 235})
	cells := game.Cells(game.Active)
	var centerX, centerY float32
	visible := 0
	for _, point := range cells {
		if point.Y >= 0 {
			centerX += float32(point.X*cell + cell/2)
			centerY += float32(point.Y*cell + cell/2)
			visible++
		}
	}
	if visible > 0 {
		centerX /= float32(visible)
		centerY /= float32(visible)
		op := &ebiten.DrawImageOptions{Blend: ebiten.BlendDestinationOut}
		spotlight := ebiten.NewImage(boardW, boardH)
		vector.FillCircle(spotlight, centerX, centerY, float32(cell)*2.8, color.White, true)
		mask.DrawImage(spotlight, op)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(boardX), float64(boardY))
	screen.DrawImage(mask, op)
}

func drawSettledCell(screen *ebiten.Image, x, y, size, value int, mini, iced bool) {
	if mini {
		miniSize := max(4, size*2/5)
		x += (size - miniSize) / 2
		y += (size - miniSize) / 2
		size = miniSize
	}
	if iced {
		// Ice removes the piece colours and nearly merges adjacent settled
		// cells. The active piece and special labels are drawn separately and
		// remain readable.
		ebitenutil.DrawRect(screen, float64(x), float64(y), float64(size), float64(size), iceBlock)
		return
	}
	ebitenutil.DrawRect(screen, float64(x+1), float64(y+1), float64(size-2), float64(size-2), pieceColors[value])
}

func drawCell(screen *ebiten.Image, x, y, size, value int) {
	drawSettledCell(screen, x, y, size, value, false, false)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return logicalWidth, logicalHeight
}
