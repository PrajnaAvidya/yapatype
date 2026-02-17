package transcribe

import "strings"

// filler sounds to filter out
var fillerSounds = map[string]bool{
	"mm-hmm": true,
	"mmhmm":  true,
	"uh-huh": true,
	"hmm":    true,
	"uh":     true,
	"um":     true,
}

// hallucination patterns - only filtered in short segments
var hallucinations = []string{
	"zeoranger",           // common whisper silence hallucination
	"thank you",           // silence hallucination
	"thanks for watching", // youtube-style
	"subs by",             // subtitle attribution
}

// IsHallucination returns true if text matches a known hallucination pattern
// only for short segments (≤ maxWords)
func IsHallucination(text string, maxWords int) bool {
	words := strings.Fields(text)
	if len(words) > maxWords {
		return false // long text passes through
	}
	lower := strings.ToLower(text)
	for _, h := range hallucinations {
		if strings.Contains(lower, h) {
			return true
		}
	}
	return false
}

// FilterTranscription normalizes and filters whisper output
// - lowercase and strip whitespace
// - remove newlines
// - filter [BLANK_AUDIO], [silence], etc.
// - filter (coughing), (sighs), etc.
// - filter filler sounds
func FilterTranscription(text string) string {
	// normalize: lowercase, strip, remove newlines
	text = strings.TrimSpace(strings.ToLower(text))
	text = strings.ReplaceAll(text, "\n", " ")

	// skip non-speech (whisper outputs [BLANK_AUDIO], [silence], etc.)
	if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
		return ""
	}

	// skip parenthetical annotations like (clears throat), (sighs), etc.
	if strings.HasPrefix(text, "(") && strings.HasSuffix(text, ")") {
		return ""
	}

	// skip filler sounds (with or without trailing punctuation)
	stripped := strings.TrimRight(text, ".,!?")
	if fillerSounds[stripped] {
		return ""
	}

	return text
}
