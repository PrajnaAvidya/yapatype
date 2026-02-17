package server

import (
	"fmt"
	"testing"

	"github.com/PrajnaAvidya/yapatype/config"
	"github.com/PrajnaAvidya/yapatype/protocol"
	"github.com/PrajnaAvidya/yapatype/server/commands"
)

func TestTargetPattern(t *testing.T) {
	tests := []struct {
		input   string
		matches bool
		target  string
	}{
		{"target desktop", true, "desktop"},
		{"switch focused", true, "focused"},
		{"control terminal", true, "terminal"},
		{"TARGET Desktop", true, "Desktop"},
		{"target desk top", true, "desk top"},
		{"hello world", false, ""},
		{"target", false, ""},
		{"", false, ""},
	}

	for _, tt := range tests {
		match := targetPattern.FindStringSubmatch(tt.input)
		if tt.matches {
			if len(match) < 2 {
				t.Errorf("targetPattern(%q) should match", tt.input)
				continue
			}
			if match[1] != tt.target {
				t.Errorf("targetPattern(%q) target = %q, want %q", tt.input, match[1], tt.target)
			}
		} else {
			if len(match) > 0 {
				t.Errorf("targetPattern(%q) should not match", tt.input)
			}
		}
	}
}

func TestResumePattern(t *testing.T) {
	tests := []struct {
		input   string
		matches bool
	}{
		{"resume commands", true},
		{"resumecommands", true},
		{"resume command", true},
		{"start commands", true},
		{"startcommands", true},
		{"start command", true},
		{"pause commands", false},
		{"hello", false},
	}

	for _, tt := range tests {
		if resumePattern.MatchString(tt.input) != tt.matches {
			t.Errorf("resumePattern(%q) = %v, want %v", tt.input, !tt.matches, tt.matches)
		}
	}
}

func TestClickPattern(t *testing.T) {
	tests := []struct {
		input   string
		matches bool
	}{
		{"click", true},
		{"Click", true},
		{"CLICK", true},
		{"click.", true},
		{"click,", true},
		{"click, .", true},
		{"clicking", false},
		{"hello click", false},
	}

	for _, tt := range tests {
		if clickPattern.MatchString(tt.input) != tt.matches {
			t.Errorf("clickPattern(%q) = %v, want %v", tt.input, !tt.matches, tt.matches)
		}
	}
}

func TestNewMainServer(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:      "localhost",
		Port:      9999,
		Model:     "/models/test.bin",
		VoskModel: "",
		Sounds:    config.SoundsConfig{Enabled: false},
		Aliases:   map[string]string{"work": "desktop"},
	}

	srv := NewMainServer(cfg, "", "")
	if srv == nil {
		t.Fatal("NewMainServer returned nil")
	}

	if srv.config != cfg {
		t.Error("config not set correctly")
	}
	if srv.wsServer == nil {
		t.Error("wsServer should be initialized")
	}
	if srv.audio == nil {
		t.Error("audio should be initialized")
	}
	if srv.whisper == nil {
		t.Error("whisper should be initialized")
	}
	if srv.sounds == nil {
		t.Error("sounds should be initialized")
	}
	if srv.vosk != nil {
		t.Error("vosk should be nil when VoskModel is empty")
	}
	if srv.prompt != defaultPrompt {
		t.Errorf("prompt = %q, want %q", srv.prompt, defaultPrompt)
	}
}

func TestNewMainServer_WithVosk(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:      "localhost",
		Port:      9999,
		Model:     "/models/test.bin",
		VoskModel: "/models/vosk-model",
		Sounds:    config.SoundsConfig{Enabled: false},
	}

	srv := NewMainServer(cfg, "", "")
	if srv.vosk == nil {
		t.Error("vosk should be initialized when VoskModel is set")
	}
}

func TestMainServer_DictationMode(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:   "localhost",
		Port:   9999,
		Model:  "/models/test.bin",
		Sounds: config.SoundsConfig{Enabled: false},
	}

	srv := NewMainServer(cfg, "", "")

	// initially not in dictation mode
	if srv.dictationMode {
		t.Error("should not start in dictation mode")
	}

	// simulate entering dictation mode
	srv.dictationMode = true
	if !srv.dictationMode {
		t.Error("should be in dictation mode after setting")
	}
}

// test dispatch with various action types
func TestMainServer_DispatchActions(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:   "localhost",
		Port:   9999,
		Model:  "/models/test.bin",
		Sounds: config.SoundsConfig{Enabled: false},
	}

	srv := NewMainServer(cfg, "", "")

	// test with empty actions - should not panic
	result := commands.ParseResult{
		Actions:      []commands.Action{},
		WasCommand:   false,
		OriginalText: "",
	}
	srv.dispatch(result)

	// test with TypeText action - will fail to send but shouldn't panic
	result = commands.ParseResult{
		Actions:      []commands.Action{protocol.NewTypeText("hello ")},
		WasCommand:   false,
		OriginalText: "hello",
	}
	srv.dispatch(result)

	// test with SendKey action
	result = commands.ParseResult{
		Actions:      []commands.Action{protocol.NewSendKey(protocol.KeyEnter, nil, 1)},
		WasCommand:   true,
		OriginalText: "enter",
	}
	srv.dispatch(result)

	// test with ScratchAction - no history, should warn
	result = commands.ParseResult{
		Actions:      []commands.Action{commands.ScratchAction{}},
		WasCommand:   true,
		OriginalText: "scratch",
	}
	srv.dispatch(result)

	// test with RepeatAction - no previous text, should warn
	result = commands.ParseResult{
		Actions:      []commands.Action{commands.RepeatAction{}},
		WasCommand:   true,
		OriginalText: "repeat",
	}
	srv.dispatch(result)

	// test PauseCommandsAction
	result = commands.ParseResult{
		Actions:      []commands.Action{commands.PauseCommandsAction{}},
		WasCommand:   true,
		OriginalText: "pause commands",
	}
	srv.dispatch(result)
	if !srv.dictationMode {
		t.Error("dispatch PauseCommandsAction should enable dictation mode")
	}

	// test ResumeCommandsAction
	result = commands.ParseResult{
		Actions:      []commands.Action{commands.ResumeCommandsAction{}},
		WasCommand:   true,
		OriginalText: "resume commands",
	}
	srv.dispatch(result)
	if srv.dictationMode {
		t.Error("dispatch ResumeCommandsAction should disable dictation mode")
	}
}

func TestMainServer_HandleTranscription_TargetSwitch(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:   "localhost",
		Port:   9999,
		Model:  "/models/test.bin",
		Sounds: config.SoundsConfig{Enabled: false},
	}

	srv := NewMainServer(cfg, "", "")

	// handleTranscription with target command - will fail (no clients)
	// but shouldn't panic
	srv.handleTranscription("target desktop")
	srv.handleTranscription("switch focused")
	srv.handleTranscription("control terminal")
}

func TestMainServer_HandleTranscription_DictationMode(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:   "localhost",
		Port:   9999,
		Model:  "/models/test.bin",
		Sounds: config.SoundsConfig{Enabled: false},
	}

	srv := NewMainServer(cfg, "", "")
	srv.dictationMode = true

	// in dictation mode, "resume commands" should exit
	srv.handleTranscription("resume commands")
	if srv.dictationMode {
		t.Error("'resume commands' should exit dictation mode")
	}

	// re-enable dictation mode
	srv.dictationMode = true

	// other text should stay in dictation mode (and try to type)
	srv.handleTranscription("hello world")
	if !srv.dictationMode {
		t.Error("regular text should not exit dictation mode")
	}
}

func TestMainServer_HandleTranscription_TwoWordFallback(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:   "localhost",
		Port:   9999,
		Model:  "/models/test.bin",
		Sounds: config.SoundsConfig{Enabled: false},
	}

	srv := NewMainServer(cfg, "", "")

	// two-word fallback - "that desktop" should try to target "desktop"
	// will fail (no clients) but shouldn't panic
	srv.handleTranscription("that desktop")
}

func TestFocusTabPattern(t *testing.T) {
	tests := []struct {
		input   string
		matches bool
		tabStr  string
	}{
		{"focus tab 2", true, "2"},
		{"focus tab two", true, "two"},
		{"Focus Tab 3", true, "3"},
		{"focus tab 2.", true, "2"},
		{"focus tab ten", true, "ten"},
		{"focus tab 1!", true, "1"},
		{"tab 2", false, ""},
		{"focus 2", false, ""},
		{"focus tab", false, ""},
		{"hello world", false, ""},
	}

	for _, tt := range tests {
		match := focusTabPattern.FindStringSubmatch(tt.input)
		if tt.matches {
			if len(match) < 2 {
				t.Errorf("focusTabPattern(%q) should match", tt.input)
				continue
			}
			if match[1] != tt.tabStr {
				t.Errorf("focusTabPattern(%q) tab = %q, want %q", tt.input, match[1], tt.tabStr)
			}
		} else {
			if len(match) > 0 {
				t.Errorf("focusTabPattern(%q) should not match", tt.input)
			}
		}
	}
}

func TestFocusTabNumberWordConversion(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"two", 2},
		{"one", 1},
		{"ten", 10},
		{"for", 4},  // common mishearing
		{"to", 2},   // common mishearing
		{"ate", 8},  // common mishearing
	}

	for _, tt := range tests {
		result := ConvertNumberWords(tt.input)
		if result != fmt.Sprintf("%d", tt.want) {
			t.Errorf("ConvertNumberWords(%q) = %q, want %q", tt.input, result, fmt.Sprintf("%d", tt.want))
		}
	}
}

func TestTrailingTabPattern(t *testing.T) {
	tests := []struct {
		input   string
		matches bool
		name    string
		tabStr  string
	}{
		{"focus tab 2", true, "focus", "2"},
		{"desktop tab two", true, "desktop", "two"},
		{"my client tab 3", true, "my client", "3"},
		{"focus tab", false, "", ""},
		{"focus", false, "", ""},
	}

	for _, tt := range tests {
		match := trailingTabPattern.FindStringSubmatch(tt.input)
		if tt.matches {
			if len(match) < 3 {
				t.Errorf("trailingTabPattern(%q) should match", tt.input)
				continue
			}
			if match[1] != tt.name {
				t.Errorf("trailingTabPattern(%q) name = %q, want %q", tt.input, match[1], tt.name)
			}
			if match[2] != tt.tabStr {
				t.Errorf("trailingTabPattern(%q) tab = %q, want %q", tt.input, match[2], tt.tabStr)
			}
		} else {
			if len(match) > 0 {
				t.Errorf("trailingTabPattern(%q) should not match", tt.input)
			}
		}
	}
}

func TestMainServer_HandleTranscription_CombinedTargetTab(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:   "localhost",
		Port:   9999,
		Model:  "/models/test.bin",
		Sounds: config.SoundsConfig{Enabled: false},
	}

	srv := NewMainServer(cfg, "", "")

	// combined target+tab command - will fail (no clients) but shouldn't panic
	srv.handleTranscription("target focus tab 2")
	srv.handleTranscription("switch desktop tab one")
}

func TestMainServer_HandleTranscription_FocusTab(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:   "localhost",
		Port:   9999,
		Model:  "/models/test.bin",
		Sounds: config.SoundsConfig{Enabled: false},
	}

	srv := NewMainServer(cfg, "", "")

	// focus tab command - will fail (no clients) but shouldn't panic
	srv.handleTranscription("focus tab 2")
	srv.handleTranscription("focus tab one")
	srv.handleTranscription("Focus Tab 3")
}

func TestMainServer_Stop(t *testing.T) {
	cfg := &config.ServerConfig{
		Host:   "localhost",
		Port:   9999,
		Model:  "/models/test.bin",
		Sounds: config.SoundsConfig{Enabled: false},
	}

	srv := NewMainServer(cfg, "", "")
	srv.running = true

	srv.Stop()

	if srv.running {
		t.Error("Stop should set running to false")
	}
}
