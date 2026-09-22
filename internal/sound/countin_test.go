package sound

import "testing"

func TestCountInSuspensionIsIndependentOfUserAudioSettings(t *testing.T) {
	for _, track := range []MusicTrack{MatchMusic, MatchMusicHandel, MatchMusicBachSonata, MatchMusicVivaldi} {
		for _, ready := range []bool{false, true} {
			for _, muted := range []bool{false, true} {
				for _, enabled := range []bool{false, true} {
					m, lobby, _ := musicTestManager()
					p := &fakeMusicPlayer{position: 999}
					m.music[track] = p
					lobby.position = 123
					m.SetMusicTrack(track)
					m.SetMusicSuspended(true)
					if err := m.RestartMusic(track); err != nil {
						t.Fatal(err)
					}
					m.muted, m.musicEnabled = muted, enabled
					m.updateMusic(ready)
					if p.playing || lobby.playing || p.position != 0 || p.rewinds != 1 {
						t.Fatal("music advanced during count-in")
					}
					m.ToggleMusic()
					m.ToggleMusic()
					m.ToggleMute()
					m.ToggleMute()
					m.updateMusic(ready)
					if p.playing || !m.musicSuspended {
						t.Fatal("user toggle bypassed count-in gate")
					}
					m.SetMusicSuspended(false)
					m.updateMusic(ready)
					if p.playing != (ready && !muted && enabled) || m.muted != muted || m.musicEnabled != enabled {
						t.Fatal("GO changed preferences or failed gating")
					}
					m.SetMusicTrack(LobbyMusic)
					m.updateMusic(ready)
					if lobby.position != 123 || lobby.rewinds != 0 || p.playing {
						t.Fatal("count-in affected lobby position or overlapped music")
					}
				}
			}
		}
	}
}

func TestCountInSoundIsAuditedShortProceduralEffect(t *testing.T) {
	if filenames[CountIn] != "count-in.wav" {
		t.Fatal("count-in sound mapping")
	}
	b, err := files.ReadFile("audio/" + filenames[CountIn])
	if err != nil {
		t.Fatal(err)
	}
	pcm, err := decodeWAV(b)
	if err != nil {
		t.Fatal(err)
	}
	seconds := float64(len(pcm)) / 4 / sampleRate
	if seconds < .03 || seconds > .09 {
		t.Fatalf("click is not short/dry: %fs", seconds)
	}
}
