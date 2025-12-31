package commands

import (
	"strings"

	"github.com/PrajnaAvidya/yapatype/protocol"
)

// normalizeForMatch prepares text for pattern matching
func normalizeForMatch(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.TrimRight(text, ".,!?")
	return text
}

// IsKnownCommand checks if text matches a known short command
func IsKnownCommand(text string) bool {
	matchText := normalizeForMatch(text)
	if len(matchText) > 20 {
		return false
	}

	// check all patterns
	for _, p := range shortCommandPatterns {
		if p.MatchString(matchText) {
			return true
		}
	}

	// two-word fallback: "enter now", "tab please", "repeat that"
	words := strings.Fields(matchText)
	if len(words) == 2 {
		if firstWordCommands[words[0]] {
			return true
		}
	}

	return false
}

// ParseShortCommand tries to parse text as a short command
// returns nil if not matched
func ParseShortCommand(text string) *ParseResult {
	matchText := normalizeForMatch(text)
	if len(matchText) > 15 {
		return nil
	}

	// double enter (check before single enter)
	if doubleEnterPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyEnter, nil, 2)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// enter
	if enterPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyEnter, nil, 1)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// escape (double for claude cli)
	if escapePatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyEscape, nil, 2)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// cancel / ctrl+c (double)
	if cancelPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyC, []protocol.Modifier{protocol.ModCtrl}, 2)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// tab
	if tabPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyTab, nil, 1)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// arrow keys
	if upPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyUp, nil, 1)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	if downPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyDown, nil, 1)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// navigate with ctrl+alt
	if leftPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyLeft, []protocol.Modifier{protocol.ModCtrl, protocol.ModAlt}, 1)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	if rightPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyRight, []protocol.Modifier{protocol.ModCtrl, protocol.ModAlt}, 1)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	if navUpPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyUp, []protocol.Modifier{protocol.ModCtrl, protocol.ModAlt}, 1)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	if navDownPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyDown, []protocol.Modifier{protocol.ModCtrl, protocol.ModAlt}, 1)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// next mode (shift+tab)
	if nextModePatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{protocol.NewSendKey(protocol.KeyTab, []protocol.Modifier{protocol.ModShift}, 1)},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// scratch / undo
	if scratchPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{ScratchAction{}},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// repeat
	if repeatPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{RepeatAction{}},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// pause commands (enter dictation mode)
	if pauseCommandsPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{PauseCommandsAction{}},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// resume commands (exit dictation mode)
	if resumeCommandsPatterns.MatchString(matchText) {
		return &ParseResult{
			Actions:      []Action{ResumeCommandsAction{}},
			WasCommand:   true,
			OriginalText: text,
		}
	}

	// two-word fallback
	words := strings.Fields(matchText)
	if len(words) == 2 {
		first := words[0]
		second := words[1]

		// first word triggers: "enter now", "tab please", etc.
		switch first {
		case "enter", "send", "submit":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyEnter, nil, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "escape", "scape":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyEscape, nil, 2)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "tab", "tap":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyTab, nil, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "up":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyUp, nil, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "down":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyDown, nil, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "repeat":
			return &ParseResult{
				Actions:      []Action{RepeatAction{}},
				WasCommand:   true,
				OriginalText: text,
			}
		}

		// second word triggers: "ok enter", "ok tab", etc.
		switch second {
		case "enter", "send", "submit":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyEnter, nil, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "escape", "scape":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyEscape, nil, 2)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "tab", "tap":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyTab, nil, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "up":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyUp, nil, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "down":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyDown, nil, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "backspace":
			// ctrl+w = delete word
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyW, []protocol.Modifier{protocol.ModCtrl}, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "newline":
			// backslash then enter (shell continuation)
			return &ParseResult{
				Actions: []Action{
					protocol.NewTypeText("\\"),
					protocol.NewSendKey(protocol.KeyEnter, nil, 1),
				},
				WasCommand:   true,
				OriginalText: text,
			}
		case "nextmode", "mode":
			return &ParseResult{
				Actions:      []Action{protocol.NewSendKey(protocol.KeyTab, []protocol.Modifier{protocol.ModShift}, 1)},
				WasCommand:   true,
				OriginalText: text,
			}
		case "repeat":
			return &ParseResult{
				Actions:      []Action{RepeatAction{}},
				WasCommand:   true,
				OriginalText: text,
			}
		}
	}

	return nil
}

// executeInlineCommand converts inline command name to action
func executeInlineCommand(cmd string) Action {
	cmd = strings.ToLower(strings.ReplaceAll(cmd, " ", ""))
	switch cmd {
	case "slash":
		return protocol.NewTypeText("/")
	case "backslash":
		return protocol.NewTypeText("\\")
	}
	return nil
}

// ParseText parses transcribed text into actions
// handles short commands, inline commands, and end-of-text commands
func ParseText(text string) ParseResult {
	if text == "" || strings.TrimSpace(text) == "" {
		return ParseResult{Actions: []Action{}, WasCommand: false, OriginalText: text}
	}

	// first try short command matching
	if short := ParseShortCommand(text); short != nil {
		return *short
	}

	// process text for inline and end commands
	var actions []Action
	remaining := text
	normalized := strings.ToLower(text)

	for remaining != "" {
		// look for inline commands (slash, backslash)
		match := inlinePattern.FindStringIndex(normalized)
		endMatch := endCommandPattern.FindStringIndex(normalized)

		if match != nil {
			start := match[0]
			end := match[1]
			cmd := normalized[start:end]

			// type text before command
			before := strings.TrimRight(remaining[:start], " .,!?")
			if before != "" {
				actions = append(actions, protocol.NewTypeText(before))
			}

			// execute inline command
			if action := executeInlineCommand(cmd); action != nil {
				actions = append(actions, action)
			}

			// continue with remaining text
			remaining = strings.TrimLeft(remaining[end:], " .,!?")
			normalized = strings.ToLower(remaining)
			continue
		}

		if endMatch != nil {
			// text before the end command
			before := strings.TrimRight(remaining[:endMatch[0]], " .,!?")
			if before != "" {
				actions = append(actions, protocol.NewTypeText(before+" "))
			}

			// send enter
			actions = append(actions, protocol.NewSendKey(protocol.KeyEnter, nil, 1))
			return ParseResult{Actions: actions, WasCommand: true, OriginalText: text}
		}

		// no commands found, type remaining text with trailing space
		if remaining != "" {
			actions = append(actions, protocol.NewTypeText(remaining+" "))
		}
		break
	}

	wasCommand := false
	for _, a := range actions {
		if _, ok := a.(*protocol.SendKey); ok {
			wasCommand = true
			break
		}
	}
	return ParseResult{Actions: actions, WasCommand: wasCommand, OriginalText: text}
}
