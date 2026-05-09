package detect

/*
#include <stdlib.h>
#include <whisper.h>
#include <ggml.h>

// ggml_backend_load_all is in libggml but not in ggml.h
extern void ggml_backend_load_all(void);

static void whisper_log_noop(enum ggml_log_level level, const char * text, void * user_data) {
	(void)level; (void)text; (void)user_data;
}

static struct whisper_context* init_no_gpu(const char* path) {
	whisper_log_set(whisper_log_noop, NULL);
	ggml_backend_load_all();
	struct whisper_context_params params = whisper_context_default_params();
	params.use_gpu = false;
	return whisper_init_from_file_with_params(path, params);
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/stahnma/speech-check/internal/audio"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go"
)

// Language loads a whisper model and detects the spoken language in a WAV file.
// Returns the language code, confidence score, and any error.
func Language(modelPath, wavPath string) (string, float64, error) {
	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))

	cCtx := C.init_no_gpu(cPath)
	if cCtx == nil {
		return "", 0, fmt.Errorf("failed to load model: %s", modelPath)
	}
	ctx := (*whisper.Context)(unsafe.Pointer(cCtx))
	defer ctx.Whisper_free()

	samples, err := audio.ReadWAVSamples(wavPath)
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
