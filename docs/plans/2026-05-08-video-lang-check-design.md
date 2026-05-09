# video-lang-check Design

A CLI tool that detects the spoken language in a media file.

## CLI Interface

```
video-lang-check [flags] <media-file>

Flags:
  --model, -m    Path to whisper ggml model file (required)
  --json         Output as JSON
  --log          Path to JSONL log file (appends results)
  --help, -h     Show help
```

## Workflow

1. Validate the input file exists
2. Check that ffmpeg is available on PATH
3. Shell out to ffmpeg to extract audio as 16kHz mono WAV to a temp file
4. Load the whisper model and run language detection on the audio
5. Print the result and clean up the temp file
6. If `--log` is specified, append a JSONL entry to the log file

## Output

Default:

```
en 0.9432
```

With `--json`:

```json
{"language": "en", "confidence": 0.9432}
```

JSONL log entry (via `--log`):

```json
{"file": "video.mp4", "language": "en", "confidence": 0.9432, "timestamp": "2026-05-08T22:15:00Z"}
```

## Project Structure

```
video-lang-check/
├── main.go          # CLI entry point, flag parsing, orchestration
├── audio.go         # ffmpeg audio extraction to temp WAV
├── detect.go        # whisper model loading + language detection
├── go.mod
├── go.sum
├── Makefile
└── docs/plans/
```

## Dependencies

- `github.com/ggerganov/whisper.cpp/bindings/go` — whisper Go bindings (CGo)
- Standard library only for everything else
- Runtime: ffmpeg on PATH

## Build

- CGo required (`CGO_ENABLED=1`, `CC=clang`)
- Makefile targets: `build` (default), `clean`, `deps`

## Audio Extraction (audio.go)

- Run: `ffmpeg -i <input> -ar 16000 -ac 1 -f wav <tempfile>`
- Temp file via `os.CreateTemp`, cleaned up with `defer`
- Whisper requires 16kHz mono PCM

## Language Detection (detect.go)

- Load model via `whisper.New(modelPath)`
- Create context, feed audio samples
- Use language detection API to get language probabilities
- Return top result (language code + confidence)

## Error Handling

- Missing ffmpeg: `"ffmpeg not found on PATH — install ffmpeg to use video-lang-check"`
- Bad input file: `"cannot open file: <path>"`
- ffmpeg failure: `"failed to extract audio: <stderr output>"`
- Model load failure: `"failed to load model: <path>"`
- Exit code 1 for all errors, messages to stderr

## Future Considerations

- Embed a small whisper model in the binary at build time for a fully self-contained tool
