package executor

import (
	"testing"

	"github.com/PrajnaAvidya/yapatype/protocol"
)

// allKeys returns all protocol.Key values that should have mappings
func allKeys() []protocol.Key {
	return []protocol.Key{
		protocol.KeyEnter,
		protocol.KeyEscape,
		protocol.KeyTab,
		protocol.KeyBackspace,
		protocol.KeyDelete,
		protocol.KeySpace,
		protocol.KeyUp,
		protocol.KeyDown,
		protocol.KeyLeft,
		protocol.KeyRight,
		protocol.KeyA,
		protocol.KeyC,
		protocol.KeyV,
		protocol.KeyW,
		protocol.KeyZ,
	}
}

// allModifiers returns all protocol.Modifier values that should have mappings
func allModifiers() []protocol.Modifier {
	return []protocol.Modifier{
		protocol.ModCtrl,
		protocol.ModAlt,
		protocol.ModShift,
	}
}

func TestLinuxKeycodesComplete(t *testing.T) {
	for _, key := range allKeys() {
		if _, ok := LinuxKeycodes[key]; !ok {
			t.Errorf("LinuxKeycodes missing key: %s", key)
		}
	}
}

func TestMacKeycodesComplete(t *testing.T) {
	for _, key := range allKeys() {
		if _, ok := MacKeycodes[key]; !ok {
			t.Errorf("MacKeycodes missing key: %s", key)
		}
	}
}

func TestKittyKeyNamesComplete(t *testing.T) {
	for _, key := range allKeys() {
		if _, ok := KittyKeyNames[key]; !ok {
			t.Errorf("KittyKeyNames missing key: %s", key)
		}
	}
}

func TestLinuxModifierCodesComplete(t *testing.T) {
	for _, mod := range allModifiers() {
		if _, ok := LinuxModifierCodes[mod]; !ok {
			t.Errorf("LinuxModifierCodes missing modifier: %s", mod)
		}
	}
}

func TestKittyModifierNamesComplete(t *testing.T) {
	for _, mod := range allModifiers() {
		if _, ok := KittyModifierNames[mod]; !ok {
			t.Errorf("KittyModifierNames missing modifier: %s", mod)
		}
	}
}

func TestKeyBackslashValue(t *testing.T) {
	// critical for ydotool backslash escaping
	if KeyBackslash != 43 {
		t.Errorf("KeyBackslash = %d, want 43", KeyBackslash)
	}
}

func TestLinuxKeycodesValues(t *testing.T) {
	// spot check critical values from input-event-codes.h
	tests := []struct {
		key  protocol.Key
		want int
	}{
		{protocol.KeyEnter, 28},
		{protocol.KeyEscape, 1},
		{protocol.KeyBackspace, 14},
		{protocol.KeySpace, 57},
	}

	for _, tc := range tests {
		if got := LinuxKeycodes[tc.key]; got != tc.want {
			t.Errorf("LinuxKeycodes[%s] = %d, want %d", tc.key, got, tc.want)
		}
	}
}

func TestMacKeycodesValues(t *testing.T) {
	// spot check critical values
	tests := []struct {
		key  protocol.Key
		want int
	}{
		{protocol.KeyEnter, 36},
		{protocol.KeyEscape, 53},
		{protocol.KeyBackspace, 51},
		{protocol.KeySpace, 49},
	}

	for _, tc := range tests {
		if got := MacKeycodes[tc.key]; got != tc.want {
			t.Errorf("MacKeycodes[%s] = %d, want %d", tc.key, got, tc.want)
		}
	}
}
