# speech-check

A CLI tool that detects the spoken language in media files using [whisper.cpp](https://github.com/ggml-org/whisper.cpp).

## Usage

```
speech-check [flags] <media-file>
```

### Flags

| Flag | Description |
|------|-------------|
| `--model`, `-m` | Path to whisper ggml model file (required) |
| `--json` | Output as JSON |
| `--log <file>` | Append results as JSONL to the specified file |

### Examples

```bash
# Basic usage
speech-check -m ggml-base.bin video.mkv
# en 0.9346

# JSON output
speech-check -m ggml-base.bin --json video.mkv
# {"confidence":0.9346,"language":"en"}

# Batch with JSONL logging
for f in *.mkv; do
  speech-check -m ggml-base.bin --log results.jsonl "$f"
done
```

## Requirements

- [ffmpeg](https://ffmpeg.org/) on `PATH` (for audio extraction from media files)
- A whisper ggml model file (e.g., `ggml-base.bin`)

### Downloading a model

```bash
curl -L -o ggml-base.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin
```

Larger models (e.g., `ggml-small.bin`, `ggml-medium.bin`) may give better accuracy.

## Building

Requires Go, a C compiler (clang), and the whisper.cpp libraries (available via [Flox](https://flox.dev/)):

```bash
make build
```

The build uses the whisper.cpp Go bindings with CGo. The Makefile handles linking against the flox-provided whisper.cpp libraries and embeds the library rpath in the binary.

## How it works

1. Extracts audio from the input file using ffmpeg (16kHz mono WAV)
2. Loads the whisper model and computes the mel spectrogram
3. Runs whisper's language auto-detection (no full transcription)
4. Reports the detected language code and confidence score
