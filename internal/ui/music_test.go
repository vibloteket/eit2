package ui

import (
	"testing"

	"github.com/vibloteket/eit2/internal/lobby"
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

func TestPauseAndGameOverRestartUseCommonRestartPath(t *testing.T) {
	g := Game{view: viewPlay, round: 4, paused: true}
	g.Lobby.Join(lobby.Device{Kind: lobby.DeviceTouch, Name: "test"})
	g.overlayFocus = 1 // Resume, Restart, Lobby
	g.activateOverlay(false)
	if g.round != 5 || g.paused || g.view != viewPlay || g.match == nil {
		t.Fatal("pause Restart did not start a new round")
	}
	first := g.match
	g.overlayFocus = 0 // Restart, Lobby on game over
	g.activateOverlay(true)
	if g.round != 6 || g.match == first || g.view != viewPlay {
		t.Fatal("game-over Restart did not use the same restart path")
	}
}

func TestNewMatchesCycleFourTracksButRestartKeepsSelection(t *testing.T) {
	g := Game{view: viewLobby}
	g.Lobby.Join(lobby.Device{Kind: lobby.DeviceTouch, Name: "test"})
	want := []sound.MusicTrack{sound.MatchMusic, sound.MatchMusicHandel, sound.MatchMusicBachSonata, sound.MatchMusicVivaldi}
	for i := 0; i < 10; i++ {
		g.start()
		if g.selectedMusicTrack() != want[i%len(want)] || g.round != 1 {
			t.Fatalf("new match %d: track=%v round=%d", i, g.selectedMusicTrack(), g.round)
		}
		next := g.nextMatchMusic
		g.restart()
		if g.selectedMusicTrack() != want[i%len(want)] || g.nextMatchMusic != next || g.round != 2 {
			t.Fatal("Restart changed playlist or failed to reset round")
		}
		g.backToLobby()
		if g.selectedMusicTrack() != sound.LobbyMusic {
			t.Fatal("lobby music changed")
		}
		g.openCredits()
		if g.selectedMusicTrack() != sound.LobbyMusic {
			t.Fatal("credits music changed")
		}
		g.closeCredits()
	}
}
