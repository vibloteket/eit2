package sound

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
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

func TestGameplayLoopIsSelected126BPMVersion(t *testing.T) {
	data, err := files.ReadFile("audio/gameplay-beethoven.wav")
	if err != nil {
		t.Fatal(err)
	}
	pcm, err := decodeWAV(data)
	if err != nil {
		t.Fatal(err)
	}
	const selectedHash = "52285ceff9158e6fb683af7c25d58c898344e5a1c4f6532e7dcd0a8e9ea3470a"
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != selectedHash {
		t.Fatalf("gameplay asset is not the selected 126 BPM loop: %s", got)
	}
	if len(pcm) != 4704000*4 {
		t.Fatalf("gameplay PCM length = %d bytes, want %d", len(pcm), 4704000*4)
	}
	for channel := 0; channel < 2; channel++ {
		first := int(int16(binary.LittleEndian.Uint16(pcm[channel*2:])))
		last := int(int16(binary.LittleEndian.Uint16(pcm[len(pcm)-4+channel*2:])))
		if delta := first - last; delta < -328 || delta > 328 {
			t.Fatalf("gameplay channel %d seam delta = %d", channel, delta)
		}
	}
}

func TestLobbyLoopIsSelectedFullCorrectedVersion(t *testing.T) {
	if musicFilenames[MatchMusic] != "gameplay-beethoven.wav" || musicFilenames[LobbyMusic] != "lobby-badinerie.wav" {
		t.Fatal("lobby and match must use separate assets")
	}
	data, err := files.ReadFile("audio/" + musicFilenames[LobbyMusic])
	if err != nil {
		t.Fatal(err)
	}
	const selectedHash = "d41d29242ea46c8df0014eb84910dc72ced90f1ecd7149d06346e943d95361b5"
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != selectedHash {
		t.Fatalf("lobby asset is not the selected full corrected loop: %s", got)
	}
	pcm, err := decodeWAV(data)
	if err != nil {
		t.Fatal(err)
	}
	// AABB: 160 quarter-note beats at 108 BPM. Do not embed the listening
	// version with its ending silence, or the superseded 18-second excerpt.
	if len(pcm) != 3920000*4 {
		t.Fatalf("lobby PCM length = %d bytes, want %d", len(pcm), 3920000*4)
	}
	for channel := 0; channel < 2; channel++ {
		first := int(int16(binary.LittleEndian.Uint16(pcm[channel*2:])))
		last := int(int16(binary.LittleEndian.Uint16(pcm[len(pcm)-4+channel*2:])))
		delta := first - last
		if delta < -2 || delta > 2 {
			t.Fatalf("lobby loop channel %d seam delta = %d", channel, delta)
		}
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

func TestAllAudioMatchesSourceAudit(t *testing.T) {
	data, err := os.ReadFile("../../music/procedural-audio-audit.json")
	if err != nil {
		t.Fatal(err)
	}
	var audit struct {
		GeneratorSHA256 string `json:"generatorSHA256"`
		Files           []struct {
			File   string `json:"file"`
			SHA256 string `json:"sha256"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &audit); err != nil {
		t.Fatal(err)
	}
	generator, err := os.ReadFile("../../scripts/generate-audio/main.go")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(generator)) != audit.GeneratorSHA256 {
		t.Fatal("procedural generator changed; reproduce its audio and update the source audit")
	}
	known := map[string]bool{"lobby-badinerie.wav": true, "gameplay-beethoven.wav": true}
	if len(audit.Files) != 14 {
		t.Fatalf("procedural audit contains %d files, want 14", len(audit.Files))
	}
	for _, entry := range audit.Files {
		b, err := files.ReadFile("audio/" + entry.File)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != entry.SHA256 {
			t.Fatalf("%s differs from the reproduced/audited asset", entry.File)
		}
		known[entry.File] = true
	}
	entries, err := files.ReadDir("audio")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !known[entry.Name()] {
			t.Fatalf("embedded audio lacks a source audit: %s", entry.Name())
		}
	}
}

func TestGameplayPlaybackLeavesRoomForImportantEffects(t *testing.T) {
	levels := func(filename string) (peak, rms float64) {
		t.Helper()
		data, err := files.ReadFile("audio/" + filename)
		if err != nil {
			t.Fatal(err)
		}
		pcm, err := decodeWAV(data)
		if err != nil {
			t.Fatal(err)
		}
		var sum float64
		for i := 0; i < len(pcm); i += 2 {
			x := float64(int16(binary.LittleEndian.Uint16(pcm[i:]))) / 32768
			peak = math.Max(peak, math.Abs(x))
			sum += x * x
		}
		return peak, math.Sqrt(sum / float64(len(pcm)/2))
	}
	if musicVolumes[LobbyMusic] != .16 || musicVolumes[MatchMusic] != .08 || effectVolume != .36 {
		t.Fatal("unexpected playback balance")
	}
	musicPeak, musicRMS := levels(musicFilenames[MatchMusic])
	for _, effect := range []Effect{HardDrop, Line, FourLine, Attack} {
		peak, rms := levels(filenames[effect])
		marginDB := 20 * math.Log10(rms*effectVolume/(musicRMS*musicVolumes[MatchMusic]))
		if marginDB < 4 {
			t.Errorf("%s RMS margin over music = %.2f dB, want at least 4 dB", effect, marginDB)
		}
		if musicPeak*musicVolumes[MatchMusic]+peak*effectVolume >= 1 {
			t.Errorf("%s plus music can clip even without other voices", effect)
		}
		t.Logf("%s: RMS margin %.2f dB (signal measure, not a perceptual listening test)", effect, marginDB)
	}
}
