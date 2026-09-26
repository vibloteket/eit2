package sound

import "testing"

func TestMusicVolumeScalesAllTracksWithoutChangingGates(t *testing.T) {
	m, lobby, match := musicTestManager()
	for _, track := range []MusicTrack{MatchMusicHandel, MatchMusicBachSonata, MatchMusicVivaldi} {
		m.music[track] = &fakeMusicPlayer{}
	}
	if got := m.MusicVolumePercent(); got != 100 {
		t.Fatalf("default music volume = %d, want 100", got)
	}
	m.SetMusicVolumePercent(70)
	if got := m.MusicVolumePercent(); got != 70 {
		t.Fatalf("music volume = %d, want 70", got)
	}
	for track, player := range m.music {
		want := musicVolumes[track] * .70
		if got := player.(*fakeMusicPlayer).volume; got != want {
			t.Fatalf("track %d volume = %f, want %f", track, got, want)
		}
	}
	m.SetMusicVolumePercent(-20)
	if got := m.MusicVolumePercent(); got != 0 {
		t.Fatal("volume must clamp at zero")
	}
	m.SetMusicVolumePercent(140)
	if got := m.MusicVolumePercent(); got != 100 {
		t.Fatal("volume must clamp at 100")
	}
	m.ToggleMusic()
	m.SetMusicVolumePercent(40)
	m.updateMusic(true)
	if lobby.playing || match.playing {
		t.Fatal("volume changes must not bypass music-off")
	}
	m.ToggleMusic()
	m.updateMusic(true)
	if !lobby.playing || lobby.volume != musicVolumes[LobbyMusic]*.40 {
		t.Fatal("volume change did not apply when music resumed")
	}
	var absent *Manager
	absent.SetMusicVolumePercent(50)
	if absent.MusicVolumePercent() != 0 {
		t.Fatal("nil manager must report no active volume")
	}
}
