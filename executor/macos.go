package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/PrajnaAvidya/yapatype/protocol"
)

// MacOSExecutor executes keystrokes on macos using osascript
type MacOSExecutor struct{}

// NewMacOSExecutor creates a new macOS executor
func NewMacOSExecutor() *MacOSExecutor {
	return &MacOSExecutor{}
}

// Name returns the executor name
func (e *MacOSExecutor) Name() string {
	return "osascript"
}

// runOsascript runs an AppleScript via osascript
func runOsascript(ctx context.Context, script string) error {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	return cmd.Run()
}

// TypeText types text via clipboard paste to avoid Kitty keyboard protocol issues
func (e *MacOSExecutor) TypeText(ctx context.Context, text string) error {
	if text == "" {
		return nil
	}

	// use pbcopy to set clipboard, then Cmd+V to paste
	// this avoids System Events keystroke which conflicts with terminals
	// using the Kitty keyboard protocol (like Claude CLI)
	cmd := exec.CommandContext(ctx, "pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pbcopy: %w", err)
	}

	// Cmd+V to paste
	script := `tell application "System Events" to keystroke "v" using command down`
	if err := runOsascript(ctx, script); err != nil {
		return fmt.Errorf("paste: %w", err)
	}

	return nil
}

// SendKey sends a keystroke using System Events key code
func (e *MacOSExecutor) SendKey(ctx context.Context, key protocol.Key, modifiers []protocol.Modifier, repeat int) error {
	keycode, ok := MacKeycodes[key]
	if !ok {
		// unknown key, skip silently
		return nil
	}

	// build modifier string for applescript
	var script string
	if len(modifiers) > 0 {
		modParts := make([]string, 0, len(modifiers))
		for _, mod := range modifiers {
			switch mod {
			case protocol.ModCtrl:
				modParts = append(modParts, "control down")
			case protocol.ModAlt:
				modParts = append(modParts, "option down")
			case protocol.ModShift:
				modParts = append(modParts, "shift down")
			}
		}
		modStr := "{" + strings.Join(modParts, ", ") + "}"
		script = fmt.Sprintf(`tell application "System Events" to key code %d using %s`, keycode, modStr)
	} else {
		script = fmt.Sprintf(`tell application "System Events" to key code %d`, keycode)
	}

	for i := 0; i < repeat; i++ {
		if err := runOsascript(ctx, script); err != nil {
			return fmt.Errorf("key code %d: %w", keycode, err)
		}
	}

	return nil
}

// Setup performs initialization (no-op for macOS)
func (e *MacOSExecutor) Setup(ctx context.Context) error {
	return nil
}

// Cleanup performs cleanup (no-op for macOS)
func (e *MacOSExecutor) Cleanup(ctx context.Context) error {
	return nil
}
