package sounds

import (
	"runtime"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Enabled {
		t.Error("default config should have Enabled=true")
	}

	if cfg.Ready == "" {
		t.Error("default config should have Ready sound path")
	}
	if cfg.CommandSuccess == "" {
		t.Error("default config should have CommandSuccess sound path")
	}
	if cfg.CommandWarning == "" {
		t.Error("default config should have CommandWarning sound path")
	}
	if cfg.DictationToggle == "" {
		t.Error("default config should have DictationToggle sound path")
	}
}

func TestDefaultConfig_PlatformSpecific(t *testing.T) {
	cfg := DefaultConfig()

	if runtime.GOOS == "darwin" {
		if cfg.Ready != "/System/Library/Sounds/Tink.aiff" {
			t.Errorf("macOS ready sound = %q, want /System/Library/Sounds/Tink.aiff", cfg.Ready)
		}
	} else {
		if cfg.Ready != "/usr/share/sounds/freedesktop/stereo/message.oga" {
			t.Errorf("Linux ready sound = %q, want freedesktop path", cfg.Ready)
		}
	}
}

func TestNewPlayer(t *testing.T) {
	cfg := Config{
		Enabled:         true,
		Ready:           "/test/ready.wav",
		CommandSuccess:  "/test/success.wav",
		CommandWarning:  "/test/warning.wav",
		DictationToggle: "/test/toggle.wav",
	}

	p := NewPlayer(cfg)
	if p == nil {
		t.Fatal("NewPlayer returned nil")
	}
	if p.config.Ready != "/test/ready.wav" {
		t.Errorf("player config.Ready = %q, want /test/ready.wav", p.config.Ready)
	}
}

func TestPlayer_DisabledNoOp(t *testing.T) {
	cfg := Config{
		Enabled: false,
		Ready:   "/nonexistent/sound.wav",
	}

	p := NewPlayer(cfg)

	// these should not panic or error even with invalid paths
	// because sounds are disabled
	p.PlayReady()
	p.PlayCommandSuccess()
	p.PlayCommandWarning()
	p.PlayDictationToggle()
}

func TestPlayer_EmptyPathNoOp(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Ready:   "", // empty path
	}

	p := NewPlayer(cfg)

	// should not panic with empty path
	p.PlayReady()
}
