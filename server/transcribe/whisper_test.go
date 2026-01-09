package transcribe

import (
	"reflect"
	"testing"
)

func TestNewWhisperEngine(t *testing.T) {
	w := NewWhisperEngine("/usr/bin/whisper-cli", "/models/tiny.bin", 4)

	if w.cli != "/usr/bin/whisper-cli" {
		t.Errorf("cli = %q, want /usr/bin/whisper-cli", w.cli)
	}
	if w.model != "/models/tiny.bin" {
		t.Errorf("model = %q, want /models/tiny.bin", w.model)
	}
	if w.threads != 4 {
		t.Errorf("threads = %d, want 4", w.threads)
	}
}

func TestNewWhisperEngine_DefaultThreads(t *testing.T) {
	w := NewWhisperEngine("/usr/bin/whisper-cli", "/models/tiny.bin", 0)
	if w.threads != 4 {
		t.Errorf("threads = %d, want 4 (default)", w.threads)
	}

	w = NewWhisperEngine("/usr/bin/whisper-cli", "/models/tiny.bin", -1)
	if w.threads != 4 {
		t.Errorf("threads = %d, want 4 (default)", w.threads)
	}
}

func TestWhisperEngine_SetPrompt(t *testing.T) {
	w := NewWhisperEngine("/usr/bin/whisper-cli", "/models/tiny.bin", 4)

	if w.Prompt() != "" {
		t.Errorf("initial prompt = %q, want empty", w.Prompt())
	}

	w.SetPrompt("claude code, terminal commands")
	if w.Prompt() != "claude code, terminal commands" {
		t.Errorf("prompt = %q, want 'claude code, terminal commands'", w.Prompt())
	}
}

func TestWhisperEngine_BuildArgs(t *testing.T) {
	w := NewWhisperEngine("/usr/bin/whisper-cli", "/models/tiny.bin", 4)

	// without prompt
	args := w.buildArgs("/tmp/audio.wav", true)
	expected := []string{
		"-m", "/models/tiny.bin",
		"-f", "/tmp/audio.wav",
		"-t", "4",
		"-nt",
		"-np",
	}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("buildArgs (no prompt) = %v, want %v", args, expected)
	}

	// with prompt
	w.SetPrompt("test prompt")
	args = w.buildArgs("/tmp/audio.wav", true)
	expected = []string{
		"-m", "/models/tiny.bin",
		"-f", "/tmp/audio.wav",
		"-t", "4",
		"-nt",
		"-np",
		"--prompt", "test prompt",
	}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("buildArgs (with prompt) = %v, want %v", args, expected)
	}

	// with prompt but usePrompt=false
	args = w.buildArgs("/tmp/audio.wav", false)
	expected = []string{
		"-m", "/models/tiny.bin",
		"-f", "/tmp/audio.wav",
		"-t", "4",
		"-nt",
		"-np",
	}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("buildArgs (usePrompt=false) = %v, want %v", args, expected)
	}
}

func TestWhisperEngine_BuildArgsThreadCount(t *testing.T) {
	w := NewWhisperEngine("/usr/bin/whisper-cli", "/models/tiny.bin", 8)

	args := w.buildArgs("/tmp/audio.wav", false)
	// find thread count
	var threadCount string
	for i, arg := range args {
		if arg == "-t" && i+1 < len(args) {
			threadCount = args[i+1]
			break
		}
	}

	if threadCount != "8" {
		t.Errorf("thread count in args = %q, want '8'", threadCount)
	}
}
