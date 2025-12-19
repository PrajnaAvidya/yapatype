package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/PrajnaAvidya/yapatype/protocol"
)

// KittyExecutor executes keystrokes via kitty remote control
type KittyExecutor struct {
	socketPath string
	match      string // optional window match expression
}

// NewKittyExecutor creates a new kitty executor
func NewKittyExecutor(socketPath, match string) *KittyExecutor {
	return &KittyExecutor{
		socketPath: socketPath,
		match:      match,
	}
}

// Name returns the executor name
func (e *KittyExecutor) Name() string {
	return "kitty"
}

// baseArgs returns the base kitten command arguments
func (e *KittyExecutor) baseArgs() []string {
	return []string{"@", "--to", "unix:" + e.socketPath}
}

// matchArgs returns the match arguments if configured
func (e *KittyExecutor) matchArgs() []string {
	if e.match != "" {
		return []string{"--match", e.match}
	}
	return nil
}

// runKitten runs a kitten command
func (e *KittyExecutor) runKitten(ctx context.Context, args ...string) error {
	cmdArgs := append(e.baseArgs(), args...)
	cmd := exec.CommandContext(ctx, "kitten", cmdArgs...)
	return cmd.Run()
}

// TypeText types text using kitten @ send-text
func (e *KittyExecutor) TypeText(ctx context.Context, text string) error {
	if text == "" {
		return nil
	}

	args := []string{"send-text"}
	args = append(args, e.matchArgs()...)
	args = append(args, "--", text)

	if err := e.runKitten(ctx, args...); err != nil {
		return fmt.Errorf("send-text: %w", err)
	}

	return nil
}

// BuildKeySpec builds the kitty key specification (e.g., "ctrl+shift+a")
// returns empty string if key is unknown
func BuildKeySpec(key protocol.Key, modifiers []protocol.Modifier) string {
	keyname, ok := KittyKeyNames[key]
	if !ok {
		return ""
	}

	if len(modifiers) == 0 {
		return keyname
	}

	modNames := make([]string, 0, len(modifiers))
	for _, mod := range modifiers {
		modName, ok := KittyModifierNames[mod]
		if ok {
			modNames = append(modNames, modName)
		}
	}
	return strings.Join(append(modNames, keyname), "+")
}

// SendKey sends a keystroke using kitten @ send-key
func (e *KittyExecutor) SendKey(ctx context.Context, key protocol.Key, modifiers []protocol.Modifier, repeat int) error {
	keyspec := BuildKeySpec(key, modifiers)
	if keyspec == "" {
		// unknown key, skip silently
		return nil
	}

	args := []string{"send-key"}
	args = append(args, e.matchArgs()...)
	args = append(args, keyspec)

	for i := 0; i < repeat; i++ {
		if err := e.runKitten(ctx, args...); err != nil {
			return fmt.Errorf("send-key %s: %w", keyspec, err)
		}
	}

	return nil
}

// Setup performs initialization (no-op for kitty)
func (e *KittyExecutor) Setup(ctx context.Context) error {
	return nil
}

// Cleanup performs cleanup (no-op for kitty)
func (e *KittyExecutor) Cleanup(ctx context.Context) error {
	return nil
}
