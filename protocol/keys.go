package protocol

// Key represents a keyboard key for SendKey messages
type Key string

const (
	KeyEnter     Key = "enter"
	KeyEscape    Key = "escape"
	KeyTab       Key = "tab"
	KeyBackspace Key = "backspace"
	KeyDelete    Key = "delete"
	KeySpace     Key = "space"
	KeyUp        Key = "up"
	KeyDown      Key = "down"
	KeyLeft      Key = "left"
	KeyRight     Key = "right"
	// letter keys (for shortcuts like ctrl+c, ctrl+w, etc)
	KeyA Key = "a"
	KeyC Key = "c"
	KeyV Key = "v"
	KeyW Key = "w"
	KeyZ Key = "z"
)

// Modifier represents a keyboard modifier for SendKey messages
type Modifier string

const (
	ModCtrl  Modifier = "ctrl"
	ModAlt   Modifier = "alt"
	ModShift Modifier = "shift"
	// note: cmd on mac maps to ctrl for most shortcuts
)
