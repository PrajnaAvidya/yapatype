package commands

import (
	"testing"

	"github.com/PrajnaAvidya/yapatype/protocol"
)

func TestIsKnownCommand(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// enter variants
		{"enter", true},
		{"send", true},
		{"send it", true},
		{"that's it", true},
		{"approve", true},

		// escape
		{"escape", true},
		{"scape", true},

		// tab
		{"tab", true},
		{"tap", true},

		// arrows
		{"up", true},
		{"down", true},
		{"go up", true},

		// not commands
		{"hello world", false},
		{"this is a long sentence", false},

		// first word triggers
		{"enter now", true},
		{"tab please", true},
		{"escape quick", true},
		{"up arrow", true},
		{"down arrow", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsKnownCommand(tt.input)
			if got != tt.want {
				t.Errorf("IsKnownCommand(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseShortCommandEnter(t *testing.T) {
	result := ParseShortCommand("enter")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(result.Actions))
	}
	key, ok := result.Actions[0].(*protocol.SendKey)
	if !ok {
		t.Fatalf("expected SendKey, got %T", result.Actions[0])
	}
	if key.Key != protocol.KeyEnter {
		t.Errorf("expected KeyEnter, got %v", key.Key)
	}
}

func TestParseShortCommandDoubleEnter(t *testing.T) {
	result := ParseShortCommand("double enter")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Repeat != 2 {
		t.Errorf("expected repeat=2, got %d", key.Repeat)
	}
}

func TestParseShortCommandApprove(t *testing.T) {
	result := ParseShortCommand("approve")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyEnter {
		t.Errorf("expected KeyEnter, got %v", key.Key)
	}
}

func TestParseShortCommandEscape(t *testing.T) {
	result := ParseShortCommand("escape")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyEscape {
		t.Errorf("expected KeyEscape, got %v", key.Key)
	}
	if key.Repeat != 2 {
		t.Errorf("expected repeat=2 for claude cli, got %d", key.Repeat)
	}
}

func TestParseShortCommandTab(t *testing.T) {
	result := ParseShortCommand("tab")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyTab {
		t.Errorf("expected KeyTab, got %v", key.Key)
	}
}

func TestParseShortCommandCancel(t *testing.T) {
	result := ParseShortCommand("cancel")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyC {
		t.Errorf("expected KeyC, got %v", key.Key)
	}
	hasCtrl := false
	for _, m := range key.Modifiers {
		if m == protocol.ModCtrl {
			hasCtrl = true
		}
	}
	if !hasCtrl {
		t.Error("expected CTRL modifier")
	}
}

func TestParseShortCommandTwoWordFallback(t *testing.T) {
	result := ParseShortCommand("okay enter")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyEnter {
		t.Errorf("expected KeyEnter, got %v", key.Key)
	}
}

func TestParseShortCommandNotCommand(t *testing.T) {
	result := ParseShortCommand("hello")
	if result != nil {
		t.Error("expected nil for non-command")
	}
}

func TestParseTextPlainText(t *testing.T) {
	result := ParseText("hello world")
	if result.WasCommand {
		t.Error("expected WasCommand = false")
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(result.Actions))
	}
	tt, ok := result.Actions[0].(*protocol.TypeText)
	if !ok {
		t.Fatalf("expected TypeText, got %T", result.Actions[0])
	}
	if tt.Text != "hello world " {
		t.Errorf("text = %q, want 'hello world '", tt.Text)
	}
}

func TestParseTextInlineSlash(t *testing.T) {
	result := ParseText("run slash test")
	// "run" + "/" + "test "
	if len(result.Actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(result.Actions))
	}
	tt0 := result.Actions[0].(*protocol.TypeText)
	if tt0.Text != "run" {
		t.Errorf("first action text = %q, want 'run'", tt0.Text)
	}
	tt1 := result.Actions[1].(*protocol.TypeText)
	if tt1.Text != "/" {
		t.Errorf("second action text = %q, want '/'", tt1.Text)
	}
	tt2 := result.Actions[2].(*protocol.TypeText)
	if tt2.Text != "test " {
		t.Errorf("third action text = %q, want 'test '", tt2.Text)
	}
}

func TestParseTextInlineBackslash(t *testing.T) {
	result := ParseText("path backslash file")
	if len(result.Actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(result.Actions))
	}
	tt0 := result.Actions[0].(*protocol.TypeText)
	if tt0.Text != "path" {
		t.Errorf("first action text = %q, want 'path'", tt0.Text)
	}
	tt1 := result.Actions[1].(*protocol.TypeText)
	if tt1.Text != "\\" {
		t.Errorf("second action text = %q, want '\\\\'", tt1.Text)
	}
	tt2 := result.Actions[2].(*protocol.TypeText)
	if tt2.Text != "file " {
		t.Errorf("third action text = %q, want 'file '", tt2.Text)
	}
}

func TestParseTextInlineNewlineMidSentence(t *testing.T) {
	// newline mid-sentence should type literally (not trigger command)
	result := ParseText("first newline second")
	if len(result.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(result.Actions))
	}
	tt := result.Actions[0].(*protocol.TypeText)
	if tt.Text != "first newline second " {
		t.Errorf("text = %q, want 'first newline second '", tt.Text)
	}
}

func TestParseTextEndCommandSendIt(t *testing.T) {
	result := ParseText("hello world send it")
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	hasEnter := false
	for _, a := range result.Actions {
		if key, ok := a.(*protocol.SendKey); ok && key.Key == protocol.KeyEnter {
			hasEnter = true
		}
	}
	if !hasEnter {
		t.Error("expected Enter key action")
	}
}

func TestParseTextEndCommandPressEnter(t *testing.T) {
	result := ParseText("type this press enter")
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	hasEnter := false
	for _, a := range result.Actions {
		if key, ok := a.(*protocol.SendKey); ok && key.Key == protocol.KeyEnter {
			hasEnter = true
		}
	}
	if !hasEnter {
		t.Error("expected Enter key action")
	}
}

func TestParseTextEmpty(t *testing.T) {
	result := ParseText("")
	if len(result.Actions) != 0 {
		t.Errorf("expected 0 actions, got %d", len(result.Actions))
	}
	if result.WasCommand {
		t.Error("expected WasCommand = false")
	}
}

func TestParseTextShortCommandDirect(t *testing.T) {
	result := ParseText("tab")
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyTab {
		t.Errorf("expected KeyTab, got %v", key.Key)
	}
}

// scratch command tests

func TestScratchThat(t *testing.T) {
	result := ParseShortCommand("scratch that")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(result.Actions))
	}
	if _, ok := result.Actions[0].(ScratchAction); !ok {
		t.Errorf("expected ScratchAction, got %T", result.Actions[0])
	}
}

func TestScratchAlone(t *testing.T) {
	result := ParseShortCommand("scratch")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ScratchAction); !ok {
		t.Errorf("expected ScratchAction, got %T", result.Actions[0])
	}
}

func TestEraseThat(t *testing.T) {
	result := ParseShortCommand("erase that")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ScratchAction); !ok {
		t.Errorf("expected ScratchAction, got %T", result.Actions[0])
	}
}

func TestEraseAlone(t *testing.T) {
	result := ParseShortCommand("erase")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ScratchAction); !ok {
		t.Errorf("expected ScratchAction, got %T", result.Actions[0])
	}
}

func TestDeleteThat(t *testing.T) {
	result := ParseShortCommand("delete that")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ScratchAction); !ok {
		t.Errorf("expected ScratchAction, got %T", result.Actions[0])
	}
}

func TestDeleteAlone(t *testing.T) {
	result := ParseShortCommand("delete")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ScratchAction); !ok {
		t.Errorf("expected ScratchAction, got %T", result.Actions[0])
	}
}

func TestUndoThat(t *testing.T) {
	result := ParseShortCommand("undo that")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ScratchAction); !ok {
		t.Errorf("expected ScratchAction, got %T", result.Actions[0])
	}
}

func TestUndoAlone(t *testing.T) {
	result := ParseShortCommand("undo")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ScratchAction); !ok {
		t.Errorf("expected ScratchAction, got %T", result.Actions[0])
	}
}

func TestScratchIsKnownCommand(t *testing.T) {
	commands := []string{
		"scratch that", "scratch",
		"erase", "erase that",
		"delete", "delete that",
		"undo", "undo that",
	}
	for _, cmd := range commands {
		if !IsKnownCommand(cmd) {
			t.Errorf("IsKnownCommand(%q) = false, want true", cmd)
		}
	}
}

// end-only command tests

func TestBackspaceMidSentenceIgnored(t *testing.T) {
	// "I need to backspace this" should type literally
	result := ParseText("I need to backspace this")
	if result.WasCommand {
		t.Error("expected WasCommand = false")
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(result.Actions))
	}
	tt := result.Actions[0].(*protocol.TypeText)
	if tt.Text != "I need to backspace this " {
		t.Errorf("text = %q", tt.Text)
	}
}

func TestNewlineMidSentenceIgnored(t *testing.T) {
	// "add newline here please" should type literally
	result := ParseText("add newline here please")
	if result.WasCommand {
		t.Error("expected WasCommand = false")
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(result.Actions))
	}
	tt := result.Actions[0].(*protocol.TypeText)
	if tt.Text != "add newline here please " {
		t.Errorf("text = %q", tt.Text)
	}
}

func TestOkBackspaceWorks(t *testing.T) {
	// "ok backspace" two-word fallback should trigger delete word
	result := ParseShortCommand("ok backspace")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyW {
		t.Errorf("expected KeyW, got %v", key.Key)
	}
	hasCtrl := false
	for _, m := range key.Modifiers {
		if m == protocol.ModCtrl {
			hasCtrl = true
		}
	}
	if !hasCtrl {
		t.Error("expected CTRL modifier")
	}
}

func TestOkNewlineWorks(t *testing.T) {
	// "ok newline" two-word fallback should trigger newline
	result := ParseShortCommand("ok newline")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	if len(result.Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(result.Actions))
	}
	tt := result.Actions[0].(*protocol.TypeText)
	if tt.Text != "\\" {
		t.Errorf("first action text = %q, want '\\\\'", tt.Text)
	}
	key := result.Actions[1].(*protocol.SendKey)
	if key.Key != protocol.KeyEnter {
		t.Errorf("expected KeyEnter, got %v", key.Key)
	}
}

func TestBackspaceAtEndTypesLiterally(t *testing.T) {
	// "fix this backspace" should type literally
	result := ParseText("fix this backspace")
	if result.WasCommand {
		t.Error("expected WasCommand = false")
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(result.Actions))
	}
	tt := result.Actions[0].(*protocol.TypeText)
	if tt.Text != "fix this backspace " {
		t.Errorf("text = %q", tt.Text)
	}
}

func TestNewlineAtEndTypesLiterally(t *testing.T) {
	// "first line newline" should type literally
	result := ParseText("first line newline")
	if result.WasCommand {
		t.Error("expected WasCommand = false")
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(result.Actions))
	}
	tt := result.Actions[0].(*protocol.TypeText)
	if tt.Text != "first line newline " {
		t.Errorf("text = %q", tt.Text)
	}
}

func TestEnterAtEndWorks(t *testing.T) {
	// "type this enter" should type "type this" then press enter
	result := ParseText("type this enter")
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	if len(result.Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(result.Actions))
	}
	tt := result.Actions[0].(*protocol.TypeText)
	if tt.Text != "type this " {
		t.Errorf("first action text = %q, want 'type this '", tt.Text)
	}
	key := result.Actions[1].(*protocol.SendKey)
	if key.Key != protocol.KeyEnter {
		t.Errorf("expected KeyEnter, got %v", key.Key)
	}
}

func TestSubmitAtEndWorks(t *testing.T) {
	// "done with this submit" should type then press enter
	result := ParseText("done with this submit")
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	hasEnter := false
	for _, a := range result.Actions {
		if key, ok := a.(*protocol.SendKey); ok && key.Key == protocol.KeyEnter {
			hasEnter = true
		}
	}
	if !hasEnter {
		t.Error("expected Enter key action")
	}
}

// repeat command tests

func TestRepeat(t *testing.T) {
	result := ParseShortCommand("repeat")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	if _, ok := result.Actions[0].(RepeatAction); !ok {
		t.Errorf("expected RepeatAction, got %T", result.Actions[0])
	}
}

func TestRepeatLast(t *testing.T) {
	result := ParseShortCommand("repeat last")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(RepeatAction); !ok {
		t.Errorf("expected RepeatAction, got %T", result.Actions[0])
	}
}

func TestRepeatPrevious(t *testing.T) {
	result := ParseShortCommand("repeat previous")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(RepeatAction); !ok {
		t.Errorf("expected RepeatAction, got %T", result.Actions[0])
	}
}

func TestRepeatTwoWordSecondMatches(t *testing.T) {
	// "X repeat" - second word triggers
	result := ParseShortCommand("ok repeat")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(RepeatAction); !ok {
		t.Errorf("expected RepeatAction, got %T", result.Actions[0])
	}
}

func TestRepeatTwoWordFirstMatches(t *testing.T) {
	// "repeat X" - first word triggers
	result := ParseShortCommand("repeat that")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(RepeatAction); !ok {
		t.Errorf("expected RepeatAction, got %T", result.Actions[0])
	}

	result = ParseShortCommand("repeat this")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(RepeatAction); !ok {
		t.Errorf("expected RepeatAction, got %T", result.Actions[0])
	}
}

func TestAgain(t *testing.T) {
	result := ParseShortCommand("again")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(RepeatAction); !ok {
		t.Errorf("expected RepeatAction, got %T", result.Actions[0])
	}
}

func TestSendAgain(t *testing.T) {
	result := ParseShortCommand("send again")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(RepeatAction); !ok {
		t.Errorf("expected RepeatAction, got %T", result.Actions[0])
	}
}

func TestRepeatIsKnownCommand(t *testing.T) {
	commands := []string{
		"repeat", "repeat last", "repeat previous",
		"repeat that", "repeat this",
		"again", "send again",
	}
	for _, cmd := range commands {
		if !IsKnownCommand(cmd) {
			t.Errorf("IsKnownCommand(%q) = false, want true", cmd)
		}
	}
}

// first word command tests

func TestEnterFirst(t *testing.T) {
	result := ParseShortCommand("enter now")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyEnter {
		t.Errorf("expected KeyEnter, got %v", key.Key)
	}
}

func TestTabFirst(t *testing.T) {
	result := ParseShortCommand("tab please")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyTab {
		t.Errorf("expected KeyTab, got %v", key.Key)
	}
}

func TestEscapeFirst(t *testing.T) {
	result := ParseShortCommand("escape quick")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyEscape {
		t.Errorf("expected KeyEscape, got %v", key.Key)
	}
}

func TestUpFirst(t *testing.T) {
	result := ParseShortCommand("up arrow")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyUp {
		t.Errorf("expected KeyUp, got %v", key.Key)
	}
}

func TestDownFirst(t *testing.T) {
	result := ParseShortCommand("down arrow")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyDown {
		t.Errorf("expected KeyDown, got %v", key.Key)
	}
}

// pause/resume command tests

func TestPauseCommands(t *testing.T) {
	result := ParseShortCommand("pause commands")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	if _, ok := result.Actions[0].(PauseCommandsAction); !ok {
		t.Errorf("expected PauseCommandsAction, got %T", result.Actions[0])
	}
}

func TestPauseCommandSingular(t *testing.T) {
	result := ParseShortCommand("pause command")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(PauseCommandsAction); !ok {
		t.Errorf("expected PauseCommandsAction, got %T", result.Actions[0])
	}
}

func TestStopCommands(t *testing.T) {
	result := ParseShortCommand("stop commands")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(PauseCommandsAction); !ok {
		t.Errorf("expected PauseCommandsAction, got %T", result.Actions[0])
	}
}

func TestStopCommandSingular(t *testing.T) {
	result := ParseShortCommand("stop command")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(PauseCommandsAction); !ok {
		t.Errorf("expected PauseCommandsAction, got %T", result.Actions[0])
	}
}

func TestResumeCommands(t *testing.T) {
	result := ParseShortCommand("resume commands")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.WasCommand {
		t.Error("expected WasCommand = true")
	}
	if _, ok := result.Actions[0].(ResumeCommandsAction); !ok {
		t.Errorf("expected ResumeCommandsAction, got %T", result.Actions[0])
	}
}

func TestResumeCommandSingular(t *testing.T) {
	result := ParseShortCommand("resume command")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ResumeCommandsAction); !ok {
		t.Errorf("expected ResumeCommandsAction, got %T", result.Actions[0])
	}
}

func TestStartCommands(t *testing.T) {
	result := ParseShortCommand("start commands")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ResumeCommandsAction); !ok {
		t.Errorf("expected ResumeCommandsAction, got %T", result.Actions[0])
	}
}

func TestStartCommandSingular(t *testing.T) {
	result := ParseShortCommand("start command")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(ResumeCommandsAction); !ok {
		t.Errorf("expected ResumeCommandsAction, got %T", result.Actions[0])
	}
}

func TestPauseResumeIsKnownCommand(t *testing.T) {
	commands := []string{
		"pause commands", "stop commands",
		"resume commands", "start commands",
	}
	for _, cmd := range commands {
		if !IsKnownCommand(cmd) {
			t.Errorf("IsKnownCommand(%q) = false, want true", cmd)
		}
	}
}

func TestPawsMistranscription(t *testing.T) {
	// "paws commands" is a common mistranscription of "pause commands"
	result := ParseShortCommand("paws commands")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := result.Actions[0].(PauseCommandsAction); !ok {
		t.Errorf("expected PauseCommandsAction, got %T", result.Actions[0])
	}
}

// navigation tests

func TestLeftWithModifiers(t *testing.T) {
	result := ParseShortCommand("left")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyLeft {
		t.Errorf("expected KeyLeft, got %v", key.Key)
	}
	hasCtrl := false
	hasAlt := false
	for _, m := range key.Modifiers {
		if m == protocol.ModCtrl {
			hasCtrl = true
		}
		if m == protocol.ModAlt {
			hasAlt = true
		}
	}
	if !hasCtrl || !hasAlt {
		t.Error("expected CTRL+ALT modifiers")
	}
}

func TestRightWithModifiers(t *testing.T) {
	result := ParseShortCommand("right")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyRight {
		t.Errorf("expected KeyRight, got %v", key.Key)
	}
	hasCtrl := false
	hasAlt := false
	for _, m := range key.Modifiers {
		if m == protocol.ModCtrl {
			hasCtrl = true
		}
		if m == protocol.ModAlt {
			hasAlt = true
		}
	}
	if !hasCtrl || !hasAlt {
		t.Error("expected CTRL+ALT modifiers")
	}
}

func TestNavigateUp(t *testing.T) {
	result := ParseShortCommand("navigate up")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyUp {
		t.Errorf("expected KeyUp, got %v", key.Key)
	}
	hasCtrl := false
	hasAlt := false
	for _, m := range key.Modifiers {
		if m == protocol.ModCtrl {
			hasCtrl = true
		}
		if m == protocol.ModAlt {
			hasAlt = true
		}
	}
	if !hasCtrl || !hasAlt {
		t.Error("expected CTRL+ALT modifiers")
	}
}

func TestNavigateDown(t *testing.T) {
	result := ParseShortCommand("navigate down")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyDown {
		t.Errorf("expected KeyDown, got %v", key.Key)
	}
	hasCtrl := false
	hasAlt := false
	for _, m := range key.Modifiers {
		if m == protocol.ModCtrl {
			hasCtrl = true
		}
		if m == protocol.ModAlt {
			hasAlt = true
		}
	}
	if !hasCtrl || !hasAlt {
		t.Error("expected CTRL+ALT modifiers")
	}
}

func TestNextMode(t *testing.T) {
	result := ParseShortCommand("next mode")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyTab {
		t.Errorf("expected KeyTab, got %v", key.Key)
	}
	hasShift := false
	for _, m := range key.Modifiers {
		if m == protocol.ModShift {
			hasShift = true
		}
	}
	if !hasShift {
		t.Error("expected SHIFT modifier")
	}
}

func TestFocusTabIsKnownCommand(t *testing.T) {
	commands := []string{
		"focus tab two", "focus tab 2",
		"focus tab one", "focus tab ten",
	}
	for _, cmd := range commands {
		if !IsKnownCommand(cmd) {
			t.Errorf("IsKnownCommand(%q) = false, want true", cmd)
		}
	}
}

func TestModeAlone(t *testing.T) {
	result := ParseShortCommand("mode")
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	key := result.Actions[0].(*protocol.SendKey)
	if key.Key != protocol.KeyTab {
		t.Errorf("expected KeyTab, got %v", key.Key)
	}
}
