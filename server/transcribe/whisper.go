package transcribe

import (
	"context"
	"os/exec"
	"strconv"
)

// WhisperEngine handles transcription via whisper-cli subprocess
type WhisperEngine struct {
	cli     string // path to whisper-cli
	model   string // path to model file
	threads int    // thread count
	prompt  string // vocabulary prompt
}

// NewWhisperEngine creates a new WhisperEngine
func NewWhisperEngine(cli, model string, threads int) *WhisperEngine {
	if threads <= 0 {
		threads = 4
	}
	return &WhisperEngine{
		cli:     cli,
		model:   model,
		threads: threads,
	}
}

// SetPrompt sets the vocabulary prompt for better recognition
func (w *WhisperEngine) SetPrompt(prompt string) {
	w.prompt = prompt
}

// Prompt returns the current vocabulary prompt
func (w *WhisperEngine) Prompt() string {
	return w.prompt
}

// buildArgs constructs command-line arguments for whisper-cli
func (w *WhisperEngine) buildArgs(audioPath string, usePrompt bool) []string {
	args := []string{
		"-m", w.model,
		"-f", audioPath,
		"-t", strconv.Itoa(w.threads),
		"-nt", // no timestamps
		"-np", // no progress
	}
	if usePrompt && w.prompt != "" {
		args = append(args, "--prompt", w.prompt)
	}
	return args
}

// Transcribe runs whisper-cli with the configured prompt
func (w *WhisperEngine) Transcribe(ctx context.Context, audioPath string) (string, error) {
	args := w.buildArgs(audioPath, true)

	cmd := exec.CommandContext(ctx, w.cli, args...)
	output, err := cmd.Output() // stderr is ignored
	if err != nil {
		return "", err
	}

	return FilterTranscription(string(output)), nil
}

// QuickTranscribe runs whisper-cli without prompt (faster for command detection)
func (w *WhisperEngine) QuickTranscribe(ctx context.Context, audioPath string) (string, error) {
	args := w.buildArgs(audioPath, false)

	cmd := exec.CommandContext(ctx, w.cli, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return FilterTranscription(string(output)), nil
}
