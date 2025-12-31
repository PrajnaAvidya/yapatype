package commands

// action types for command parsing

// ScratchAction signals to delete the last sent text fragment
type ScratchAction struct{}

// RepeatAction signals to repeat the last typed text
type RepeatAction struct{}

// PauseCommandsAction signals to enter dictation mode
type PauseCommandsAction struct{}

// ResumeCommandsAction signals to exit dictation mode
type ResumeCommandsAction struct{}

// Action is any action the parser can produce
// valid types: *protocol.TypeText, *protocol.SendKey, ScratchAction, RepeatAction,
// PauseCommandsAction, ResumeCommandsAction
type Action interface{}

// ParseResult holds the parsed actions and metadata
type ParseResult struct {
	Actions      []Action
	WasCommand   bool   // true if text matched a known command
	OriginalText string
}
