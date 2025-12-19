package executor

import (
	"reflect"
	"testing"

	"github.com/PrajnaAvidya/yapatype/protocol"
)

// test Name() methods

func TestMacOSExecutorName(t *testing.T) {
	e := NewMacOSExecutor()
	if e.Name() != "osascript" {
		t.Errorf("Name() = %q, want 'osascript'", e.Name())
	}
}

func TestYdotoolExecutorName(t *testing.T) {
	e := NewYdotoolExecutor()
	if e.Name() != "ydotool" {
		t.Errorf("Name() = %q, want 'ydotool'", e.Name())
	}
}

func TestKittyExecutorName(t *testing.T) {
	e := NewKittyExecutor("/tmp/kitty.sock", "")
	if e.Name() != "kitty" {
		t.Errorf("Name() = %q, want 'kitty'", e.Name())
	}
}

// test SplitBackslashes

func TestSplitBackslashes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantParts    []string
		wantHasSlash bool
	}{
		{
			name:         "no backslash",
			input:        "hello",
			wantParts:    []string{"hello"},
			wantHasSlash: false,
		},
		{
			name:         "middle backslashes",
			input:        "path\\to\\file",
			wantParts:    []string{"path", "to", "file"},
			wantHasSlash: true,
		},
		{
			name:         "leading backslash",
			input:        "\\start",
			wantParts:    []string{"", "start"},
			wantHasSlash: true,
		},
		{
			name:         "trailing backslash",
			input:        "end\\",
			wantParts:    []string{"end", ""},
			wantHasSlash: true,
		},
		{
			name:         "single backslash",
			input:        "\\",
			wantParts:    []string{"", ""},
			wantHasSlash: true,
		},
		{
			name:         "consecutive backslashes",
			input:        "a\\\\b",
			wantParts:    []string{"a", "", "b"},
			wantHasSlash: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parts, hasSlash := SplitBackslashes(tc.input)
			if hasSlash != tc.wantHasSlash {
				t.Errorf("hasSlash = %v, want %v", hasSlash, tc.wantHasSlash)
			}
			if !reflect.DeepEqual(parts, tc.wantParts) {
				t.Errorf("parts = %v, want %v", parts, tc.wantParts)
			}
		})
	}
}

// test BuildKeySequence

func TestBuildKeySequence(t *testing.T) {
	tests := []struct {
		name      string
		key       protocol.Key
		modifiers []protocol.Modifier
		want      []string
	}{
		{
			name:      "simple key",
			key:       protocol.KeyEnter,
			modifiers: nil,
			want:      []string{"28:1", "28:0"},
		},
		{
			name:      "single modifier",
			key:       protocol.KeyC,
			modifiers: []protocol.Modifier{protocol.ModCtrl},
			want:      []string{"29:1", "46:1", "46:0", "29:0"},
		},
		{
			name:      "multiple modifiers reversed",
			key:       protocol.KeyA,
			modifiers: []protocol.Modifier{protocol.ModCtrl, protocol.ModShift},
			// ctrl down, shift down, a down, a up, shift up, ctrl up
			want: []string{"29:1", "42:1", "30:1", "30:0", "42:0", "29:0"},
		},
		{
			name:      "alt modifier",
			key:       protocol.KeyTab,
			modifiers: []protocol.Modifier{protocol.ModAlt},
			want:      []string{"56:1", "15:1", "15:0", "56:0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildKeySequence(tc.key, tc.modifiers)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("BuildKeySequence = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBuildKeySequenceUnknownKey(t *testing.T) {
	got := BuildKeySequence(protocol.Key("unknown"), nil)
	if got != nil {
		t.Errorf("expected nil for unknown key, got %v", got)
	}
}

// test BuildModifierString

func TestBuildModifierString(t *testing.T) {
	tests := []struct {
		name      string
		modifiers []protocol.Modifier
		want      string
	}{
		{
			name:      "empty",
			modifiers: nil,
			want:      "",
		},
		{
			name:      "ctrl",
			modifiers: []protocol.Modifier{protocol.ModCtrl},
			want:      "{control down}",
		},
		{
			name:      "alt",
			modifiers: []protocol.Modifier{protocol.ModAlt},
			want:      "{option down}",
		},
		{
			name:      "shift",
			modifiers: []protocol.Modifier{protocol.ModShift},
			want:      "{shift down}",
		},
		{
			name:      "multiple",
			modifiers: []protocol.Modifier{protocol.ModCtrl, protocol.ModShift},
			want:      "{control down, shift down}",
		},
		{
			name:      "all three",
			modifiers: []protocol.Modifier{protocol.ModCtrl, protocol.ModAlt, protocol.ModShift},
			want:      "{control down, option down, shift down}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildModifierString(tc.modifiers)
			if got != tc.want {
				t.Errorf("BuildModifierString = %q, want %q", got, tc.want)
			}
		})
	}
}

// test BuildKeyScript

func TestBuildKeyScript(t *testing.T) {
	tests := []struct {
		name      string
		key       protocol.Key
		modifiers []protocol.Modifier
		want      string
	}{
		{
			name:      "simple key",
			key:       protocol.KeyEnter,
			modifiers: nil,
			want:      `tell application "System Events" to key code 36`,
		},
		{
			name:      "with modifier",
			key:       protocol.KeyC,
			modifiers: []protocol.Modifier{protocol.ModCtrl},
			want:      `tell application "System Events" to key code 8 using {control down}`,
		},
		{
			name:      "with multiple modifiers",
			key:       protocol.KeyA,
			modifiers: []protocol.Modifier{protocol.ModCtrl, protocol.ModShift},
			want:      `tell application "System Events" to key code 0 using {control down, shift down}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildKeyScript(tc.key, tc.modifiers)
			if got != tc.want {
				t.Errorf("BuildKeyScript = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildKeyScriptUnknownKey(t *testing.T) {
	got := BuildKeyScript(protocol.Key("unknown"), nil)
	if got != "" {
		t.Errorf("expected empty for unknown key, got %q", got)
	}
}

// test BuildKeySpec

func TestBuildKeySpec(t *testing.T) {
	tests := []struct {
		name      string
		key       protocol.Key
		modifiers []protocol.Modifier
		want      string
	}{
		{
			name:      "simple key",
			key:       protocol.KeyEnter,
			modifiers: nil,
			want:      "enter",
		},
		{
			name:      "ctrl+c",
			key:       protocol.KeyC,
			modifiers: []protocol.Modifier{protocol.ModCtrl},
			want:      "ctrl+c",
		},
		{
			name:      "ctrl+shift+a",
			key:       protocol.KeyA,
			modifiers: []protocol.Modifier{protocol.ModCtrl, protocol.ModShift},
			want:      "ctrl+shift+a",
		},
		{
			name:      "alt+tab",
			key:       protocol.KeyTab,
			modifiers: []protocol.Modifier{protocol.ModAlt},
			want:      "alt+tab",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildKeySpec(tc.key, tc.modifiers)
			if got != tc.want {
				t.Errorf("BuildKeySpec = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildKeySpecUnknownKey(t *testing.T) {
	got := BuildKeySpec(protocol.Key("unknown"), nil)
	if got != "" {
		t.Errorf("expected empty for unknown key, got %q", got)
	}
}
