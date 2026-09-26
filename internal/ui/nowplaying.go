package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/vibloteket/eit2/internal/sound"
)

var nowPlayingTitles = map[sound.MusicTrack]string{
	sound.LobbyMusic:           "Bach — Badinerie",
	sound.MatchMusic:           "Bach — Bourrée I/II",
	sound.MatchMusicHandel:     "Handel — HWV 369 IV",
	sound.MatchMusicBachSonata: "Bach — BWV 1035 II",
	sound.MatchMusicVivaldi:    "Vivaldi — RV 428 III",
}

func nowPlayingLabelFor(track sound.MusicTrack, audible bool) string {
	title := nowPlayingTitles[track]
	if title == "" {
		return ""
	}
	label := "♪ " + title
	if !audible {
		label += " · OFF"
	}
	return label
}

func (g *Game) nowPlayingLabel() string {
	audible := g.sound != nil && !g.sound.Muted() && g.sound.MusicEnabled() && g.sound.MusicVolumePercent() > 0
	return nowPlayingLabelFor(g.selectedMusicTrack(), audible)
}

func (g *Game) drawNowPlaying(screen *ebiten.Image, y float64) {
	label := g.nowPlayingLabel()
	if label == "" {
		return
	}
	face := g.face(16)
	width, _ := text.Measure(label, face, 0)
	drawText(screen, label, face, logicalWidth-40-width, y, muted)
}
