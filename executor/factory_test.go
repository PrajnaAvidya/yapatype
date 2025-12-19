package executor

import (
	"runtime"
	"testing"

	"github.com/PrajnaAvidya/yapatype/config"
)

func ptr(s string) *string {
	return &s
}

func TestNewExecutorOsascript(t *testing.T) {
	cfg := &config.ClientConfig{Executor: "osascript"}
	exec, err := NewExecutor(cfg)
	if err != nil {
		t.Fatalf("NewExecutor: %v", err)
	}
	if exec.Name() != "osascript" {
		t.Errorf("Name() = %q, want 'osascript'", exec.Name())
	}
	if _, ok := exec.(*MacOSExecutor); !ok {
		t.Errorf("expected *MacOSExecutor, got %T", exec)
	}
}

func TestNewExecutorYdotool(t *testing.T) {
	cfg := &config.ClientConfig{Executor: "ydotool"}
	exec, err := NewExecutor(cfg)
	if err != nil {
		t.Fatalf("NewExecutor: %v", err)
	}
	if exec.Name() != "ydotool" {
		t.Errorf("Name() = %q, want 'ydotool'", exec.Name())
	}
	if _, ok := exec.(*YdotoolExecutor); !ok {
		t.Errorf("expected *YdotoolExecutor, got %T", exec)
	}
}

func TestNewExecutorKitty(t *testing.T) {
	cfg := &config.ClientConfig{
		Executor:    "kitty",
		KittySocket: ptr("/tmp/kitty.sock"),
	}
	exec, err := NewExecutor(cfg)
	if err != nil {
		t.Fatalf("NewExecutor: %v", err)
	}
	if exec.Name() != "kitty" {
		t.Errorf("Name() = %q, want 'kitty'", exec.Name())
	}
	if _, ok := exec.(*KittyExecutor); !ok {
		t.Errorf("expected *KittyExecutor, got %T", exec)
	}
}

func TestNewExecutorKittyNoSocket(t *testing.T) {
	cfg := &config.ClientConfig{Executor: "kitty"}
	_, err := NewExecutor(cfg)
	if err == nil {
		t.Error("expected error for kitty without socket")
	}
}

func TestNewExecutorKittyEmptySocket(t *testing.T) {
	cfg := &config.ClientConfig{
		Executor:    "kitty",
		KittySocket: ptr(""),
	}
	_, err := NewExecutor(cfg)
	if err == nil {
		t.Error("expected error for kitty with empty socket")
	}
}

func TestNewExecutorAutoWithSocket(t *testing.T) {
	cfg := &config.ClientConfig{
		Executor:    "auto",
		KittySocket: ptr("/tmp/kitty.sock"),
	}
	exec, err := NewExecutor(cfg)
	if err != nil {
		t.Fatalf("NewExecutor: %v", err)
	}
	if exec.Name() != "kitty" {
		t.Errorf("Name() = %q, want 'kitty'", exec.Name())
	}
}

func TestNewExecutorAutoLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping on non-linux")
	}
	cfg := &config.ClientConfig{Executor: "auto"}
	exec, err := NewExecutor(cfg)
	if err != nil {
		t.Fatalf("NewExecutor: %v", err)
	}
	if exec.Name() != "ydotool" {
		t.Errorf("Name() = %q, want 'ydotool'", exec.Name())
	}
}

func TestNewExecutorAutoDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-darwin")
	}
	cfg := &config.ClientConfig{Executor: "auto"}
	exec, err := NewExecutor(cfg)
	if err != nil {
		t.Fatalf("NewExecutor: %v", err)
	}
	if exec.Name() != "osascript" {
		t.Errorf("Name() = %q, want 'osascript'", exec.Name())
	}
}

func TestNewExecutorEmpty(t *testing.T) {
	// empty string should behave like "auto"
	cfg := &config.ClientConfig{Executor: ""}
	exec, err := NewExecutor(cfg)
	if err != nil {
		t.Fatalf("NewExecutor: %v", err)
	}
	// should get platform default
	if exec == nil {
		t.Error("expected non-nil executor")
	}
}

func TestNewExecutorUnknown(t *testing.T) {
	cfg := &config.ClientConfig{Executor: "unknown"}
	_, err := NewExecutor(cfg)
	if err == nil {
		t.Error("expected error for unknown executor")
	}
}
