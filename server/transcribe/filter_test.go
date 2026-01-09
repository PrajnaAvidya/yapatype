package transcribe

import "testing"

func TestFilterTranscription_Normalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello world"},
		{"  HELLO  ", "hello"},
		{"Hello\nWorld", "hello world"},
		{"  Mixed\n  Case  ", "mixed   case"},
	}

	for _, tt := range tests {
		got := FilterTranscription(tt.input)
		if got != tt.want {
			t.Errorf("FilterTranscription(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFilterTranscription_BlankAudio(t *testing.T) {
	tests := []string{
		"[BLANK_AUDIO]",
		"[blank_audio]",
		"[silence]",
		"[SILENCE]",
		"[music]",
	}

	for _, input := range tests {
		got := FilterTranscription(input)
		if got != "" {
			t.Errorf("FilterTranscription(%q) = %q, want empty", input, got)
		}
	}
}

func TestFilterTranscription_Parenthetical(t *testing.T) {
	tests := []string{
		"(coughing)",
		"(sighs)",
		"(clears throat)",
		"(LAUGHING)",
	}

	for _, input := range tests {
		got := FilterTranscription(input)
		if got != "" {
			t.Errorf("FilterTranscription(%q) = %q, want empty", input, got)
		}
	}
}

func TestFilterTranscription_Fillers(t *testing.T) {
	tests := []string{
		"hmm",
		"Hmm",
		"uh",
		"um",
		"mm-hmm",
		"mmhmm",
		"uh-huh",
		"hmm.",
		"uh!",
		"um?",
	}

	for _, input := range tests {
		got := FilterTranscription(input)
		if got != "" {
			t.Errorf("FilterTranscription(%q) = %q, want empty", input, got)
		}
	}
}

func TestFilterTranscription_NormalText(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello, how are you?", "hello, how are you?"},
		{"Type this text please", "type this text please"},
		{"enter", "enter"},
		{"send it", "send it"},
	}

	for _, tt := range tests {
		got := FilterTranscription(tt.input)
		if got != tt.want {
			t.Errorf("FilterTranscription(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFilterTranscription_EdgeCases(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"   ", ""},
		{"[partial", "[partial"},      // not a complete bracket
		{"partial]", "partial]"},      // not a complete bracket
		{"(partial", "(partial"},      // not complete parens
		{"partial)", "partial)"},      // not complete parens
		{"hmmm", "hmmm"},              // not a filler (extra m)
		{"ummm", "ummm"},              // not a filler
		{"hello [note] world", "hello [note] world"}, // brackets in middle
	}

	for _, tt := range tests {
		got := FilterTranscription(tt.input)
		if got != tt.want {
			t.Errorf("FilterTranscription(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
