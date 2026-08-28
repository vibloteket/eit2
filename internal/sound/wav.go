package sound

import (
	"encoding/binary"
	"errors"
)

// decodeWAV validates our generated PCM WAV and returns headerless stereo PCM.
func decodeWAV(data []byte) ([]byte, error) {
	if len(data) < 44 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" || string(data[12:16]) != "fmt " {
		return nil, errors.New("invalid WAV header")
	}
	if binary.LittleEndian.Uint16(data[20:22]) != 1 || binary.LittleEndian.Uint16(data[22:24]) != 2 || binary.LittleEndian.Uint32(data[24:28]) != sampleRate || binary.LittleEndian.Uint16(data[34:36]) != 16 {
		return nil, errors.New("expected 44.1 kHz 16-bit stereo PCM")
	}
	for offset := 12; offset+8 <= len(data); {
		name := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		start, end := offset+8, offset+8+size
		if end > len(data) {
			return nil, errors.New("truncated WAV chunk")
		}
		if name == "data" {
			return append([]byte(nil), data[start:end]...), nil
		}
		offset = end + size%2
	}
	return nil, errors.New("WAV data chunk missing")
}
