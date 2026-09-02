// Command generate-audio creates project-owned Doodle Party WAV audio.
package main

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
)

const sampleRate = 44100

type tone struct {
	start, duration float64
	frequency       float64
	volume          float64
	wave            string
}

var effects = map[string][]tone{
	"menu-focus.wav":  {{0, .055, 740, .13, "wood"}},
	"menu-select.wav": {{0, .07, 520, .18, "wood"}, {.045, .10, 780, .13, "mallet"}},
	"join.wav":        {{0, .08, 392, .17, "mallet"}, {.07, .10, 523, .18, "mallet"}, {.15, .15, 659, .16, "mallet"}},
	"leave.wav":       {{0, .08, 523, .14, "mallet"}, {.07, .13, 330, .15, "wood"}},
	"rotate.wav":      {{0, .055, 285, .11, "wood"}, {.025, .045, 390, .07, "noise"}},
	"lock.wav":        {{0, .10, 150, .29, "wood"}, {.025, .08, 92, .16, "noise"}},
	"hard-drop.wav":   {{0, .13, 95, .32, "wood"}, {.015, .09, 58, .23, "sine"}, {.045, .10, 130, .12, "noise"}},
	"line.wav":        {{0, .09, 392, .16, "mallet"}, {.075, .10, 523, .18, "mallet"}, {.15, .13, 659, .17, "mallet"}},
	"four-line.wav":   {{0, .09, 330, .17, "mallet"}, {.07, .09, 440, .18, "mallet"}, {.14, .10, 554, .20, "mallet"}, {.22, .20, 880, .18, "bell"}},
	"pickup.wav":      {{0, .08, 587, .16, "wood"}, {.055, .12, 880, .17, "bell"}},
	"attack.wav":      {{0, .12, 145, .25, "wood"}, {.035, .18, 105, .16, "noise"}, {.10, .15, 190, .11, "square"}},
	"antidote.wav":    {{0, .06, 230, .17, "pop"}, {.055, .15, 494, .14, "mallet"}, {.14, .22, 740, .16, "bell"}},
	"game-over.wav":   {{0, .13, 392, .16, "mallet"}, {.12, .15, 311, .16, "mallet"}, {.26, .28, 196, .17, "wood"}},
	"winner.wav":      {{0, .10, 392, .18, "mallet"}, {.09, .10, 494, .18, "mallet"}, {.18, .11, 587, .20, "mallet"}, {.28, .35, 784, .18, "bell"}},
}

var melody = []float64{392, 0, 494, 587, 659, 587, 494, 440, 392, 494, 0, 587, 523, 440, 392, 330}
var harmony = []float64{196, 220, 165, 196}

func main() {
	output := "internal/sound/audio"
	if len(os.Args) == 2 {
		output = os.Args[1]
	}
	if err := os.MkdirAll(output, 0o755); err != nil {
		panic(err)
	}
	for name, notes := range effects {
		if err := writeWAV(filepath.Join(output, name), render(notes, 0)); err != nil {
			panic(err)
		}
	}
	if err := writeWAV(filepath.Join(output, "music-loop.wav"), renderMusic()); err != nil {
		panic(err)
	}
}

func renderMusic() []int16 {
	const bpm = 108.0
	beat := 60 / bpm
	bars := 16
	duration := float64(bars) * 4 * beat
	notes := make([]tone, 0, 220)
	for beatIndex := 0; beatIndex < bars*4; beatIndex++ {
		start := float64(beatIndex) * beat
		bar := beatIndex / 4
		step := beatIndex % len(melody)
		if frequency := melody[step]; frequency > 0 {
			if bar >= 8 && bar%2 == 1 {
				frequency *= 2
			}
			notes = append(notes, tone{start, beat * .62, frequency, .075, "mallet"})
		}
		if beatIndex%4 == 0 {
			root := harmony[bar%len(harmony)]
			notes = append(notes, tone{start, beat * 3.7, root, .035, "sine"})
		}
		// Dry tabletop pulse and shaker leave room for gameplay sounds.
		notes = append(notes, tone{start, .055, 105, .035, "wood"})
		if beatIndex%2 == 1 {
			notes = append(notes, tone{start, .065, 1800, .018, "noise"})
		}
	}
	return render(notes, duration)
}

func render(notes []tone, fixedDuration float64) []int16 {
	end := fixedDuration
	for _, note := range notes {
		end = max(end, note.start+note.duration)
	}
	samples := make([]float64, int(end*sampleRate))
	for noteIndex, note := range notes {
		start := int(note.start * sampleRate)
		length := int(note.duration * sampleRate)
		for i := 0; i < length && start+i < len(samples); i++ {
			t := float64(i) / sampleRate
			phase := 2 * math.Pi * note.frequency * t
			value := wave(note.wave, phase, i, noteIndex)
			attack := min(1, t/.004)
			release := math.Pow(1-float64(i)/float64(length), 2.25)
			samples[start+i] += value * note.volume * attack * release
		}
	}
	pcm := make([]int16, len(samples)*2)
	for i, sample := range samples {
		sample = math.Tanh(sample * 1.15)
		value := int16(max(-1, min(1, sample)) * 32767)
		pcm[i*2], pcm[i*2+1] = value, value
	}
	return pcm
}

func wave(kind string, phase float64, sample, seed int) float64 {
	sine := math.Sin(phase)
	switch kind {
	case "mallet":
		return sine + .32*math.Sin(phase*2.01) + .13*math.Sin(phase*3.99)
	case "wood":
		return .55*sine + .35*math.Sin(phase*2.7) + .18*noise(sample, seed)
	case "bell":
		return .7*sine + .24*math.Sin(phase*2.41) + .12*math.Sin(phase*4.08)
	case "square":
		if sine >= 0 {
			return 1
		}
		return -1
	case "noise":
		return noise(sample, seed)
	case "pop":
		return math.Sin(phase * (1 + .8*math.Exp(-float64(sample)/900)))
	default:
		return sine
	}
}

func noise(sample, seed int) float64 {
	x := uint32(sample*1664525 + seed*1013904223 + 12345)
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	return float64(int32(x)) / float64(math.MaxInt32)
}

func writeWAV(path string, samples []int16) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	dataSize := uint32(len(samples) * 2)
	file.WriteString("RIFF")
	binary.Write(file, binary.LittleEndian, uint32(36)+dataSize)
	file.WriteString("WAVEfmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))
	binary.Write(file, binary.LittleEndian, uint16(1))
	binary.Write(file, binary.LittleEndian, uint16(2))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate*4))
	binary.Write(file, binary.LittleEndian, uint16(4))
	binary.Write(file, binary.LittleEndian, uint16(16))
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, dataSize)
	return binary.Write(file, binary.LittleEndian, samples)
}
