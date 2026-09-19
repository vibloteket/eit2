package sound

import "testing"

func TestEveryClassicalTrackGatesAndRestartsIndependently(t *testing.T) {
	tracks := []MusicTrack{LobbyMusic, MatchMusic, MatchMusicHandel, MatchMusicBachSonata, MatchMusicVivaldi}
	for _, selected := range tracks[1:] {
		for _, ready := range []bool{false, true} {
			for _, muted := range []bool{false, true} {
				for _, enabled := range []bool{false, true} {
					players := map[MusicTrack]*fakeMusicPlayer{}
					m := &Manager{music: map[MusicTrack]musicPlayer{}, musicTrack: LobbyMusic, muted: muted, musicEnabled: enabled}
					for _, track := range tracks {
						p := &fakeMusicPlayer{position: 100 + int(track)}
						players[track], m.music[track] = p, p
					}
					m.SetMusicTrack(selected)
					if err := m.RestartMusic(selected); err != nil {
						t.Fatal(err)
					}
					m.updateMusic(ready)
					for _, track := range tracks {
						p := players[track]
						want := track == selected && ready && !muted && enabled
						if p.playing != want || p.rewoundWhilePlaying {
							t.Fatalf("selected=%v track=%v ready=%v muted=%v enabled=%v: invalid playback", selected, track, ready, muted, enabled)
						}
						if track == selected {
							if p.rewinds != 1 || p.position != 0 {
								t.Fatal("selected track was not rewound exactly once")
							}
						} else if p.rewinds != 0 || p.position != 100+int(track) {
							t.Fatal("Restart changed another track's position")
						}
					}
					m.SetMusicTrack(LobbyMusic)
					m.updateMusic(ready)
					if players[LobbyMusic].position != 100 || players[LobbyMusic].rewinds != 0 || m.muted != muted || m.musicEnabled != enabled {
						t.Fatal("returning to lobby changed music position or preferences")
					}
				}
			}
		}
	}
}
