package ui

import (
	"testing"

	"github.com/vibloteket/eit2/internal/sound"
)

func TestSceneMusicSelection(t *testing.T) {
	if got := viewLobby.musicTrack(); got != sound.LobbyMusic {
		t.Fatalf("lobby selected %v", got)
	}
	if got := viewPlay.musicTrack(); got != sound.MatchMusic {
		t.Fatalf("match selected %v", got)
	}
	// Pause, match settings and results retain viewPlay, so they must not
	// accidentally start the lobby track while a match is still on screen.
	g := Game{view: viewPlay, paused: true}
	if got := g.view.musicTrack(); got != sound.MatchMusic {
		t.Fatalf("paused match selected %v", got)
	}
}
