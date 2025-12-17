package executor

import (
	"fmt"
	"runtime"

	"github.com/PrajnaAvidya/yapatype/config"
)

// NewExecutor creates an executor based on config
func NewExecutor(cfg *config.ClientConfig) (Executor, error) {
	switch cfg.Executor {
	case "osascript":
		return NewMacOSExecutor(), nil
	case "ydotool":
		return NewYdotoolExecutor(), nil
	case "kitty":
		if cfg.KittySocket == nil || *cfg.KittySocket == "" {
			return nil, fmt.Errorf("kitty executor requires --kitty-socket")
		}
		return NewKittyExecutor(*cfg.KittySocket, ""), nil
	case "auto", "":
		// auto-detection
		if cfg.KittySocket != nil && *cfg.KittySocket != "" {
			return NewKittyExecutor(*cfg.KittySocket, ""), nil
		}
		if runtime.GOOS == "darwin" {
			return NewMacOSExecutor(), nil
		}
		return NewYdotoolExecutor(), nil
	default:
		return nil, fmt.Errorf("unknown executor: %s", cfg.Executor)
	}
}
