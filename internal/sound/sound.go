// Package sound plays semantic game effects and scene-specific looping music.
package sound

import (
	"bytes"
	"embed"
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100
const effectVolume = .36

type Effect string

const (
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
	MatchMusicG2
)

var musicFilenames = map[MusicTrack]string{
	LobbyMusic:   "lobby-badinerie.wav",
	MatchMusic:   "gameplay-beethoven.wav",
	MatchMusicG2: "gameplay-beethoven-g2.wav",
}

// The gameplay track is mastered much louder than the former procedural loop.
// Keep the selected PCM unchanged, but leave room for lock/drop/line/attack cues.
var musicVolumes = map[MusicTrack]float64{
	LobbyMusic:   .16,
	MatchMusic:   .08,
	MatchMusicG2: .08,
}

// Keep playback control testable without opening an audio device.
type musicPlayer interface {
	Play()
	Pause()
	IsPlaying() bool
	Rewind() error
	Close() error
}

type Manager struct {
	context      *audio.Context
	pcm          map[Effect][]byte
	players      map[Effect][]*audio.Player
	music        map[MusicTrack]musicPlayer
	musicTrack   MusicTrack
	muted        bool
	musicEnabled bool
	mu           sync.Mutex
}

func New() (*Manager, error) {
	context := audio.CurrentContext()
	if context == nil {
		context = audio.NewContext(sampleRate)
	}
	manager := &Manager{
		context: context, pcm: make(map[Effect][]byte), players: make(map[Effect][]*audio.Player),
		music: make(map[MusicTrack]musicPlayer), musicTrack: LobbyMusic, musicEnabled: true,
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
	for _, track := range []MusicTrack{LobbyMusic, MatchMusic, MatchMusicG2} {
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
		player.SetVolume(musicVolumes[track])
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

// Update starts music once the platform audio context becomes ready. Browsers
// normally reach this state after the first user interaction.
func (m *Manager) Update() { m.updateMusic(m.Ready()) }

func (m *Manager) updateMusic(ready bool) {
	player := m.activeMusic()
	if player == nil || m.muted || !m.musicEnabled || !ready || player.IsPlaying() {
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
