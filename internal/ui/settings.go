package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/vibloteket/eit2/internal/controls"
	"github.com/vibloteket/eit2/internal/sound"
)

const lobbySettingsIndex = 1

const (
	settingsSoundIndex = iota
	settingsMusicIndex
	settingsVolumeIndex
	settingsControllerDebugIndex
	settingsDebugModeIndex
	settingsAboutIndex
	settingsBackIndex
)

const settingsItemCount = 7

func settingsRow(index int) imageRect {
	return imageRect{X: 330, Y: 165 + index*64, W: 620, H: 54}
}

func (g *Game) openSettings() {
	g.view = viewSettings
	g.settingsFocus = settingsSoundIndex
	g.controllerDebugOpen = false
	if g.sound != nil {
		g.sound.Play(sound.MenuSelect)
	}
}

func (g *Game) closeSettings() {
	g.view = viewLobby
	g.lobbyFocus = lobbySettingsIndex
	g.controllerDebugOpen = false
}

func (g *Game) settingsSoundOn() bool {
	return g.sound != nil && !g.sound.Muted()
}

func (g *Game) settingsMusicOn() bool {
	return g.sound != nil && g.sound.MusicEnabled()
}

func (g *Game) settingsVolumePercent() int {
	if g.sound == nil {
		return 100
	}
	return g.sound.MusicVolumePercent()
}

func (g *Game) settingsValue(index int) string {
	switch index {
	case settingsSoundIndex:
		if g.settingsSoundOn() {
			return "ON"
		}
		return "OFF"
	case settingsMusicIndex:
		if g.settingsMusicOn() {
			return "ON"
		}
		return "OFF"
	case settingsVolumeIndex:
		return fmt.Sprintf("< %d%% >", g.settingsVolumePercent())
	case settingsControllerDebugIndex:
		return "OPEN"
	case settingsDebugModeIndex:
		if g.debugEnabled {
			return "ON"
		}
		return "OFF"
	case settingsAboutIndex:
		return "OPEN"
	case settingsBackIndex:
		return "LOBBY"
	}
	return ""
}

func (g *Game) adjustSettings(direction int) {
	if direction == 0 {
		return
	}
	changed := false
	switch g.settingsFocus {
	case settingsSoundIndex:
		if g.sound != nil {
			g.sound.ToggleMute()
			changed = true
		}
	case settingsMusicIndex:
		if g.sound != nil {
			g.sound.ToggleMusic()
			changed = true
		}
	case settingsVolumeIndex:
		if g.sound != nil {
			volume := g.sound.MusicVolumePercent() + direction*10
			if direction > 0 && volume > 100 {
				volume = 0
			}
			g.sound.SetMusicVolumePercent(volume)
			changed = true
		}
	case settingsDebugModeIndex:
		g.debugEnabled = !g.debugEnabled
		changed = true
	}
	if changed && g.sound != nil {
		g.sound.Play(sound.MenuFocus)
	}
}

func (g *Game) activateSettingsItem() {
	switch g.settingsFocus {
	case settingsSoundIndex, settingsMusicIndex, settingsDebugModeIndex:
		g.adjustSettings(1)
	case settingsVolumeIndex:
		g.adjustSettings(1)
	case settingsControllerDebugIndex:
		g.controllerDebugOpen = true
		if g.sound != nil {
			g.sound.Play(sound.MenuSelect)
		}
	case settingsAboutIndex:
		g.openCreditsFromSettings()
	case settingsBackIndex:
		if g.sound != nil {
			g.sound.Play(sound.MenuSelect)
		}
		g.closeSettings()
	}
}

func (g *Game) updateSettings() {
	if g.controllerDebugOpen {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || len(inpututil.AppendJustPressedKeys(nil)) > 0 || len(g.pressedIDs) > 0 || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.controllerDebugOpen = false
		}
		for _, id := range ebiten.AppendGamepadIDs(g.gamepadIDs[:0]) {
			if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) {
				g.controllerDebugOpen = false
			}
		}
		return
	}

	navigate := func(direction controls.MenuDirection) {
		previous := g.settingsFocus
		switch direction {
		case controls.MenuUp:
			g.settingsFocus = (g.settingsFocus - 1 + settingsItemCount) % settingsItemCount
		case controls.MenuDown:
			g.settingsFocus = (g.settingsFocus + 1) % settingsItemCount
		}
		if g.settingsFocus != previous && g.sound != nil {
			g.sound.Play(sound.MenuFocus)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		navigate(controls.MenuUp)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		navigate(controls.MenuDown)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		g.adjustSettings(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		g.adjustSettings(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.activateSettingsItem()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.sound != nil {
			g.sound.Play(sound.MenuSelect)
		}
		g.closeSettings()
		return
	}

	for _, id := range ebiten.AppendGamepadIDs(g.gamepadIDs[:0]) {
		xDirection := controls.AxisDirection(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal), g.stickX[int(id)])
		yDirection := controls.AxisDirection(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical), g.stickY[int(id)])
		if yDirection != g.stickY[int(id)] && yDirection != 0 {
			if yDirection < 0 {
				navigate(controls.MenuUp)
			} else {
				navigate(controls.MenuDown)
			}
		}
		if xDirection != g.stickX[int(id)] && xDirection != 0 {
			g.adjustSettings(xDirection)
		}
		g.stickX[int(id)], g.stickY[int(id)] = xDirection, yDirection
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftTop) {
			navigate(controls.MenuUp)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftBottom) {
			navigate(controls.MenuDown)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftLeft) {
			g.adjustSettings(-1)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftRight) {
			g.adjustSettings(1)
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) {
			g.activateSettingsItem()
			return
		}
		if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) {
			if g.sound != nil {
				g.sound.Play(sound.MenuSelect)
			}
			g.closeSettings()
			return
		}
	}

	for _, id := range g.pressedIDs {
		x, y := ebiten.TouchPosition(id)
		if g.handleSettingsPointer(x, y) {
			return
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.handleSettingsPointer(x, y)
	}
}

func (g *Game) handleSettingsPointer(x, y int) bool {
	for index := 0; index < settingsItemCount; index++ {
		row := settingsRow(index)
		if !row.contains(x, y) {
			continue
		}
		if g.settingsFocus != index {
			g.settingsFocus = index
			if g.sound != nil {
				g.sound.Play(sound.MenuFocus)
			}
		}
		if index == settingsVolumeIndex {
			if x < row.X+row.W/2 {
				g.adjustSettings(-1)
			} else {
				g.adjustSettings(1)
			}
		} else {
			g.activateSettingsItem()
		}
		return true
	}
	return false
}

func (g *Game) settingsStatus() string {
	soundState := "SOUND OFF"
	if g.sound != nil && !g.sound.Muted() {
		soundState = "SOUND ON"
	}
	musicState := "MUSIC OFF"
	if g.sound != nil && g.sound.MusicEnabled() {
		musicState = "MUSIC ON"
	}
	return fmt.Sprintf("%s · %s · %d%%", soundState, musicState, g.settingsVolumePercent())
}

func (g *Game) drawSettings(screen *ebiten.Image) {
	screen.Fill(background)
	drawPaperDoodles(screen)
	drawCenteredText(screen, "SETTINGS", g.face(52), logicalWidth/2, 36, accent)
	drawCenteredText(screen, "Audio, controller tools and project information", g.face(20), logicalWidth/2, 104, muted)
	labels := []string{"SOUND", "MUSIC", "MUSIC VOLUME", "CONTROLLER DEBUG", "DEBUG MODE", "ABOUT & CREDITS", "BACK"}
	for index, label := range labels {
		row := settingsRow(index)
		ebitenutil.DrawRect(screen, float64(row.X), float64(row.Y), float64(row.W), float64(row.H), panel)
		if g.settingsFocus == index {
			ebitenutil.DrawRect(screen, float64(row.X-4), float64(row.Y-4), float64(row.W+8), 4, white)
			ebitenutil.DrawRect(screen, float64(row.X-4), float64(row.Y+row.H), float64(row.W+8), 4, white)
			ebitenutil.DrawRect(screen, float64(row.X-4), float64(row.Y-4), 4, float64(row.H+8), white)
			ebitenutil.DrawRect(screen, float64(row.X+row.W), float64(row.Y-4), 4, float64(row.H+8), white)
		}
		drawText(screen, label, g.face(20), float64(row.X+24), float64(row.Y+16), white)
		value := g.settingsValue(index)
		face := g.face(20)
		width, _ := text.Measure(value, face, 0)
		drawText(screen, value, face, float64(row.X+row.W-24)-width, float64(row.Y+16), accent)
	}
	drawCenteredText(screen, "Up/Down select · Left/Right adjust · A/Enter activates · B/Esc returns", g.face(17), logicalWidth/2, 650, muted)
	g.drawControllerDebug(screen)
}
