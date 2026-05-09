package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/stahnma/video-lang-check/internal/audio"
	"github.com/stahnma/video-lang-check/internal/detect"
)

func main() {
	modelPath := flag.String("model", "", "path to whisper ggml model file")
	flag.StringVar(modelPath, "m", "", "path to whisper ggml model file (shorthand)")
	jsonOutput := flag.Bool("json", false, "output as JSON")
	logFile := flag.String("log", "", "path to JSONL log file (appends results)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: video-lang-check [flags] <media-file>...\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	if *modelPath == "" {
		fmt.Fprintln(os.Stderr, "error: --model is required")
		os.Exit(1)
	}

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Fprintln(os.Stderr, "ffmpeg not found on PATH — install ffmpeg to use video-lang-check")
		os.Exit(1)
	}

	// Expand globs in arguments
	var files []string
	for _, arg := range flag.Args() {
		matches, err := filepath.Glob(arg)
		if err != nil || len(matches) == 0 {
			files = append(files, arg)
		} else {
			files = append(files, matches...)
		}
	}

	var logFd *os.File
	if *logFile != "" {
		var err error
		logFd, err = os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file: %s\n", err)
			os.Exit(1)
		}
		defer logFd.Close()
	}

	multiFile := len(files) > 1
	hasError := false

	for _, inputFile := range files {
		if _, err := os.Stat(inputFile); err != nil {
			fmt.Fprintf(os.Stderr, "cannot open file: %s\n", inputFile)
			hasError = true
			continue
		}

		wavPath, err := audio.Extract(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to extract audio from %s: %s\n", inputFile, err)
			hasError = true
			continue
		}

		lang, confidence, err := detect.Language(*modelPath, wavPath)
		os.Remove(wavPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to detect language for %s: %s\n", inputFile, err)
			hasError = true
			continue
		}

		if *jsonOutput {
			out, _ := json.Marshal(map[string]any{
				"file":       inputFile,
				"language":   lang,
				"confidence": confidence,
			})
			fmt.Println(string(out))
		} else if multiFile {
			fmt.Printf("%s\t%s %.4f\n", inputFile, lang, confidence)
		} else {
			fmt.Printf("%s %.4f\n", lang, confidence)
		}

		if logFd != nil {
			entry, _ := json.Marshal(map[string]any{
				"file":       inputFile,
				"language":   lang,
				"confidence": confidence,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			})
			fmt.Fprintln(logFd, string(entry))
		}
	}

	if hasError {
		os.Exit(1)
	}
}
