package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
)

// extractAudio shells out to ffmpeg to convert a media file to 16kHz mono WAV.
// Returns the path to the temporary WAV file.
func extractAudio(inputPath string) (string, error) {
	tmpFile, err := os.CreateTemp("", "speech-check-*.wav")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-ar", "16000",
		"-ac", "1",
		"-f", "wav",
		"-y",
		tmpFile.Name(),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("%s: %s", err, string(output))
	}

	return tmpFile.Name(), nil
}

// readWAVSamples reads a 16kHz mono WAV file and returns float32 samples
// normalized to [-1.0, 1.0].
func readWAVSamples(path string) ([]float32, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Skip the 44-byte WAV header
	if len(data) < 44 {
		return nil, fmt.Errorf("WAV file too short")
	}

	// Verify RIFF header
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a valid WAV file")
	}

	// Find the data chunk
	offset := 12
	for offset+8 < len(data) {
		chunkID := string(data[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		if chunkID == "data" {
			pcmData := data[offset+8 : offset+8+chunkSize]
			samples := make([]float32, len(pcmData)/2)
			for i := 0; i < len(pcmData)-1; i += 2 {
				sample := int16(binary.LittleEndian.Uint16(pcmData[i : i+2]))
				samples[i/2] = float32(sample) / float32(math.MaxInt16)
			}
			return samples, nil
		}
		offset += 8 + chunkSize
		// Chunks are word-aligned
		if chunkSize%2 != 0 {
			offset++
		}
	}

	return nil, fmt.Errorf("no data chunk found in WAV file")
}
