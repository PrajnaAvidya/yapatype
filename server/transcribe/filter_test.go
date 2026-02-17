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

func TestIsHallucination_ShortSegments(t *testing.T) {
	tests := []struct {
		input    string
		maxWords int
		want     bool
	}{
		// short segments with hallucination patterns - should filter
		{"subs by zeoranger", 6, true},
		{"subs by www zeoranger co uk", 6, true},
		{"thank you", 6, true},
		{"Thank You.", 6, true},
		{"thanks for watching", 6, true},
		{"subs by someone", 6, true},

		// long segments with hallucination patterns - should pass
		{"I said thank you to him earlier", 6, false},
		{"I was talking about zeoranger stuff today", 6, false},
		{"thanks for watching this long video tutorial", 6, false},

		// no hallucination pattern
		{"hello world", 6, false},
		{"enter", 6, false},
		{"target desktop", 6, false},

		// edge cases
		{"", 6, false},
		{"   ", 6, false},
	}

	for _, tt := range tests {
		got := IsHallucination(tt.input, tt.maxWords)
		if got != tt.want {
			t.Errorf("IsHallucination(%q, %d) = %v, want %v", tt.input, tt.maxWords, got, tt.want)
		}
	}
}

func TestIsHallucination_Threshold(t *testing.T) {
	// test exact threshold behavior
	text := "one two three four five six" // 6 words

	if !IsHallucination("thank you for this thing here", 6) {
		// 6 words with pattern - should filter
		t.Error("6 words at threshold should filter")
	}

	if IsHallucination("thank you for this thing here now", 6) {
		// 7 words with pattern - should pass
		t.Error("7 words above threshold should pass")
	}

	// text without pattern never filtered
	if IsHallucination(text, 6) {
		t.Error("text without hallucination pattern should never filter")
	}
}
