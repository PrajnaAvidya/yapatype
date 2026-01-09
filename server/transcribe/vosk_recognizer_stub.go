//go:build !vosk

package transcribe

import "errors"

// voskRecognizerStub is a stub implementation when vosk is not available
type voskRecognizerStub struct{}

// NewVoskRecognizer returns stub implementation when vosk build tag is not set
func NewVoskRecognizer() VoskRecognizer {
	return &voskRecognizerStub{}
}

// Setup returns error since vosk is not available
func (r *voskRecognizerStub) Setup(modelPath string) error {
	return errors.New("vosk not available: build with -tags vosk")
}

// Transcribe returns empty since vosk is not available
func (r *voskRecognizerStub) Transcribe(audioPath string, grammar []string) (string, error) {
	return "", nil
}

// Cleanup is a no-op
func (r *voskRecognizerStub) Cleanup() {}
