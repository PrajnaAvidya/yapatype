package executor

import "github.com/PrajnaAvidya/yapatype/protocol"

// linux keycodes (from input-event-codes.h)
var LinuxKeycodes = map[protocol.Key]int{
	protocol.KeyEnter:     28,
	protocol.KeyEscape:    1,
	protocol.KeyTab:       15,
	protocol.KeyBackspace: 14,
	protocol.KeyDelete:    111,
	protocol.KeySpace:     57,
	protocol.KeyUp:        103,
	protocol.KeyDown:      108,
	protocol.KeyLeft:      105,
	protocol.KeyRight:     106,
	protocol.KeyA:         30,
	protocol.KeyC:         46,
	protocol.KeyV:         47,
	protocol.KeyW:         17,
	protocol.KeyZ:         44,
}

// linux modifier keycodes
var LinuxModifierCodes = map[protocol.Modifier]int{
	protocol.ModCtrl:  29,
	protocol.ModAlt:   56,
	protocol.ModShift: 42,
}

// special keycode for backslash (ydotool type interprets \ as escape)
const KeyBackslash = 43

// mac keycodes (System Events key codes)
var MacKeycodes = map[protocol.Key]int{
	protocol.KeyEnter:     36,
	protocol.KeyEscape:    53,
	protocol.KeyTab:       48,
	protocol.KeyBackspace: 51,
	protocol.KeyDelete:    117,
	protocol.KeySpace:     49,
	protocol.KeyUp:        126,
	protocol.KeyDown:      125,
	protocol.KeyLeft:      123,
	protocol.KeyRight:     124,
	protocol.KeyA:         0,
	protocol.KeyC:         8,
	protocol.KeyV:         9,
	protocol.KeyW:         13,
	protocol.KeyZ:         6,
}

// kitty key names (direct string mapping)
var KittyKeyNames = map[protocol.Key]string{
	protocol.KeyEnter:     "enter",
	protocol.KeyEscape:    "escape",
	protocol.KeyTab:       "tab",
	protocol.KeyBackspace: "backspace",
	protocol.KeyDelete:    "delete",
	protocol.KeySpace:     "space",
	protocol.KeyUp:        "up",
	protocol.KeyDown:      "down",
	protocol.KeyLeft:      "left",
	protocol.KeyRight:     "right",
	protocol.KeyA:         "a",
	protocol.KeyC:         "c",
	protocol.KeyV:         "v",
	protocol.KeyW:         "w",
	protocol.KeyZ:         "z",
}

// kitty modifier names
var KittyModifierNames = map[protocol.Modifier]string{
	protocol.ModCtrl:  "ctrl",
	protocol.ModAlt:   "alt",
	protocol.ModShift: "shift",
}
