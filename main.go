package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	modelPath := flag.String("model", "", "path to whisper ggml model file")
	flag.StringVar(modelPath, "m", "", "path to whisper ggml model file (shorthand)")
	jsonOutput := flag.Bool("json", false, "output as JSON")
	logFile := flag.String("log", "", "path to JSONL log file (appends results)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: speech-check [flags] <media-file>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputFile := flag.Arg(0)

	if *modelPath == "" {
		fmt.Fprintln(os.Stderr, "error: --model is required")
		os.Exit(1)
	}

	if _, err := os.Stat(inputFile); err != nil {
		fmt.Fprintf(os.Stderr, "cannot open file: %s\n", inputFile)
		os.Exit(1)
	}

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Fprintln(os.Stderr, "ffmpeg not found on PATH — install ffmpeg to use speech-check")
		os.Exit(1)
	}

	wavPath, err := extractAudio(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract audio: %s\n", err)
		os.Exit(1)
	}
	defer os.Remove(wavPath)

	lang, confidence, err := detectLanguage(*modelPath, wavPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect language: %s\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		out, _ := json.Marshal(map[string]any{
			"language":   lang,
			"confidence": confidence,
		})
		fmt.Println(string(out))
	} else {
		fmt.Printf("%s %.4f\n", lang, confidence)
	}

	if *logFile != "" {
		entry, _ := json.Marshal(map[string]any{
			"file":       inputFile,
			"language":   lang,
			"confidence": confidence,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
		f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write log: %s\n", err)
			os.Exit(1)
		}
		defer f.Close()
		fmt.Fprintln(f, string(entry))
	}
}
