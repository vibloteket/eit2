package sound

import "testing"

type fakeMusicPlayer struct {
	playing  bool
	plays    int
	pauses   int
	position int
}

func (p *fakeMusicPlayer) Play()           { p.playing = true; p.plays++ }
func (p *fakeMusicPlayer) Pause()          { p.playing = false; p.pauses++ }
func (p *fakeMusicPlayer) IsPlaying() bool { return p.playing }
func (p *fakeMusicPlayer) Close() error    { p.playing = false; return nil }

func musicTestManager() (*Manager, *fakeMusicPlayer, *fakeMusicPlayer) {
	lobby, match := &fakeMusicPlayer{}, &fakeMusicPlayer{}
	return &Manager{
		music:      map[MusicTrack]musicPlayer{LobbyMusic: lobby, MatchMusic: match},
		musicTrack: LobbyMusic, musicEnabled: true,
	}, lobby, match
}

func TestMusicSceneTransitionsResumeWithoutOverlap(t *testing.T) {
	m, lobby, match := musicTestManager()
	m.updateMusic(true)
	if !lobby.playing || match.playing || !m.MusicPlaying() {
		t.Fatal("initial scene must play only lobby music")
	}
	lobby.position = 12345
	m.SetMusicTrack(LobbyMusic)
	for range 10 {
		m.updateMusic(true)
	}
	if lobby.plays != 1 || lobby.pauses != 0 {
		t.Fatal("repeated updates of the same scene must not restart its track")
	}

	m.SetMusicTrack(MatchMusic)
	if lobby.playing || match.playing || m.MusicPlaying() {
		t.Fatal("old player must stop before new player starts")
	}
	m.updateMusic(true)
	if lobby.playing || !match.playing {
		t.Fatal("match must play only its own track")
	}
	match.position = 6789
	m.SetMusicTrack(LobbyMusic)
	m.updateMusic(true)
	if !lobby.playing || match.playing || lobby.position != 12345 || lobby.plays != 2 {
		t.Fatal("returning to lobby must resume, not rewind or overlap")
	}
	m.SetMusicTrack(MatchMusic)
	m.updateMusic(true)
	if !match.playing || lobby.playing || match.position != 6789 {
		t.Fatal("match track position must also be retained")
	}
}

func TestMusicReadinessMuteAndEnabledGates(t *testing.T) {
	for _, ready := range []bool{false, true} {
		for _, muted := range []bool{false, true} {
			for _, enabled := range []bool{false, true} {
				m, lobby, match := musicTestManager()
				m.muted, m.musicEnabled = muted, enabled
				m.SetMusicTrack(MatchMusic)
				m.updateMusic(ready)
				want := ready && !muted && enabled
				if match.playing != want || lobby.playing {
					t.Errorf("ready=%v muted=%v enabled=%v: playing=%v, want %v", ready, muted, enabled, match.playing, want)
				}
			}
		}
	}
}

func TestSceneChangesBeforeBrowserAudioIsReady(t *testing.T) {
	m, lobby, match := musicTestManager()
	m.updateMusic(false)
	m.SetMusicTrack(MatchMusic)
	m.updateMusic(false)
	m.SetMusicTrack(LobbyMusic)
	m.updateMusic(false)
	if lobby.plays != 0 || match.plays != 0 {
		t.Fatal("scene changes must not bypass browser readiness")
	}
	m.updateMusic(true)
	if lobby.plays != 1 || match.plays != 0 {
		t.Fatal("readiness must start only the latest selected scene")
	}
}

func TestMusicTogglesRemainIndependentAcrossScenes(t *testing.T) {
	m, lobby, match := musicTestManager()
	m.updateMusic(true)
	if !m.ToggleMute() || lobby.playing {
		t.Fatal("mute must stop music immediately")
	}
	m.SetMusicTrack(MatchMusic)
	m.updateMusic(true)
	if match.playing {
		t.Fatal("scene selection must not override mute")
	}
	if m.ToggleMusic() || !m.Muted() {
		t.Fatal("music-off must not change master mute")
	}
	if m.ToggleMute() {
		t.Fatal("expected unmuted")
	}
	m.updateMusic(true)
	if match.playing {
		t.Fatal("unmuting must not override music-off")
	}
	if !m.ToggleMusic() {
		t.Fatal("expected music-on")
	}
	m.updateMusic(true)
	if !match.playing || lobby.playing {
		t.Fatal("enabling music must play only the selected scene")
	}
	m.ToggleMusic()
	if match.playing || m.Muted() {
		t.Fatal("music-off must stop music without muting effects")
	}
	m.SetMusicTrack(LobbyMusic)
	m.updateMusic(true)
	if lobby.playing {
		t.Fatal("music-off must persist when returning to lobby")
	}
}

func TestMissingAndInvalidMusicTracksAreSafe(t *testing.T) {
	m, lobby, _ := musicTestManager()
	m.updateMusic(true)
	m.SetMusicTrack(MusicTrack(99))
	if m.musicTrack != LobbyMusic || !lobby.playing || lobby.pauses != 0 {
		t.Fatal("unknown track should leave current playback intact")
	}
	var absent *Manager
	absent.SetMusicTrack(LobbyMusic)
	absent.Update()
	if absent.MusicPlaying() || absent.Ready() || absent.MusicEnabled() {
		t.Fatal("nil manager must be inactive")
	}
	empty := &Manager{musicEnabled: true}
	empty.updateMusic(true)
	if empty.MusicPlaying() {
		t.Fatal("missing player must be inactive")
	}
}
