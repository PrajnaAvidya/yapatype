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

	// save current clipboard contents
	var originalClipboard []byte
	pbpasteCmd := exec.CommandContext(ctx, "pbpaste")
	originalClipboard, _ = pbpasteCmd.Output() // ignore error, clipboard may be empty or non-text

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

	// restore original clipboard contents
	if len(originalClipboard) > 0 {
		restoreCmd := exec.CommandContext(ctx, "pbcopy")
		restoreCmd.Stdin = strings.NewReader(string(originalClipboard))
		restoreCmd.Run() // best effort, ignore errors
	}

	return nil
}

// BuildModifierString builds the AppleScript modifier clause
// returns empty string if no modifiers
func BuildModifierString(modifiers []protocol.Modifier) string {
	if len(modifiers) == 0 {
		return ""
	}

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
	return "{" + strings.Join(modParts, ", ") + "}"
}

// BuildKeyScript builds the full AppleScript for a key code
// returns empty string if key is unknown
func BuildKeyScript(key protocol.Key, modifiers []protocol.Modifier) string {
	keycode, ok := MacKeycodes[key]
	if !ok {
		return ""
	}

	modStr := BuildModifierString(modifiers)
	if modStr != "" {
		return fmt.Sprintf(`tell application "System Events" to key code %d using %s`, keycode, modStr)
	}
	return fmt.Sprintf(`tell application "System Events" to key code %d`, keycode)
}

// SendKey sends a keystroke using System Events key code
func (e *MacOSExecutor) SendKey(ctx context.Context, key protocol.Key, modifiers []protocol.Modifier, repeat int) error {
	script := BuildKeyScript(key, modifiers)
	if script == "" {
		// unknown key, skip silently
		return nil
	}

	for i := 0; i < repeat; i++ {
		if err := runOsascript(ctx, script); err != nil {
			return fmt.Errorf("key code: %w", err)
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
