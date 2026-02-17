package protocol

// server -> client messages

// TypeText instructs the client to type the given text
type TypeText struct {
	Type string `json:"type"` // always "type"
	Text string `json:"text"`
}

// SendKey instructs the client to send a key press
type SendKey struct {
	Type      string     `json:"type"`      // always "key"
	Key       Key        `json:"key"`       // the key to press
	Modifiers []Modifier `json:"modifiers"` // modifier keys (ctrl, alt, shift)
	Repeat    int        `json:"repeat"`    // number of times to repeat
}

// Ping is a keepalive message from server to client
type Ping struct {
	Type string `json:"type"` // always "ping"
}

// TargetStatus notifies the client whether it is the active target
type TargetStatus struct {
	Type     string `json:"type"`      // always "target_status"
	IsActive bool   `json:"is_active"` // true if client is active target
}

// FocusTab instructs the client to focus a specific terminal tab
type FocusTab struct {
	Type  string `json:"type"`  // always "focus_tab"
	Index int    `json:"index"` // 1-indexed tab number
}

// Registered confirms client registration with assigned name
type Registered struct {
	Type string `json:"type"` // always "registered"
	Name string `json:"name"` // assigned client name (may differ from requested)
}

// client -> server messages

// Register requests registration with the server
type Register struct {
	Type     string `json:"type"`     // always "register"
	Name     string `json:"name"`     // requested client name
	Platform string `json:"platform"` // "darwin" or "linux"
}

// Prompt sends a whisper prompt to the server
type Prompt struct {
	Type   string `json:"type"`   // always "prompt"
	Prompt string `json:"prompt"` // whisper prompt text
}

// Pong is a keepalive response from client to server
type Pong struct {
	Type string `json:"type"` // always "pong"
}

// constructors to ensure correct type field

func NewTypeText(text string) *TypeText {
	return &TypeText{Type: "type", Text: text}
}

func NewSendKey(key Key, modifiers []Modifier, repeat int) *SendKey {
	if modifiers == nil {
		modifiers = []Modifier{}
	}
	if repeat < 1 {
		repeat = 1
	}
	return &SendKey{Type: "key", Key: key, Modifiers: modifiers, Repeat: repeat}
}

func NewPing() *Ping {
	return &Ping{Type: "ping"}
}

func NewTargetStatus(isActive bool) *TargetStatus {
	return &TargetStatus{Type: "target_status", IsActive: isActive}
}

func NewFocusTab(index int) *FocusTab {
	return &FocusTab{Type: "focus_tab", Index: index}
}

func NewRegistered(name string) *Registered {
	return &Registered{Type: "registered", Name: name}
}

func NewRegister(name, platform string) *Register {
	return &Register{Type: "register", Name: name, Platform: platform}
}

func NewPrompt(prompt string) *Prompt {
	return &Prompt{Type: "prompt", Prompt: prompt}
}

func NewPong() *Pong {
	return &Pong{Type: "pong"}
}
