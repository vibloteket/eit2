// Package sound maps semantic game events to short embedded prototype effects.
package sound

import (
	"bytes"
	"embed"
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100

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

type Manager struct {
	context      *audio.Context
	pcm          map[Effect][]byte
	players      map[Effect][]*audio.Player
	music        *audio.Player
	muted        bool
	musicEnabled bool
	mu           sync.Mutex
}

func New() (*Manager, error) {
	context := audio.CurrentContext()
	if context == nil {
		context = audio.NewContext(sampleRate)
	}
	manager := &Manager{context: context, pcm: make(map[Effect][]byte), players: make(map[Effect][]*audio.Player), musicEnabled: true}
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
	musicWAV, err := files.ReadFile("audio/music-loop.wav")
	if err != nil {
		return nil, fmt.Errorf("read music-loop.wav: %w", err)
	}
	musicPCM, err := decodeWAV(musicWAV)
	if err != nil {
		return nil, fmt.Errorf("decode music-loop.wav: %w", err)
	}
	loop := audio.NewInfiniteLoop(bytes.NewReader(musicPCM), int64(len(musicPCM)))
	manager.music, err = context.NewPlayer(loop)
	if err != nil {
		return nil, fmt.Errorf("create music player: %w", err)
	}
	manager.music.SetVolume(.16)
	return manager, nil
}

func (m *Manager) Ready() bool { return m != nil && m.context.IsReady() }
func (m *Manager) Muted() bool { return m == nil || m.muted }
func (m *Manager) ToggleMute() bool {
	m.muted = !m.muted
	if m.muted && m.music != nil {
		m.music.Pause()
	}
	return m.muted
}

func (m *Manager) MusicEnabled() bool { return m != nil && m.musicEnabled }
func (m *Manager) MusicPlaying() bool { return m != nil && m.music != nil && m.music.IsPlaying() }
func (m *Manager) ToggleMusic() bool {
	m.musicEnabled = !m.musicEnabled
	if !m.musicEnabled && m.music != nil {
		m.music.Pause()
	}
	return m.musicEnabled
}

// Update starts music once the platform audio context becomes ready. Browsers
// normally reach this state after the first user interaction.
func (m *Manager) Update() {
	if m == nil || m.music == nil || m.muted || !m.musicEnabled || !m.context.IsReady() || m.music.IsPlaying() {
		return
	}
	m.music.Play()
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
	player.SetVolume(.36)
	m.players[effect] = append(pool, player)
	player.Play()
}
