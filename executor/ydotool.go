package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/PrajnaAvidya/yapatype/protocol"
)

// YdotoolExecutor executes keystrokes on linux using ydotool
type YdotoolExecutor struct{}

// NewYdotoolExecutor creates a new ydotool executor
func NewYdotoolExecutor() *YdotoolExecutor {
	return &YdotoolExecutor{}
}

// Name returns the executor name
func (e *YdotoolExecutor) Name() string {
	return "ydotool"
}

// runYdotool runs a ydotool command
func runYdotool(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "ydotool", args...)
	return cmd.Run()
}

// SplitBackslashes splits text for ydotool escaping
// returns the segments to type and whether backslash handling is needed
func SplitBackslashes(text string) ([]string, bool) {
	if !strings.Contains(text, "\\") {
		return []string{text}, false
	}
	return strings.Split(text, "\\"), true
}

// TypeText types text using ydotool type
func (e *YdotoolExecutor) TypeText(ctx context.Context, text string) error {
	if text == "" {
		return nil
	}

	parts, hasBackslash := SplitBackslashes(text)
	if !hasBackslash {
		return runYdotool(ctx, "type", "--key-delay=0", "--key-hold=0", text)
	}

	// type segments with backslash key presses between
	for i, part := range parts {
		if part != "" {
			if err := runYdotool(ctx, "type", "--key-delay=0", "--key-hold=0", part); err != nil {
				return fmt.Errorf("type segment: %w", err)
			}
		}
		// insert backslash key between parts (not after last)
		if i < len(parts)-1 {
			// keycode 43 is backslash, :1 press, :0 release
			if err := runYdotool(ctx, "key", fmt.Sprintf("%d:1", KeyBackslash), fmt.Sprintf("%d:0", KeyBackslash)); err != nil {
				return fmt.Errorf("inject backslash: %w", err)
			}
		}
	}

	return nil
}

// BuildKeySequence builds the ydotool key sequence for a keystroke
// returns nil if the key is unknown
func BuildKeySequence(key protocol.Key, modifiers []protocol.Modifier) []string {
	keycode, ok := LinuxKeycodes[key]
	if !ok {
		return nil
	}

	// build key sequence: modifiers down, key down/up, modifiers up (reversed)
	seq := make([]string, 0, 2+len(modifiers)*2)

	// modifiers down
	for _, mod := range modifiers {
		modCode, ok := LinuxModifierCodes[mod]
		if ok {
			seq = append(seq, fmt.Sprintf("%d:1", modCode))
		}
	}

	// key down and up
	seq = append(seq, fmt.Sprintf("%d:1", keycode))
	seq = append(seq, fmt.Sprintf("%d:0", keycode))

	// modifiers up (reverse order)
	for i := len(modifiers) - 1; i >= 0; i-- {
		modCode, ok := LinuxModifierCodes[modifiers[i]]
		if ok {
			seq = append(seq, fmt.Sprintf("%d:0", modCode))
		}
	}

	return seq
}

// SendKey sends a keystroke using ydotool key
func (e *YdotoolExecutor) SendKey(ctx context.Context, key protocol.Key, modifiers []protocol.Modifier, repeat int) error {
	seq := BuildKeySequence(key, modifiers)
	if seq == nil {
		// unknown key, skip silently
		return nil
	}

	for i := 0; i < repeat; i++ {
		if err := runYdotool(ctx, append([]string{"key"}, seq...)...); err != nil {
			return fmt.Errorf("key: %w", err)
		}
	}

	return nil
}

// FocusTab is a no-op for ydotool (kitty-only feature)
func (e *YdotoolExecutor) FocusTab(ctx context.Context, index int) error {
	return nil
}

// Setup performs initialization (no-op for ydotool)
func (e *YdotoolExecutor) Setup(ctx context.Context) error {
	return nil
}

// Cleanup performs cleanup (no-op for ydotool)
func (e *YdotoolExecutor) Cleanup(ctx context.Context) error {
	return nil
}
