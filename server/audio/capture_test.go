package audio

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultAudioConfig(t *testing.T) {
	cfg := DefaultAudioConfig()

	if cfg.SilenceThreshold != "2%" {
		t.Errorf("SilenceThreshold = %q, want '2%%'", cfg.SilenceThreshold)
	}
	if cfg.SilenceDuration != 0.7 {
		t.Errorf("SilenceDuration = %v, want 0.7", cfg.SilenceDuration)
	}
	if cfg.SampleRate != 16000 {
		t.Errorf("SampleRate = %d, want 16000", cfg.SampleRate)
	}
	if cfg.Channels != 1 {
		t.Errorf("Channels = %d, want 1", cfg.Channels)
	}
}

func TestNewAudioCapture(t *testing.T) {
	cfg := DefaultAudioConfig()
	ac := NewAudioCapture(cfg, "hw:1,0")

	if ac.Config.SampleRate != 16000 {
		t.Error("Config not set correctly")
	}
	if ac.Device != "hw:1,0" {
		t.Errorf("Device = %q, want 'hw:1,0'", ac.Device)
	}
	if ac.tempDir == "" {
		t.Error("tempDir should not be empty")
	}
}

func TestAudioCaptureGetEnv(t *testing.T) {
	cfg := DefaultAudioConfig()

	// no device - should return nil
	ac := NewAudioCapture(cfg, "")
	env := ac.getEnv()
	if env != nil {
		t.Error("getEnv should return nil when no device set")
	}

	// with device - should include AUDIODEV
	ac = NewAudioCapture(cfg, "hw:1,0")
	env = ac.getEnv()
	if env == nil {
		t.Fatal("getEnv should return env when device is set")
	}

	found := false
	for _, e := range env {
		if e == "AUDIODEV=hw:1,0" {
			found = true
			break
		}
	}
	if !found {
		t.Error("AUDIODEV=hw:1,0 not found in env")
	}
}

func TestAudioCaptureGetTempPath(t *testing.T) {
	cfg := DefaultAudioConfig()
	ac := NewAudioCapture(cfg, "")

	path := ac.getTempPath()
	if !strings.HasSuffix(path, "yapatype-recording.wav") {
		t.Errorf("getTempPath = %q, should end with yapatype-recording.wav", path)
	}
}

func TestAudioCaptureCleanup(t *testing.T) {
	cfg := DefaultAudioConfig()
	ac := NewAudioCapture(cfg, "")

	// create a temp file
	tmpFile := filepath.Join(os.TempDir(), "yapatype-test-cleanup.wav")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// verify it exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("temp file should exist")
	}

	// cleanup
	ac.Cleanup(tmpFile)

	// verify it's gone
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("temp file should be deleted after Cleanup")
	}
}

func TestAudioCaptureCleanupEmpty(t *testing.T) {
	cfg := DefaultAudioConfig()
	ac := NewAudioCapture(cfg, "")

	// should not panic with empty path
	ac.Cleanup("")
}

func TestAudioCaptureCancel(t *testing.T) {
	cfg := DefaultAudioConfig()
	ac := NewAudioCapture(cfg, "")

	// should not panic when no process running
	ac.Cancel()
}

// TestGetDurationWithRealFile tests GetDuration with a small real wav file
// This test is skipped if sox is not available
func TestGetDurationWithRealFile(t *testing.T) {
	// check if sox is available
	_, err := os.Stat("/usr/bin/sox")
	if os.IsNotExist(err) {
		// try common paths
		if _, err := os.Stat("/opt/homebrew/bin/sox"); os.IsNotExist(err) {
			t.Skip("sox not found, skipping duration test")
		}
	}

	cfg := DefaultAudioConfig()
	ac := NewAudioCapture(cfg, "")

	// create a minimal wav file using sox
	tmpFile := filepath.Join(os.TempDir(), "yapatype-duration-test.wav")
	defer os.Remove(tmpFile)

	// generate 0.5s of silence
	ctx := context.Background()
	cmd := commandContext(ctx, "sox", "-n", "-r", "16000", "-c", "1", tmpFile, "trim", "0", "0.5")
	if err := cmd.Run(); err != nil {
		t.Skipf("could not create test wav file: %v", err)
	}

	duration, err := ac.GetDuration(tmpFile)
	if err != nil {
		t.Fatalf("GetDuration failed: %v", err)
	}

	// should be approximately 0.5s
	if duration < 0.4 || duration > 0.6 {
		t.Errorf("duration = %v, want ~0.5", duration)
	}
}

// TestConcatenateWithRealFiles tests Concatenate with real wav files
// This test is skipped if sox is not available
func TestConcatenateWithRealFiles(t *testing.T) {
	// check if sox is available
	if _, err := lookPath("sox"); err != nil {
		t.Skip("sox not found, skipping concatenate test")
	}

	cfg := DefaultAudioConfig()
	ac := NewAudioCapture(cfg, "")

	tmpDir := os.TempDir()
	file1 := filepath.Join(tmpDir, "yapatype-concat-1.wav")
	file2 := filepath.Join(tmpDir, "yapatype-concat-2.wav")
	output := filepath.Join(tmpDir, "yapatype-concat-out.wav")
	defer os.Remove(file1)
	defer os.Remove(file2)
	defer os.Remove(output)

	ctx := context.Background()

	// generate two short wav files
	cmd := commandContext(ctx, "sox", "-n", "-r", "16000", "-c", "1", file1, "trim", "0", "0.3")
	if err := cmd.Run(); err != nil {
		t.Skipf("could not create test wav file 1: %v", err)
	}

	cmd = commandContext(ctx, "sox", "-n", "-r", "16000", "-c", "1", file2, "trim", "0", "0.2")
	if err := cmd.Run(); err != nil {
		t.Skipf("could not create test wav file 2: %v", err)
	}

	// concatenate
	err := ac.Concatenate(file1, file2, output)
	if err != nil {
		t.Fatalf("Concatenate failed: %v", err)
	}

	// check output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Fatal("output file should exist")
	}

	// check duration is ~0.5s
	duration, err := ac.GetDuration(output)
	if err != nil {
		t.Fatalf("GetDuration failed: %v", err)
	}

	if duration < 0.4 || duration > 0.6 {
		t.Errorf("concatenated duration = %v, want ~0.5", duration)
	}
}

// helper for exec.LookPath
func lookPath(cmd string) (string, error) {
	return exec.LookPath(cmd)
}

// helper for exec.CommandContext
func commandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// TestListAudioDevicesParsingMacOS tests macOS device parsing
func TestListAudioDevicesParsingMacOS(t *testing.T) {
	// test the regex parsing logic directly
	sampleOutput := `sox WARN coreaudio: Can't open output device: Invalid device
sox INFO coreaudio: Found Audio Device "MacBook Pro Microphone"
sox INFO coreaudio: Found Audio Device "External Microphone"
sox FAIL formats: can't open output file`

	var devices []string
	for _, line := range strings.Split(sampleOutput, "\n") {
		if strings.Contains(line, `Found Audio Device "`) {
			start := strings.Index(line, `"`)
			end := strings.LastIndex(line, `"`)
			if start != -1 && end > start {
				devices = append(devices, line[start+1:end])
			}
		}
	}

	if len(devices) != 2 {
		t.Errorf("parsed %d devices, want 2", len(devices))
	}
	if len(devices) > 0 && devices[0] != "MacBook Pro Microphone" {
		t.Errorf("first device = %q, want 'MacBook Pro Microphone'", devices[0])
	}
	if len(devices) > 1 && devices[1] != "External Microphone" {
		t.Errorf("second device = %q, want 'External Microphone'", devices[1])
	}
}

// TestListAudioDevicesParsingLinux tests Linux device parsing
func TestListAudioDevicesParsingLinux(t *testing.T) {
	// test the regex parsing logic directly
	sampleOutput := `**** List of CAPTURE Hardware Devices ****
card 0: PCH [HDA Intel PCH], device 0: ALC897 Analog [ALC897 Analog]
  Subdevices: 1/1
  Subdevice #0: subdevice #0
card 1: Device [USB Audio Device], device 0: USB Audio [USB Audio]
  Subdevices: 1/1
  Subdevice #0: subdevice #0`

	var devices []string
	for _, line := range strings.Split(sampleOutput, "\n") {
		// look for "card X:" pattern
		if strings.HasPrefix(strings.TrimSpace(line), "card ") && strings.Contains(line, "device ") {
			// extract card and device numbers
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				cardNum := strings.TrimSuffix(parts[1], ":")
				// find device number
				for i, p := range parts {
					if p == "device" && i+1 < len(parts) {
						deviceNum := strings.TrimSuffix(parts[i+1], ":")
						devices = append(devices, "hw:"+cardNum+","+deviceNum)
						break
					}
				}
			}
		}
	}

	if len(devices) != 2 {
		t.Errorf("parsed %d devices, want 2", len(devices))
	}
	if len(devices) > 0 && devices[0] != "hw:0,0" {
		t.Errorf("first device = %q, want 'hw:0,0'", devices[0])
	}
	if len(devices) > 1 && devices[1] != "hw:1,0" {
		t.Errorf("second device = %q, want 'hw:1,0'", devices[1])
	}
}

// TestFindMatchingDeviceLogic tests the substring matching logic
func TestFindMatchingDeviceLogic(t *testing.T) {
	devices := []string{
		"MacBook Pro Microphone",
		"External USB Mic",
		"Bluetooth Headset",
	}

	tests := []struct {
		target string
		want   string
	}{
		{"macbook", "MacBook Pro Microphone"},
		{"MACBOOK", "MacBook Pro Microphone"},
		{"usb", "External USB Mic"},
		{"bluetooth", "Bluetooth Headset"},
		{"headset", "Bluetooth Headset"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			targetLower := strings.ToLower(tt.target)
			var found string
			for _, device := range devices {
				if strings.Contains(strings.ToLower(device), targetLower) {
					found = device
					break
				}
			}
			if found != tt.want {
				t.Errorf("match(%q) = %q, want %q", tt.target, found, tt.want)
			}
		})
	}
}

// TestGetDefaultDeviceParsingMacOS tests macOS default device parsing
func TestGetDefaultDeviceParsingMacOS(t *testing.T) {
	sampleOutput := `Audio:

    Devices:

        External Microphone:

          Default Input Device: Yes
          Input Channels: 2

        Built-in Output:

          Default Output Device: Yes
          Output Channels: 2`

	lines := strings.Split(sampleOutput, "\n")
	var result string

	for i, line := range lines {
		if strings.Contains(line, "Default Input Device: Yes") {
			// look back for device name
			for j := i - 1; j >= 0 && j > i-10; j-- {
				candidate := strings.TrimSpace(lines[j])
				if strings.HasSuffix(candidate, ":") && len(candidate) > 1 {
					firstChar := candidate[0]
					if (firstChar >= 'A' && firstChar <= 'Z') || (firstChar >= 'a' && firstChar <= 'z') {
						result = strings.TrimSuffix(candidate, ":")
						break
					}
				}
			}
		}
	}

	if result != "External Microphone" {
		t.Errorf("parsed default = %q, want 'External Microphone'", result)
	}
}
