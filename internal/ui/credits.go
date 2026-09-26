package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/vibloteket/eit2/internal/sound"
	"github.com/vibloteket/eit2/internal/version"
)

// The complete text fits in the 1280x720 design space; no scrolling or modal
// overlay is needed. Full source/license details ship in CREDITS.md and ASSETS.md.
var creditLines = []struct {
	Text    string
	Y, Size float64
	Heading bool
}{
	{"GAME", 164, 18, true},
	{"Eit 2 — Victor Blomqvist and contributors", 192, 24, false},
	{"Based on the original Eit by Victor Blomqvist", 222, 22, false},
	{"MUSIC & AUDIO", 265, 18, true},
	{"Lobby: J. S. Bach — Badinerie, BWV 1067/VII", 293, 23, false},
	{"Matches: Bach — Bourrees / BWV 1035 · Handel — HWV 369", 321, 21, false},
	{"Vivaldi — RV 428 · Arrangements, rendering & effects: Eit 2", 349, 21, false},
	{"Flute & cello: VSCO 2 CE — Sam Gossner and Simon Dalzell", 377, 22, false},
	{"Sample editing: Elan Hickler / Soundemote · Samples: CC0", 405, 22, false},
	{"Historical public-domain scores · Full sources: MUSIC-SOURCES.md", 433, 20, false},
	{"TECHNOLOGY & LICENSES", 471, 18, true},
	{"Ebitengine — Hajime Hoshi and contributors · Apache 2.0", 499, 21, false},
	{"Go Regular — The Go Authors · BSD 3-Clause", 527, 21, false},
	{"Code: AGPL-3.0-or-later · Compositions: public domain", 555, 21, false},
	{"Full sources and licenses: github.com/vibloteket/eit2", 593, 20, false},
}

func (g *Game) openCredits() {
	g.creditsReturnToSettings = false
	g.openCreditsView()
}

func (g *Game) openCreditsFromSettings() {
	g.creditsReturnToSettings = true
	g.openCreditsView()
}

func (g *Game) openCreditsView() {
	g.view = viewCredits
	g.controllerDebugOpen = false
	if g.sound != nil {
		g.sound.Play(sound.MenuSelect)
	}
}

func (g *Game) closeCredits() {
	if g.creditsReturnToSettings {
		g.creditsReturnToSettings = false
		g.view = viewSettings
		g.settingsFocus = settingsAboutIndex
		return
	}
	g.view = viewLobby
	g.lobbyFocus = lobbySettingsIndex
}

func (g *Game) updateCredits() {
	// Only fresh presses dismiss. The held key/button/touch that opened this
	// view must not immediately close it. Returning also consumes the input:
	// it must not join a player, start a match or toggle audio in the lobby.
	pressed := len(inpututil.AppendJustPressedKeys(nil)) > 0 || len(g.pressedIDs) > 0
	for button := ebiten.MouseButton(0); button <= ebiten.MouseButtonMax; button++ {
		pressed = pressed || inpututil.IsMouseButtonJustPressed(button)
	}
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		// Deliberately include unjoined and non-standard controllers here.
		for button := ebiten.GamepadButton(0); button <= ebiten.GamepadButtonMax; button++ {
			pressed = pressed || inpututil.IsGamepadButtonJustPressed(id, button)
		}
	}
	if pressed {
		g.closeCredits()
	}
}

func (g *Game) drawCredits(screen *ebiten.Image) {
	screen.Fill(background)
	drawPaperDoodles(screen)
	drawCenteredText(screen, "CREDITS", g.face(52), logicalWidth/2, 36, accent)
	drawCenteredText(screen, "EIT 2 · v"+version.Value, g.face(21), logicalWidth/2, 105, muted)
	ebitenutil.DrawRect(screen, 140, 142, 1000, 2, panel)
	for _, line := range creditLines {
		ink := white
		if line.Heading {
			ink = accent
		}
		drawCenteredText(screen, line.Text, g.face(line.Size), logicalWidth/2, line.Y, ink)
	}
	drawCenteredText(screen, "Press any key or button, click, or tap to return.", g.face(21), logicalWidth/2, 651, muted)
}
