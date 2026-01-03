package audio

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// AudioConfig holds audio capture settings
type AudioConfig struct {
	SilenceThreshold string  // "2%" - audio level considered silence
	SilenceDuration  float64 // 0.7 - seconds of silence before stopping
	SampleRate       int     // 16000
	Channels         int     // 1 (mono)
}

// DefaultAudioConfig returns default audio capture settings
func DefaultAudioConfig() AudioConfig {
	return AudioConfig{
		SilenceThreshold: "2%",
		SilenceDuration:  0.7,
		SampleRate:       16000,
		Channels:         1,
	}
}

// AudioCapture handles audio recording via sox
type AudioCapture struct {
	Config           AudioConfig
	Device           string // exact device name/id for AUDIODEV
	tempDir          string
	currentCmd       *exec.Cmd
	mu               sync.Mutex
	consecutiveEmpty int
	lastDevice       string
}

// NewAudioCapture creates a new AudioCapture with given config
func NewAudioCapture(config AudioConfig, device string) *AudioCapture {
	return &AudioCapture{
		Config:  config,
		Device:  device,
		tempDir: os.TempDir(),
	}
}

// getEnv returns environment with AUDIODEV set if using specific device
func (a *AudioCapture) getEnv() []string {
	if a.Device == "" {
		return nil
	}
	env := os.Environ()
	return append(env, "AUDIODEV="+a.Device)
}

// getTempPath returns path for temporary audio file
func (a *AudioCapture) getTempPath() string {
	return filepath.Join(a.tempDir, "yapatype-recording.wav")
}

// Record captures audio until silence detected
// Returns path to wav file and duration, or empty string if no audio
func (a *AudioCapture) Record(ctx context.Context) (string, float64, error) {
	outputPath := a.getTempPath()

	// build sox rec command
	// rec: record audio
	// -q: quiet (no progress)
	// -b 16: 16-bit
	// rate: sample rate
	// channels: mono
	// silence: stop on silence (1 = require 1 period of silence, duration, threshold)
	args := []string{
		"-q",
		"-b", "16",
		outputPath,
		"rate", strconv.Itoa(a.Config.SampleRate),
		"channels", strconv.Itoa(a.Config.Channels),
		"silence",
		"1", "0.0", "0%", // no start silence detection (avoid clipping first words)
		"1", fmt.Sprintf("%.1f", a.Config.SilenceDuration), a.Config.SilenceThreshold,
	}

	// default 120s timeout
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rec", args...)
	if env := a.getEnv(); env != nil {
		cmd.Env = env
	}

	a.mu.Lock()
	a.currentCmd = cmd
	a.mu.Unlock()

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	a.mu.Lock()
	a.currentCmd = nil
	a.mu.Unlock()

	// check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		a.consecutiveEmpty++
		return "", 0, nil
	}

	// check for sox errors
	if err != nil {
		a.consecutiveEmpty++
		if a.consecutiveEmpty == 1 || a.consecutiveEmpty%10 == 0 {
			errMsg := stderrBuf.String()
			if errMsg == "" {
				errMsg = err.Error()
			}
			log.Printf("[audio] sox error (exit): %s", strings.TrimSpace(errMsg))
		}
		return "", 0, nil
	}

	// check if we got any audio
	stat, err := os.Stat(outputPath)
	if err != nil || stat.Size() == 0 {
		a.consecutiveEmpty++

		// after many consecutive empty recordings, check audio health
		if a.consecutiveEmpty >= 20 && a.consecutiveEmpty%20 == 0 {
			healthy, msg := a.CheckAudioHealth(context.Background())
			log.Printf("[audio] %d empty recordings, health check: healthy=%v, %s", a.consecutiveEmpty, healthy, msg)
		}

		return "", 0, nil
	}

	// got audio, reset counter and get duration
	a.consecutiveEmpty = 0
	duration, _ := a.GetDuration(outputPath)
	return outputPath, duration, nil
}

// RecordContinuation captures short continuation for VAD extension
// maxDuration is max seconds to record (default 1.2s)
func (a *AudioCapture) RecordContinuation(ctx context.Context, maxDuration float64) (string, error) {
	if maxDuration <= 0 {
		maxDuration = 1.2
	}

	outputPath := filepath.Join(a.tempDir, "yapatype-continuation.wav")

	args := []string{
		"-q",
		outputPath,
		"rate", strconv.Itoa(a.Config.SampleRate),
		"trim", "0", fmt.Sprintf("%.1f", maxDuration),
		"silence",
		"1", "0.1", "1%", // start threshold
		"1", "0.3", a.Config.SilenceThreshold, // stop threshold
	}

	// use maxDuration + buffer as timeout
	timeout := time.Duration(maxDuration*1000+500) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rec", args...)
	if env := a.getEnv(); env != nil {
		cmd.Env = env
	}

	a.mu.Lock()
	a.currentCmd = cmd
	a.mu.Unlock()

	err := cmd.Run()

	a.mu.Lock()
	a.currentCmd = nil
	a.mu.Unlock()

	if err != nil {
		return "", nil // ignore errors for continuation
	}

	stat, err := os.Stat(outputPath)
	if err != nil || stat.Size() == 0 {
		return "", nil
	}

	return outputPath, nil
}

// GetDuration returns audio file duration in seconds
func (a *AudioCapture) GetDuration(audioPath string) (float64, error) {
	cmd := exec.Command("sox", audioPath, "-n", "stat")
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	_ = cmd.Run() // sox stat outputs to stderr

	// parse "Length (seconds):  1.234" from stat output
	output := stderrBuf.String()
	lengthRegex := regexp.MustCompile(`Length \(seconds\):\s+([\d.]+)`)
	if matches := lengthRegex.FindStringSubmatch(output); len(matches) > 1 {
		return strconv.ParseFloat(matches[1], 64)
	}

	return 0, fmt.Errorf("could not parse duration from sox stat output")
}

// Concatenate combines two audio files into output
func (a *AudioCapture) Concatenate(path1, path2, output string) error {
	cmd := exec.Command("sox", path1, path2, output)
	return cmd.Run()
}

// Cancel kills any running recording subprocess
func (a *AudioCapture) Cancel() {
	a.mu.Lock()
	cmd := a.currentCmd
	a.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

// CheckAudioHealth verifies audio capture is working
// Returns (healthy bool, message string)
func (a *AudioCapture) CheckAudioHealth(ctx context.Context) (bool, string) {
	currentDevice, _ := GetDefaultInputDevice(ctx)

	// first time initialization
	if a.lastDevice == "" {
		a.lastDevice = currentDevice
	}

	// check for device change
	if currentDevice != a.lastDevice {
		oldDevice := a.lastDevice
		a.lastDevice = currentDevice
		return true, fmt.Sprintf("audio device changed: '%s' -> '%s'", oldDevice, currentDevice)
	}

	// try a quick test recording (0.1s) to see if sox can access the device
	testPath := filepath.Join(a.tempDir, "yapatype-test.wav")
	defer func() {
		_ = os.Remove(testPath)
	}()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	args := []string{
		"-q",
		testPath,
		"rate", strconv.Itoa(a.Config.SampleRate),
		"trim", "0", "0.1",
	}

	cmd := exec.CommandContext(ctx, "rec", args...)
	if env := a.getEnv(); env != nil {
		cmd.Env = env
	}

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return false, "audio test recording timed out"
	}

	if err != nil {
		errMsg := stderrBuf.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return false, fmt.Sprintf("audio test failed: %s", strings.TrimSpace(errMsg))
	}

	return true, fmt.Sprintf("audio ok (device: %s)", currentDevice)
}

// Cleanup removes temporary audio file
func (a *AudioCapture) Cleanup(path string) {
	if path != "" {
		_ = os.Remove(path)
	}
}
