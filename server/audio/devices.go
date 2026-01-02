package audio

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// ListAudioDevices returns available audio input devices
// macOS: parses sox -V6 -n -t coreaudio output
// Linux: parses arecord -l output (returns hw:X,Y format)
func ListAudioDevices(ctx context.Context) ([]string, error) {
	if runtime.GOOS == "darwin" {
		return listMacOSDevices(ctx)
	}
	return listLinuxDevices(ctx)
}

// listMacOSDevices uses sox to list coreaudio devices
func listMacOSDevices(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// sox -V6 -n -t coreaudio nosuchdevice prints available devices
	cmd := exec.CommandContext(ctx, "sox", "-V6", "-n", "-t", "coreaudio", "nosuchdevice")
	output, _ := cmd.CombinedOutput() // ignore error, sox exits non-zero for invalid device

	// parse lines like: sox INFO coreaudio: Found Audio Device "MacBook Pro Microphone"
	var devices []string
	deviceRegex := regexp.MustCompile(`Found Audio Device "([^"]+)"`)
	for _, line := range strings.Split(string(output), "\n") {
		if matches := deviceRegex.FindStringSubmatch(line); len(matches) > 1 {
			devices = append(devices, matches[1])
		}
	}
	return devices, nil
}

// listLinuxDevices uses arecord to list ALSA devices
func listLinuxDevices(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "arecord", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("arecord -l failed: %w", err)
	}

	// parse lines like: card 1: Device [USB Audio Device], device 0: USB Audio [USB Audio]
	var devices []string
	cardRegex := regexp.MustCompile(`card (\d+):.*\[([^\]]+)\].*device (\d+):`)
	for _, line := range strings.Split(string(output), "\n") {
		if matches := cardRegex.FindStringSubmatch(line); len(matches) > 3 {
			cardNum := matches[1]
			deviceNum := matches[3]
			// return hw:X,Y format for AUDIODEV
			devices = append(devices, fmt.Sprintf("hw:%s,%s", cardNum, deviceNum))
		}
	}
	return devices, nil
}

// FindMatchingDevice finds device matching target substring (case-insensitive)
// Returns device name/id for AUDIODEV env var, or empty string if not found
func FindMatchingDevice(ctx context.Context, target string) (string, error) {
	devices, err := ListAudioDevices(ctx)
	if err != nil {
		return "", err
	}

	targetLower := strings.ToLower(target)
	for _, device := range devices {
		if strings.Contains(strings.ToLower(device), targetLower) {
			return device, nil
		}
	}
	return "", nil
}

// GetDefaultInputDevice returns the default audio input device name
// macOS: uses system_profiler SPAudioDataType
// Linux: uses pactl get-default-source
func GetDefaultInputDevice(ctx context.Context) (string, error) {
	if runtime.GOOS == "darwin" {
		return getMacOSDefaultDevice(ctx)
	}
	return getLinuxDefaultDevice(ctx)
}

// getMacOSDefaultDevice uses system_profiler to find default input device
func getMacOSDefaultDevice(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "system_profiler", "SPAudioDataType")
	output, err := cmd.Output()
	if err != nil {
		return "unknown", nil
	}

	// find line before "Default Input Device: Yes"
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if strings.Contains(line, "Default Input Device: Yes") {
			// look back for device name (line ending with ":")
			for j := i - 1; j >= 0 && j > i-10; j-- {
				candidate := strings.TrimSpace(lines[j])
				if strings.HasSuffix(candidate, ":") && len(candidate) > 1 {
					firstChar := candidate[0]
					if (firstChar >= 'A' && firstChar <= 'Z') || (firstChar >= 'a' && firstChar <= 'z') {
						return strings.TrimSuffix(candidate, ":"), nil
					}
				}
			}
		}
	}
	return "unknown", nil
}

// getLinuxDefaultDevice uses pactl to get default source
func getLinuxDefaultDevice(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pactl", "get-default-source")
	output, err := cmd.Output()
	if err != nil {
		return "default", nil
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "default", nil
	}
	return result, nil
}
