package commands

import "regexp"

// command patterns for short fragment matching

var (
	// enter variants
	enterPatterns = regexp.MustCompile(
		`^(that'?s?\s?it|said\s?it|set\s?it|send\s?it|sent\s?it|did\s?it|hit\s?it|` +
			`sender|send|enter|and\s?enter|hit\s?end|had\s?enter|submit|approve)$`)

	// double enter
	doubleEnterPatterns = regexp.MustCompile(`^(double\s?enter|enter\s?enter|two\s?enters)$`)

	// escape (sent twice for claude cli)
	escapePatterns = regexp.MustCompile(`^(escape|scape|a\s?scape)$`)

	// cancel / ctrl+c (sent twice)
	cancelPatterns = regexp.MustCompile(`^(cancel|control\s?c|ctrl\s?c)$`)

	// tab
	tabPatterns = regexp.MustCompile(`^(tab|tap|tad)$`)

	// arrow keys
	upPatterns   = regexp.MustCompile(`^(up|go\s?up)$`)
	downPatterns = regexp.MustCompile(`^(down|go\s?down)$`)

	// left/right with ctrl+alt for window navigation
	leftPatterns  = regexp.MustCompile(`^(left|navigate\s?left)$`)
	rightPatterns = regexp.MustCompile(`^(right|navigate\s?right)$`)

	// navigate up/down with ctrl+alt
	navUpPatterns   = regexp.MustCompile(`^(navigate\s?up)$`)
	navDownPatterns = regexp.MustCompile(`^(navigate\s?down)$`)

	// next mode (shift+tab)
	nextModePatterns = regexp.MustCompile(`^(next|mode|next\s?mode)$`)

	// scratch / undo (delete last text)
	scratchPatterns = regexp.MustCompile(
		`^(scratch(\s?that)?|erase(\s?that)?|delete(\s?that)?|undo(\s?that)?)$`)

	// repeat last text
	repeatPatterns = regexp.MustCompile(`^(repeat(\s+(last|previous))?|again|send\s?again)$`)

	// pause commands (enter dictation mode)
	pauseCommandsPatterns = regexp.MustCompile(`^(pause\s?commands?|paws?\s?commands?|stop\s?commands?)$`)

	// resume commands (exit dictation mode)
	resumeCommandsPatterns = regexp.MustCompile(`^(resume\s?commands?|start\s?commands?)$`)

	// target switching
	targetPatterns = regexp.MustCompile(`^(target|switch|control)\s+\S+$`)

	// inline commands (can appear mid-sentence)
	inlinePattern = regexp.MustCompile(`(?i)(slash|back\s?slash)`)

	// end-of-text commands (enter variants at end of sentence)
	endCommandPattern = regexp.MustCompile(
		`(?i)(enter|send(\s?it)?|submit|hit\s?enter|(keypress|press)\s+(enter|return))\s*[.,!?]*$`)
)

// all short command patterns for IsKnownCommand
var shortCommandPatterns = []*regexp.Regexp{
	enterPatterns,
	doubleEnterPatterns,
	escapePatterns,
	cancelPatterns,
	tabPatterns,
	upPatterns,
	downPatterns,
	leftPatterns,
	rightPatterns,
	navUpPatterns,
	navDownPatterns,
	nextModePatterns,
	targetPatterns,
	scratchPatterns,
	repeatPatterns,
	pauseCommandsPatterns,
	resumeCommandsPatterns,
}

// first words that trigger two-word command matching
var firstWordCommands = map[string]bool{
	"enter":  true,
	"send":   true,
	"submit": true,
	"escape": true,
	"scape":  true,
	"tab":    true,
	"tap":    true,
	"up":     true,
	"down":   true,
	"repeat": true,
}
