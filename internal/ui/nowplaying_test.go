package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/vibloteket/eit2/internal/sound"
	"golang.org/x/image/font/gofont/goregular"
)

func TestNowPlayingLabels(t *testing.T) {
	want := map[sound.MusicTrack]string{
		sound.LobbyMusic:           "♪ Bach — Badinerie",
		sound.MatchMusic:           "♪ Bach — Bourrée I/II",
		sound.MatchMusicHandel:     "♪ Handel — HWV 369 IV",
		sound.MatchMusicBachSonata: "♪ Bach — BWV 1035 II",
		sound.MatchMusicVivaldi:    "♪ Vivaldi — RV 428 III",
	}
	if len(nowPlayingTitles) != len(want) {
		t.Fatal("missing now-playing title")
	}
	for track, label := range want {
		if got := nowPlayingLabelFor(track, true); got != label {
			t.Fatalf("track %d: %q, want %q", track, got, label)
		}
		if got := nowPlayingLabelFor(track, false); got != label+" · OFF" {
			t.Fatalf("track %d off: %q", track, got)
		}
	}
	if got := nowPlayingLabelFor(sound.MusicTrack(255), true); got != "" {
		t.Fatal("unknown track must not claim a title")
	}
}

func TestNowPlayingZeroVolumeIsMarkedOff(t *testing.T) {
	manager, err := sound.New()
	if err != nil {
		t.Fatal(err)
	}
	g := Game{view: viewLobby, sound: manager}
	if !strings.Contains(g.nowPlayingLabel(), "Bach — Badinerie") || strings.HasSuffix(g.nowPlayingLabel(), "OFF") {
		t.Fatalf("default label = %q", g.nowPlayingLabel())
	}
	manager.SetMusicVolumePercent(0)
	if !strings.HasSuffix(g.nowPlayingLabel(), " · OFF") {
		t.Fatalf("zero-volume label = %q", g.nowPlayingLabel())
	}
}

func TestNowPlayingFitsTopRight(t *testing.T) {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		t.Fatal(err)
	}
	g := Game{fontSource: source}
	face := g.face(16)
	for _, track := range []sound.MusicTrack{sound.LobbyMusic, sound.MatchMusic, sound.MatchMusicHandel, sound.MatchMusicBachSonata, sound.MatchMusicVivaldi} {
		label := nowPlayingLabelFor(track, false) // Include the longest OFF suffix.
		width, _ := text.Measure(label, face, 0)
		if width > 330 || logicalWidth-40-width < 850 {
			t.Fatalf("%q is too wide for the top-right area: %f", label, width)
		}
	}
}
