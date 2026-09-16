package ui

import (
	"testing"

	"bytes"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/vibloteket/eit2/internal/lobby"
	"github.com/vibloteket/eit2/internal/sound"
	"golang.org/x/image/font/gofont/goregular"
)

func TestCreditsIsSeparateViewAndKeepsLobbyMusic(t *testing.T) {
	g := Game{view: viewLobby, lobbyFocus: 2, controllerDebugOpen: true}
	g.Lobby.Join(lobby.Device{Kind: lobby.DeviceTouch, Name: "test player"})
	if g.activateLobbyMenu(5) {
		t.Fatal("Credits must not request application exit")
	}
	if g.view != viewCredits || g.view == viewLobby || g.view == viewPlay {
		t.Fatal("Credits must use its own view")
	}
	if g.view.musicTrack() != sound.LobbyMusic || g.controllerDebugOpen {
		t.Fatal("Credits must retain lobby music and hide debug overlay")
	}
	g.closeCredits()
	if g.view != viewLobby || g.lobbyFocus != 5 || len(g.Lobby.Slots) != 1 {
		t.Fatal("return must preserve players and focus Credits")
	}
	if !isWeb() && !g.activateLobbyMenu(6) {
		t.Fatal("native Exit must remain available after adding Credits")
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

func TestCreditsButtonFitsUtilityRow(t *testing.T) {
	debug, credits, exit := debugLobbyButton(), creditsButton(), exitButton()
	if credits.X < debug.X+debug.W || exit.X < credits.X+credits.W || exit.X+exit.W > logicalWidth {
		t.Fatal("Credits/Exit utility buttons overlap or leave screen")
	}
	buttons := lobbyMenuButtons()
	if buttons[5] != credits {
		t.Fatal("navigation and rendered Credits positions disagree")
	}
}
