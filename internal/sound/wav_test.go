package sound

import (
	"testing"
)

func TestEmbeddedEffectsDecode(t *testing.T) {
	for effect, filename := range filenames {
		data, err := files.ReadFile("audio/" + filename)
		if err != nil {
			t.Fatalf("%s: %v", effect, err)
		}
		pcm, err := decodeWAV(data)
		if err != nil {
			t.Fatalf("%s: %v", effect, err)
		}
		if len(pcm) < sampleRate/10 {
			t.Fatalf("%s is unexpectedly short: %d bytes", effect, len(pcm))
		}
	}
}

func TestMusicLoopDecodes(t *testing.T) {
	data, err := files.ReadFile("audio/music-loop.wav")
	if err != nil {
		t.Fatal(err)
	}
	pcm, err := decodeWAV(data)
	if err != nil {
		t.Fatal(err)
	}
	seconds := len(pcm) / (sampleRate * 4)
	if seconds < 30 || seconds > 40 {
		t.Fatalf("music loop duration = %d seconds", seconds)
	}
}

func TestEffectSetIsComplete(t *testing.T) {
	for _, effect := range []Effect{
		MenuFocus, MenuSelect, Join, Leave, Rotate, Lock, HardDrop,
		Line, FourLine, Pickup, Attack, Antidote, GameOver, Winner,
	} {
		if filenames[effect] == "" {
			t.Fatalf("missing filename for %s", effect)
		}
	}
}
