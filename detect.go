package main

import (
	"fmt"
	"runtime"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go"
)

// detectLanguage loads a whisper model and detects the spoken language in a WAV file.
// Returns the language code, confidence score, and any error.
func detectLanguage(modelPath, wavPath string) (string, float64, error) {
	ctx := whisper.Whisper_init(modelPath)
	if ctx == nil {
		return "", 0, fmt.Errorf("failed to load model: %s", modelPath)
	}
	defer ctx.Whisper_free()

	samples, err := readWAVSamples(wavPath)
	if err != nil {
		return "", 0, fmt.Errorf("reading audio: %w", err)
	}

	nThreads := runtime.NumCPU()

	if err := ctx.Whisper_pcm_to_mel(samples, nThreads); err != nil {
		return "", 0, fmt.Errorf("computing mel spectrogram: %w", err)
	}

	probs, err := ctx.Whisper_lang_auto_detect(0, nThreads)
	if err != nil {
		return "", 0, fmt.Errorf("detecting language: %w", err)
	}

	bestIdx := 0
	for i, p := range probs {
		if p > probs[bestIdx] {
			bestIdx = i
		}
	}

	lang := whisper.Whisper_lang_str(bestIdx)
	confidence := float64(probs[bestIdx])

	return lang, confidence, nil
}
