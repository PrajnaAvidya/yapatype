package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/PrajnaAvidya/yapatype/protocol"
)

// kitty ls JSON structure (only fields we need)
type kittyLsWindow struct {
	ID        int  `json:"id"`
	IsFocused bool `json:"is_focused"`
}

type kittyLsTab struct {
	ID        int             `json:"id"`
	IsFocused bool            `json:"is_focused"`
	Windows   []kittyLsWindow `json:"windows"`
}

type kittyLsOSWindow struct {
	ID        int          `json:"id"`
	IsFocused bool         `json:"is_focused"`
	Tabs      []kittyLsTab `json:"tabs"`
}

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

// runKittenOutput runs a kitten command and returns stdout
func (e *KittyExecutor) runKittenOutput(ctx context.Context, args ...string) ([]byte, error) {
	cmdArgs := append(e.baseArgs(), args...)
	cmd := exec.CommandContext(ctx, "kitten", cmdArgs...)
	return cmd.Output()
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

// FocusTab switches which kitty tab receives keystrokes (1-indexed).
// queries kitten @ ls to find the active window in the target tab,
// then updates match to target that window by id.
func (e *KittyExecutor) FocusTab(ctx context.Context, index int) error {
	output, err := e.runKittenOutput(ctx, "ls")
	if err != nil {
		return fmt.Errorf("kitten ls: %w", err)
	}

	var osWindows []kittyLsOSWindow
	if err := json.Unmarshal(output, &osWindows); err != nil {
		return fmt.Errorf("parse kitten ls: %w", err)
	}

	if len(osWindows) == 0 {
		return fmt.Errorf("no kitty windows found")
	}

	// use focused OS window, or first if none focused
	osWin := osWindows[0]
	for _, w := range osWindows {
		if w.IsFocused {
			osWin = w
			break
		}
	}

	// find target tab (1-indexed input)
	tabIdx := index - 1
	if tabIdx < 0 || tabIdx >= len(osWin.Tabs) {
		return fmt.Errorf("tab %d out of range (have %d tabs)", index, len(osWin.Tabs))
	}

	tab := osWin.Tabs[tabIdx]

	// find focused window in tab, fall back to first window
	var windowID int
	found := false
	for _, w := range tab.Windows {
		if w.IsFocused {
			windowID = w.ID
			found = true
			break
		}
	}
	if !found && len(tab.Windows) > 0 {
		windowID = tab.Windows[0].ID
		found = true
	}

	if !found {
		return fmt.Errorf("no windows in tab %d", index)
	}

	e.match = fmt.Sprintf("id:%d", windowID)
	fmt.Printf("targeting kitty window %d (tab %d)\n", windowID, index)
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
