package sounds

import (
	"os/exec"
	"runtime"
)

// platform-specific default sound paths
var defaultSoundsLinux = map[string]string{
	"ready":           "/usr/share/sounds/freedesktop/stereo/message.oga",
	"command_success": "/usr/share/sounds/freedesktop/stereo/complete.oga",
	"command_warning": "/usr/share/sounds/freedesktop/stereo/dialog-warning.oga",
	"dictation_toggle": "/usr/share/sounds/freedesktop/stereo/service-login.oga",
}

var defaultSoundsMacOS = map[string]string{
	"ready":           "/System/Library/Sounds/Tink.aiff",
	"command_success": "/System/Library/Sounds/Pop.aiff",
	"command_warning": "/System/Library/Sounds/Basso.aiff",
	"dictation_toggle": "/System/Library/Sounds/Submarine.aiff",
}

// Config holds sound configuration
type Config struct {
	Enabled         bool   `json:"enabled"`
	Ready           string `json:"ready"`
	CommandSuccess  string `json:"command_success"`
	CommandWarning  string `json:"command_warning"`
	DictationToggle string `json:"dictation_toggle"`
}

// DefaultConfig returns platform-specific default sound config
func DefaultConfig() Config {
	defaults := defaultSoundsLinux
	if runtime.GOOS == "darwin" {
		defaults = defaultSoundsMacOS
	}

	return Config{
		Enabled:         true,
		Ready:           defaults["ready"],
		CommandSuccess:  defaults["command_success"],
		CommandWarning:  defaults["command_warning"],
		DictationToggle: defaults["dictation_toggle"],
	}
}

// Player handles sound playback
type Player struct {
	config Config
}

// NewPlayer creates a new sound player
func NewPlayer(config Config) *Player {
	return &Player{config: config}
}

// play plays a sound file (fire-and-forget)
func (p *Player) play(path string) {
	if !p.config.Enabled || path == "" {
		return
	}

	// run in goroutine to not block
	go func() {
		var cmd *exec.Cmd
		if runtime.GOOS == "darwin" {
			cmd = exec.Command("afplay", "-v", "2.5", path)
		} else {
			cmd = exec.Command("paplay", path)
		}
		_ = cmd.Run() // ignore errors
	}()
}

// PlayReady plays the ready sound
func (p *Player) PlayReady() {
	p.play(p.config.Ready)
}

// PlayCommandSuccess plays success sound
func (p *Player) PlayCommandSuccess() {
	p.play(p.config.CommandSuccess)
}

// PlayCommandWarning plays warning sound
func (p *Player) PlayCommandWarning() {
	p.play(p.config.CommandWarning)
}

// PlayDictationToggle plays dictation mode toggle sound
func (p *Player) PlayDictationToggle() {
	p.play(p.config.DictationToggle)
}
