package audio

import (
	"encoding/binary"
	"math"
	"os"
	"testing"
)

// makeWAV creates a minimal valid WAV file with the given 16-bit PCM samples.
func makeWAV(t *testing.T, samples []int16) string {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "test-*.wav")
	if err != nil {
		t.Fatal(err)
	}

	numSamples := len(samples)
	dataSize := numSamples * 2
	fileSize := 36 + dataSize

	// RIFF header
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(fileSize))
	f.Write([]byte("WAVE"))

	// fmt chunk
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16)) // chunk size
	binary.Write(f, binary.LittleEndian, uint16(1))   // PCM format
	binary.Write(f, binary.LittleEndian, uint16(1))   // mono
	binary.Write(f, binary.LittleEndian, uint32(16000)) // sample rate
	binary.Write(f, binary.LittleEndian, uint32(32000)) // byte rate
	binary.Write(f, binary.LittleEndian, uint16(2))   // block align
	binary.Write(f, binary.LittleEndian, uint16(16))  // bits per sample

	// data chunk
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataSize))
	for _, s := range samples {
		binary.Write(f, binary.LittleEndian, s)
	}

	f.Close()
	return f.Name()
}

func TestReadWAVSamples(t *testing.T) {
	input := []int16{0, 16383, -16384, 32767, -32768}
	path := makeWAV(t, input)

	samples, err := ReadWAVSamples(path)
	if err != nil {
		t.Fatalf("ReadWAVSamples: %v", err)
	}

	if len(samples) != len(input) {
		t.Fatalf("expected %d samples, got %d", len(input), len(samples))
	}

	for i, s := range input {
		expected := float32(s) / float32(math.MaxInt16)
		if diff := samples[i] - expected; diff > 1e-6 || diff < -1e-6 {
			t.Errorf("sample %d: expected %f, got %f", i, expected, samples[i])
		}
	}
}

func TestReadWAVSamples_Silence(t *testing.T) {
	input := make([]int16, 16000) // 1 second of silence
	path := makeWAV(t, input)

	samples, err := ReadWAVSamples(path)
	if err != nil {
		t.Fatalf("ReadWAVSamples: %v", err)
	}

	if len(samples) != 16000 {
		t.Fatalf("expected 16000 samples, got %d", len(samples))
	}

	for i, s := range samples {
		if s != 0 {
			t.Errorf("sample %d: expected 0, got %f", i, s)
			break
		}
	}
}

func TestReadWAVSamples_InvalidFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "bad-*.wav")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("not a wav file"))
	f.Close()

	_, err = ReadWAVSamples(f.Name())
	if err == nil {
		t.Fatal("expected error for invalid WAV file")
	}
}

func TestReadWAVSamples_TooShort(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "short-*.wav")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("RIFF"))
	f.Close()

	_, err = ReadWAVSamples(f.Name())
	if err == nil {
		t.Fatal("expected error for too-short WAV file")
	}
}

func TestReadWAVSamples_NoDataChunk(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "nodata-*.wav")
	if err != nil {
		t.Fatal(err)
	}

	// Valid RIFF/WAVE header but only a fmt chunk, no data chunk
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(28))
	f.Write([]byte("WAVE"))
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))
	f.Write(make([]byte, 16))
	f.Close()

	_, err = ReadWAVSamples(f.Name())
	if err == nil {
		t.Fatal("expected error for WAV with no data chunk")
	}
}

func TestExtract_MissingFile(t *testing.T) {
	_, err := Extract("/nonexistent/file.mkv")
	if err == nil {
		t.Fatal("expected error for missing input file")
	}
}
