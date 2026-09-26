// Package sound plays semantic game effects and scene-specific looping music.
package sound

import (
	"bytes"
	"embed"
	"fmt"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100
const effectVolume = .36

type Effect string

const (
	CountIn    Effect = "count-in"
	MenuFocus  Effect = "menu-focus"
	MenuSelect Effect = "menu-select"
	Join       Effect = "join"
	Leave      Effect = "leave"
	Rotate     Effect = "rotate"
	Lock       Effect = "lock"
	HardDrop   Effect = "hard-drop"
	Line       Effect = "line"
	FourLine   Effect = "four-line"
	Pickup     Effect = "pickup"
	Attack     Effect = "attack"
	Antidote   Effect = "antidote"
	GameOver   Effect = "game-over"
	Winner     Effect = "winner"
)

//go:embed audio/*.wav
var files embed.FS

var filenames = map[Effect]string{
	CountIn:   "count-in.wav",
	MenuFocus: "menu-focus.wav", MenuSelect: "menu-select.wav",
	Join: "join.wav", Leave: "leave.wav", Rotate: "rotate.wav",
	Lock: "lock.wav", HardDrop: "hard-drop.wav", Line: "line.wav",
	FourLine: "four-line.wav", Pickup: "pickup.wav", Attack: "attack.wav",
	Antidote: "antidote.wav", GameOver: "game-over.wav", Winner: "winner.wav",
}

// MusicTrack selects a scene's music without changing the user's audio settings.
type MusicTrack uint8

const (
	LobbyMusic MusicTrack = iota
	MatchMusic
	MatchMusicHandel
	MatchMusicBachSonata
	MatchMusicVivaldi
)

var musicFilenames = map[MusicTrack]string{
	LobbyMusic:           "lobby-badinerie.wav",
	MatchMusic:           "gameplay-bach-bourrees.wav",
	MatchMusicHandel:     "gameplay-handel-allegro.wav",
	MatchMusicBachSonata: "gameplay-bach-sonata.wav",
	MatchMusicVivaldi:    "gameplay-vivaldi-allegro.wav",
}

// The four gameplay loops are RMS-matched; leave room for lock/drop/line/attack cues.
var musicVolumes = map[MusicTrack]float64{
	LobbyMusic:           .16,
	MatchMusic:           .08,
	MatchMusicHandel:     .08,
	MatchMusicBachSonata: .08,
	MatchMusicVivaldi:    .08,
}

// Keep playback control testable without opening an audio device.
type musicPlayer interface {
	Play()
	Pause()
	IsPlaying() bool
	Rewind() error
	Close() error
	SetVolume(volume float64)
}

type Manager struct {
	context        *audio.Context
	pcm            map[Effect][]byte
	players        map[Effect][]*audio.Player
	music          map[MusicTrack]musicPlayer
	musicTrack     MusicTrack
	muted          bool
	musicEnabled   bool
	musicSuspended bool
	musicVolume    float64
	mu             sync.Mutex
}

func New() (*Manager, error) {
	context := audio.CurrentContext()
	if context == nil {
		context = audio.NewContext(sampleRate)
	}
	manager := &Manager{
		context: context, pcm: make(map[Effect][]byte), players: make(map[Effect][]*audio.Player),
		music: make(map[MusicTrack]musicPlayer), musicTrack: LobbyMusic, musicEnabled: true,
		musicVolume: 1,
	}
	// If initialization fails, release any music players already created.
	initialized := false
	defer func() {
		if !initialized {
			for _, player := range manager.music {
				_ = player.Close()
			}
		}
	}()
	for effect, filename := range filenames {
		data, err := files.ReadFile("audio/" + filename)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", filename, err)
		}
		pcm, err := decodeWAV(data)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", filename, err)
		}
		manager.pcm[effect] = pcm
	}
	for _, track := range []MusicTrack{LobbyMusic, MatchMusic, MatchMusicHandel, MatchMusicBachSonata, MatchMusicVivaldi} {
		filename := musicFilenames[track]
		musicWAV, err := files.ReadFile("audio/" + filename)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", filename, err)
		}
		musicPCM, err := decodeWAV(musicWAV)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", filename, err)
		}
		loop := audio.NewInfiniteLoop(bytes.NewReader(musicPCM), int64(len(musicPCM)))
		player, err := context.NewPlayer(loop)
		if err != nil {
			return nil, fmt.Errorf("create %s player: %w", filename, err)
		}
		player.SetVolume(musicVolumes[track] * manager.musicVolume)
		manager.music[track] = player
	}
	initialized = true
	return manager, nil
}

func (m *Manager) Ready() bool { return m != nil && m.context != nil && m.context.IsReady() }
func (m *Manager) Muted() bool { return m == nil || m.muted }
func (m *Manager) ToggleMute() bool {
	m.muted = !m.muted
	if m.muted {
		m.pauseMusic()
	}
	return m.muted
}

func (m *Manager) MusicEnabled() bool { return m != nil && m.musicEnabled }

// MusicVolumePercent is the session-scoped user volume multiplier. It is not
// persisted yet; 100 preserves the established per-track lobby/match balance.
func (m *Manager) MusicVolumePercent() int {
	if m == nil {
		return 0
	}
	return int(math.Round(m.musicVolume * 100))
}

func (m *Manager) SetMusicVolumePercent(percent int) {
	if m == nil {
		return
	}
	percent = max(0, min(100, percent))
	m.musicVolume = float64(percent) / 100
	for track, player := range m.music {
		player.SetVolume(musicVolumes[track] * m.musicVolume)
	}
}

func (m *Manager) MusicPlaying() bool {
	player := m.activeMusic()
	return player != nil && player.IsPlaying()
}
func (m *Manager) ToggleMusic() bool {
	m.musicEnabled = !m.musicEnabled
	if !m.musicEnabled {
		m.pauseMusic()
	}
	return m.musicEnabled
}

// SetMusicTrack pauses the old scene before selecting the new one. Each track
// resumes from its previous position: returning to the lobby does not force the
// same opening again. Selection never overrides mute, music-off or browser gating.
func (m *Manager) SetMusicTrack(track MusicTrack) {
	if m == nil || m.musicTrack == track || m.music[track] == nil {
		return
	}
	m.pauseMusic()
	m.musicTrack = track
}

// RestartMusic resets only the requested track. Update resumes it only if it
// is selected, audio is ready, and the user's mute/music settings allow it.
// Pausing before Rewind avoids playing part of the old position during reset.
func (m *Manager) RestartMusic(track MusicTrack) error {
	if m == nil || m.music[track] == nil {
		return nil
	}
	player := m.music[track]
	player.Pause()
	if err := player.Rewind(); err != nil {
		return fmt.Errorf("restart music %d: %w", track, err)
	}
	return nil
}

func (m *Manager) activeMusic() musicPlayer {
	if m == nil {
		return nil
	}
	return m.music[m.musicTrack]
}

func (m *Manager) pauseMusic() {
	for _, player := range m.music {
		player.Pause()
	}
}

// SetMusicSuspended is a temporary presentation gate, not a user preference.
// Effects (including preparation clicks) retain their normal mute/readiness rules.
func (m *Manager) SetMusicSuspended(suspended bool) {
	if m == nil || m.musicSuspended == suspended {
		return
	}
	m.musicSuspended = suspended
	if suspended {
		m.pauseMusic()
	}
}

// Update starts music once the platform audio context becomes ready. Browsers
// normally reach this state after the first user interaction.
func (m *Manager) Update() { m.updateMusic(m.Ready()) }

func (m *Manager) updateMusic(ready bool) {
	player := m.activeMusic()
	if player == nil || m.muted || !m.musicEnabled || m.musicSuspended || !ready || player.IsPlaying() {
		return
	}
	player.Play()
}

func (m *Manager) Play(effect Effect) {
	if m == nil || m.muted || !m.context.IsReady() {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	pool := m.players[effect]
	for _, player := range pool {
		if !player.IsPlaying() {
			_ = player.Rewind()
			player.Play()
			return
		}
	}
	if len(pool) >= 3 {
		return
	}
	player := m.context.NewPlayerFromBytes(m.pcm[effect])
	player.SetVolume(effectVolume)
	m.players[effect] = append(pool, player)
	player.Play()
}
