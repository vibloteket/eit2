// Command generate-audio creates project-owned prototype WAV effects.
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
	"lock.wav":      {{0, .10, 145, .34, "triangle"}, {.035, .07, 92, .20, "noise"}},
	"line.wav":      {{0, .11, 440, .25, "sine"}, {.09, .15, 660, .28, "sine"}},
	"four-line.wav": {{0, .10, 392, .24, "sine"}, {.08, .11, 523, .27, "sine"}, {.17, .20, 784, .30, "sine"}},
	"pickup.wav":    {{0, .09, 587, .23, "triangle"}, {.07, .14, 880, .28, "sine"}},
	"attack.wav":    {{0, .15, 170, .28, "square"}, {.11, .20, 120, .25, "triangle"}},
	"game-over.wav": {{0, .14, 330, .25, "triangle"}, {.12, .15, 247, .27, "triangle"}, {.25, .25, 165, .29, "triangle"}},
}

var musicPattern = []float64{261.63, 329.63, 392.00, 329.63, 293.66, 349.23, 440.00, 349.23, 246.94, 293.66, 392.00, 293.66, 220.00, 277.18, 329.63, 277.18}

func main() {
	output := "internal/sound/audio"
	if len(os.Args) == 2 {
		output = os.Args[1]
	}
	if err := os.MkdirAll(output, 0o755); err != nil {
		panic(err)
	}
	for name, notes := range effects {
		if err := writeWAV(filepath.Join(output, name), render(notes)); err != nil {
			panic(err)
		}
	}
	if err := writeWAV(filepath.Join(output, "music-loop.wav"), renderMusic()); err != nil {
		panic(err)
	}
}

func renderMusic() []int16 {
	const step = .25
	notes := make([]tone, 0, len(musicPattern)*2)
	for i, frequency := range musicPattern {
		start := float64(i) * step
		notes = append(notes, tone{start, step * .86, frequency, .12, "triangle"})
		if i%4 == 0 {
			notes = append(notes, tone{start, step * 3.7, frequency / 2, .055, "sine"})
		}
	}
	return render(notes)
}

func render(notes []tone) []int16 {
	end := 0.0
	for _, note := range notes {
		end = max(end, note.start+note.duration)
	}
	samples := make([]float64, int(end*sampleRate)+1)
	for _, note := range notes {
		start := int(note.start * sampleRate)
		length := int(note.duration * sampleRate)
		for i := 0; i < length; i++ {
			t := float64(i) / sampleRate
			phase := 2 * math.Pi * note.frequency * t
			value := math.Sin(phase)
			switch note.wave {
			case "triangle":
				value = 2 / math.Pi * math.Asin(math.Sin(phase))
			case "square":
				if value >= 0 {
					value = 1
				} else {
					value = -1
				}
			case "noise":
				value = math.Sin(float64(i * i * 7919))
			}
			attack := min(1, t/.008)
			release := math.Pow(1-float64(i)/float64(length), 2)
			samples[start+i] += value * note.volume * attack * release
		}
	}
	pcm := make([]int16, len(samples)*2)
	for i, sample := range samples {
		sample = max(-1, min(1, sample))
		value := int16(sample * 32767)
		pcm[i*2], pcm[i*2+1] = value, value
	}
	return pcm
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
