package ui

import (
	"bytes"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/vibloteket/eit2/internal/sound"
	"golang.org/x/image/font/gofont/goregular"
)

func TestSettingsIndicesMatchRows(t *testing.T) {
	if settingsSoundIndex != 0 || settingsMusicIndex != 1 || settingsVolumeIndex != 2 || settingsAboutIndex != 5 || settingsBackIndex != 6 {
		t.Fatal("settings indices must match the rendered row order")
	}
}

func TestSettingsMenuStructureAndReturn(t *testing.T) {
	g := Game{view: viewLobby, lobbyFocus: lobbySettingsIndex}
	g.openSettings()
	if g.view != viewSettings || g.settingsFocus != settingsSoundIndex || g.controllerDebugOpen {
		t.Fatal("Settings must open as a separate lobby view")
	}
	if g.view.musicTrack() != sound.LobbyMusic {
		t.Fatal("Settings must keep lobby music")
	}
	g.settingsFocus = settingsBackIndex
	g.activateSettingsItem()
	if g.view != viewLobby || g.lobbyFocus != lobbySettingsIndex {
		t.Fatal("Back must return to lobby with Settings focused")
	}
}

func TestSettingsValuesAndDebugItems(t *testing.T) {
	g := Game{}
	if got := g.settingsValue(settingsSoundIndex); got != "OFF" {
		t.Fatalf("sound without manager = %q", got)
	}
	if got := g.settingsValue(settingsMusicIndex); got != "OFF" {
		t.Fatalf("music without manager = %q", got)
	}
	if got := g.settingsValue(settingsVolumeIndex); got != "< 100% >" {
		t.Fatalf("fallback volume = %q", got)
	}
	g.settingsFocus = settingsDebugModeIndex
	g.activateSettingsItem()
	if !g.debugEnabled || g.settingsValue(settingsDebugModeIndex) != "ON" {
		t.Fatal("Debug Mode did not toggle")
	}
	g.settingsFocus = settingsControllerDebugIndex
	g.activateSettingsItem()
	if !g.controllerDebugOpen {
		t.Fatal("Controller Debug did not open")
	}
	g.controllerDebugOpen = false
	g.settingsFocus = settingsAboutIndex
	g.activateSettingsItem()
	if g.view != viewCredits || !g.creditsReturnToSettings {
		t.Fatal("About did not open Credits with Settings return")
	}
}

func TestSettingsTextFits(t *testing.T) {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		t.Fatal(err)
	}
	g := Game{fontSource: source}
	labels := []string{"SOUND", "MUSIC", "MUSIC VOLUME", "CONTROLLER DEBUG", "DEBUG MODE", "ABOUT & CREDITS", "BACK"}
	for index, label := range labels {
		row := settingsRow(index)
		if row.X < 0 || row.X+row.W > logicalWidth || row.Y < 0 || row.Y+row.H > 630 {
			t.Fatalf("row %d outside screen: %+v", index, row)
		}
		labelWidth, _ := text.Measure(label, g.face(20), 0)
		valueWidth, _ := text.Measure("< 100% >", g.face(20), 0)
		if labelWidth+valueWidth+72 > float64(row.W) {
			t.Fatalf("row %d text does not fit", index)
		}
	}
	if width, _ := text.Measure(g.settingsStatus(), g.face(12), 0); width > float64(settingsButton().W-30) {
		t.Fatal("lobby Settings status does not fit")
	}
}
