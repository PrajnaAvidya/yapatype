//go:build vosk

package transcribe

import (
	"encoding/json"
	"os"
	"strings"

	vosk "github.com/alphacep/vosk-api/go"
)

// voskRecognizerImpl implements VoskRecognizer using the vosk library
type voskRecognizerImpl struct {
	model *vosk.VoskModel
}

// NewVoskRecognizer creates a new VoskRecognizer implementation
func NewVoskRecognizer() VoskRecognizer {
	return &voskRecognizerImpl{}
}

// Setup loads the vosk model
func (r *voskRecognizerImpl) Setup(modelPath string) error {
	// suppress vosk logging
	os.Setenv("VOSK_LOG_LEVEL", "-1")
	vosk.SetLogLevel(-1)

	model, err := vosk.NewModel(modelPath)
	if err != nil {
		return err
	}

	r.model = model
	return nil
}

// Transcribe audio using constrained grammar
func (r *voskRecognizerImpl) Transcribe(audioPath string, grammar []string) (string, error) {
	if r.model == nil {
		return "", nil
	}

	// open audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// skip wav header (44 bytes)
	_, err = file.Seek(44, 0)
	if err != nil {
		return "", err
	}

	// build grammar JSON
	grammarJSON, err := json.Marshal(grammar)
	if err != nil {
		return "", err
	}

	// create recognizer with grammar
	rec, err := vosk.NewRecognizerGrm(r.model, 16000, string(grammarJSON))
	if err != nil {
		return "", err
	}
	defer rec.Free()

	// process audio in chunks
	buf := make([]byte, 4000)
	for {
		n, err := file.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			break
		}
		rec.AcceptWaveform(buf[:n])
	}

	// get final result
	result := rec.FinalResult()

	// parse JSON result
	var res struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(result), &res); err != nil {
		return "", nil
	}

	return strings.TrimSpace(res.Text), nil
}

// Cleanup releases resources
func (r *voskRecognizerImpl) Cleanup() {
	if r.model != nil {
		r.model.Free()
		r.model = nil
	}
}
