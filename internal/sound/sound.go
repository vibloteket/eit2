// Package sound maps semantic game events to short embedded prototype effects.
package sound

import (
	"embed"
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100

type Effect string

const (
	Lock     Effect = "lock"
	Line     Effect = "line"
	FourLine Effect = "four-line"
	Pickup   Effect = "pickup"
	Attack   Effect = "attack"
	GameOver Effect = "game-over"
)

//go:embed audio/*.wav
var files embed.FS

var filenames = map[Effect]string{
	Lock: "lock.wav", Line: "line.wav", FourLine: "four-line.wav",
	Pickup: "pickup.wav", Attack: "attack.wav", GameOver: "game-over.wav",
}

type Manager struct {
	context *audio.Context
	pcm     map[Effect][]byte
	players map[Effect][]*audio.Player
	muted   bool
	mu      sync.Mutex
}

func New() (*Manager, error) {
	context := audio.CurrentContext()
	if context == nil {
		context = audio.NewContext(sampleRate)
	}
	manager := &Manager{context: context, pcm: make(map[Effect][]byte), players: make(map[Effect][]*audio.Player)}
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
	return manager, nil
}

func (m *Manager) Ready() bool      { return m != nil && m.context.IsReady() }
func (m *Manager) Muted() bool      { return m == nil || m.muted }
func (m *Manager) ToggleMute() bool { m.muted = !m.muted; return m.muted }

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
