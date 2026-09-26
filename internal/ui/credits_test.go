package ui

import (
	"bytes"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/vibloteket/eit2/internal/lobby"
	"github.com/vibloteket/eit2/internal/sound"
	"golang.org/x/image/font/gofont/goregular"
)

func TestCreditsIsSeparateViewAndKeepsLobbyMusic(t *testing.T) {
	g := Game{view: viewLobby, lobbyFocus: lobbySettingsIndex, controllerDebugOpen: true}
	g.Lobby.Join(lobby.Device{Kind: lobby.DeviceTouch, Name: "test player"})
	g.openSettings()
	g.settingsFocus = settingsAboutIndex
	g.activateSettingsItem()
	if g.view != viewCredits || g.view == viewLobby || g.view == viewPlay {
		t.Fatal("About & Credits must use its own view")
	}
	if g.view.musicTrack() != sound.LobbyMusic || g.controllerDebugOpen {
		t.Fatal("Credits must retain lobby music and hide debug overlay")
	}
	g.closeCredits()
	if g.view != viewSettings || g.settingsFocus != settingsAboutIndex || len(g.Lobby.Slots) != 1 {
		t.Fatal("return must go back to Settings and preserve players")
	}
	g.closeSettings()
	if g.view != viewLobby || g.lobbyFocus != lobbySettingsIndex {
		t.Fatal("Settings return must focus the lobby Settings button")
	}
	if !isWeb() && !g.activateLobbyMenu(2) {
		t.Fatal("native Exit must remain available after moving Credits")
	}
}

func TestCreditsTextFitsWithoutScrolling(t *testing.T) {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		t.Fatal(err)
	}
	g := Game{fontSource: source}
	lastBottom := 142.0
	for _, line := range creditLines {
		width, _ := text.Measure(line.Text, g.face(line.Size), 0)
		if width > 1160 || line.Y < lastBottom || line.Y+line.Size*1.3 > 625 {
			t.Fatalf("credit does not fit fixed screen: %q, width=%f y=%f", line.Text, width, line.Y)
		}
		lastBottom = line.Y + line.Size
	}
}

func TestLobbyButtonsFitUtilityRow(t *testing.T) {
	settings, exit := settingsButton(), exitButton()
	if settings.X < 0 || exit.X < settings.X+settings.W || exit.X+exit.W > logicalWidth {
		t.Fatal("Settings/Exit utility buttons overlap or leave screen")
	}
	buttons := lobbyMenuButtons()
	if buttons[1] != settings {
		t.Fatal("navigation and rendered Settings positions disagree")
	}
}
