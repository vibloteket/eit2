package sound

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

type fakeMusicPlayer struct {
	playing             bool
	plays               int
	pauses              int
	position            int
	rewinds             int
	rewindErr           error
	rewoundWhilePlaying bool
	volume              float64
}

func (p *fakeMusicPlayer) Play()                    { p.playing = true; p.plays++ }
func (p *fakeMusicPlayer) Pause()                   { p.playing = false; p.pauses++ }
func (p *fakeMusicPlayer) IsPlaying() bool          { return p.playing }
func (p *fakeMusicPlayer) Close() error             { p.playing = false; return nil }
func (p *fakeMusicPlayer) SetVolume(volume float64) { p.volume = volume }
func (p *fakeMusicPlayer) Rewind() error {
	p.rewinds++
	p.rewoundWhilePlaying = p.playing
	if p.rewindErr != nil {
		return p.rewindErr
	}
	p.position = 0
	return nil
}

func musicTestManager() (*Manager, *fakeMusicPlayer, *fakeMusicPlayer) {
	lobby, match := &fakeMusicPlayer{}, &fakeMusicPlayer{}
	return &Manager{
		music:      map[MusicTrack]musicPlayer{LobbyMusic: lobby, MatchMusic: match},
		musicTrack: LobbyMusic, musicEnabled: true, musicVolume: 1,
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

func TestRestartMusicResetsOnlyMatchAndRespectsAllGates(t *testing.T) {
	for _, ready := range []bool{false, true} {
		for _, muted := range []bool{false, true} {
			for _, enabled := range []bool{false, true} {
				m, lobby, match := musicTestManager()
				lobby.position = 6789
				m.SetMusicTrack(MatchMusic)
				m.updateMusic(true)
				match.position = 12345
				m.muted, m.musicEnabled = muted, enabled
				if err := m.RestartMusic(MatchMusic); err != nil {
					t.Fatal(err)
				}
				if match.position != 0 || match.rewinds != 1 || match.playing || match.rewoundWhilePlaying {
					t.Fatal("restart must pause before rewinding to the beginning")
				}
				if lobby.position != 6789 || lobby.rewinds != 0 || lobby.playing {
					t.Fatal("match restart must not reset or play lobby music")
				}
				if m.muted != muted || m.musicEnabled != enabled || m.musicTrack != MatchMusic {
					t.Fatal("restart must not change settings or track selection")
				}
				m.updateMusic(ready)
				if match.playing != (ready && !muted && enabled) {
					t.Fatal("restart bypassed an audio gate")
				}
				m.muted, m.musicEnabled = false, true
				m.updateMusic(true)
				if !match.playing || match.position != 0 {
					t.Fatal("reenabling must start at the reset position")
				}
				m.updateMusic(true)
				if match.rewinds != 1 {
					t.Fatal("ordinary update must not repeatedly restart")
				}
			}
		}
	}
}

func TestRestartInactiveMusicAndErrors(t *testing.T) {
	m, lobby, match := musicTestManager()
	m.updateMusic(true)
	lobby.position, match.position = 20, 40
	if err := m.RestartMusic(MatchMusic); err != nil {
		t.Fatal(err)
	}
	if !lobby.playing || lobby.position != 20 || match.position != 0 || match.playing || m.musicTrack != LobbyMusic {
		t.Fatal("resetting inactive match music must leave the current scene alone")
	}
	failure := errors.New("test seek error")
	match.rewindErr = failure
	if err := m.RestartMusic(MatchMusic); !errors.Is(err, failure) {
		t.Fatal("seek error must be reported")
	}
	if err := m.RestartMusic(MusicTrack(99)); err != nil {
		t.Fatal(err)
	}
	var absent *Manager
	if err := absent.RestartMusic(MatchMusic); err != nil {
		t.Fatal(err)
	}
}

func TestGameplayInfiniteLoopCanRewindToExactOpening(t *testing.T) {
	data, err := files.ReadFile("audio/" + musicFilenames[MatchMusic])
	if err != nil {
		t.Fatal(err)
	}
	pcm, err := decodeWAV(data)
	if err != nil {
		t.Fatal(err)
	}
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
	if _, err := io.CopyN(io.Discard, loop, int64(len(pcm)+sampleRate*4)); err != nil {
		t.Fatal(err)
	}
	if _, err := loop.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	opening := make([]byte, 4096)
	if _, err := io.ReadFull(loop, opening); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(opening, pcm[:len(opening)]) {
		t.Fatal("rewound loop differs from the opening PCM")
	}
}

func TestSecondGameplayTrackSwitchAndRestart(t *testing.T) {
	m, lobby, first := musicTestManager()
	second := &fakeMusicPlayer{position: 500}
	m.music[MatchMusicHandel] = second
	m.SetMusicTrack(MatchMusic)
	m.updateMusic(true)
	first.position = 123
	m.SetMusicTrack(MatchMusicHandel)
	m.updateMusic(true)
	if first.playing || lobby.playing || !second.playing {
		t.Fatal("only Handel should play")
	}
	if err := m.RestartMusic(MatchMusicHandel); err != nil {
		t.Fatal(err)
	}
	if second.position != 0 || second.rewinds != 1 || first.position != 123 || first.rewinds != 0 {
		t.Fatal("Restart must reset only selected Handel")
	}
	m.ToggleMusic()
	m.updateMusic(true)
	if second.playing {
		t.Fatal("music-off must gate Handel")
	}
	m.ToggleMusic()
	m.updateMusic(true)
	if !second.playing || first.playing {
		t.Fatal("Handel must resume alone")
	}
}
