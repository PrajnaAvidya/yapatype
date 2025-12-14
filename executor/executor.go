package executor

import (
	"context"

	"github.com/PrajnaAvidya/yapatype/protocol"
)

// Executor handles platform-specific keyboard input
type Executor interface {
	// Name returns the executor name for logging
	Name() string
	// TypeText types text to the target window
	TypeText(ctx context.Context, text string) error
	// SendKey sends a keystroke with optional modifiers
	SendKey(ctx context.Context, key protocol.Key, modifiers []protocol.Modifier, repeat int) error
	// Setup performs any required initialization
	Setup(ctx context.Context) error
	// Cleanup performs cleanup on shutdown
	Cleanup(ctx context.Context) error
}
