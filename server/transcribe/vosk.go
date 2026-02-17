package transcribe

import (
	"regexp"
	"strings"
	"sync"
)

// DigitToWord maps digits to word equivalents for grammar
var DigitToWord = map[string]string{
	"2":  "two",
	"3":  "three",
	"4":  "four",
	"5":  "five",
	"6":  "six",
	"7":  "seven",
	"8":  "eight",
	"9":  "nine",
	"10": "ten",
}

// BaseGrammar contains known command phrases
var BaseGrammar = []string{
	"enter", "return", "double enter",
	"escape", "cancel", "control c",
	"tab", "up", "down", "left", "right",
	"up arrow", "down arrow", "left arrow", "right arrow",
	"navigate", "navigate left", "navigate right", "navigate up", "navigate down",
	"go up", "go down",
	"next", "mode", "next mode", "cycle mode",
	"back space", "delete", "space",
	"new line", "slash", "forward slash", "back slash",
	"select all", "copy", "paste", "undo",
	"send", "send it", "submit", "approve",
	"repeat", "repeat last", "again", "send again",
	"hit enter", "press enter", "press return",
	"that's it",
	"scratch", "scratch that", "erase", "erase that", "delete that", "undo that",
	"pause commands", "stop commands",
	"resume commands", "start commands",
	"focus tab", "focus tab one", "focus tab two", "focus tab three",
	"focus tab four", "focus tab five", "focus tab six", "focus tab seven",
	"focus tab eight", "focus tab nine", "focus tab ten",
}

// VoskEngine handles fast command recognition with grammar constraints
type VoskEngine struct {
	modelPath   string
	recognizer  VoskRecognizer
	available   bool
	clientNames map[string]bool
	aliases     map[string]bool
	mu          sync.RWMutex
}

// VoskRecognizer interface for vosk recognition (implemented by voskRecognizerImpl)
type VoskRecognizer interface {
	Setup(modelPath string) error
	Transcribe(audioPath string, grammar []string) (string, error)
	Cleanup()
}

// NewVoskEngine creates a new VoskEngine
func NewVoskEngine(modelPath string) *VoskEngine {
	return &VoskEngine{
		modelPath:   modelPath,
		clientNames: make(map[string]bool),
		aliases:     make(map[string]bool),
	}
}

// SetRecognizer sets the recognizer implementation (for testing or dependency injection)
func (v *VoskEngine) SetRecognizer(rec VoskRecognizer) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.recognizer = rec
}

// Setup loads the vosk model (blocking)
func (v *VoskEngine) Setup() error {
	v.mu.Lock()
	if v.recognizer == nil {
		v.recognizer = NewVoskRecognizer()
	}
	rec := v.recognizer
	v.mu.Unlock()

	if err := rec.Setup(v.modelPath); err != nil {
		return err
	}

	v.mu.Lock()
	v.available = true
	v.mu.Unlock()

	return nil
}

// Available returns true if vosk is loaded and ready
func (v *VoskEngine) Available() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.available
}

// SetAliases sets client aliases (alias -> target mapping)
func (v *VoskEngine) SetAliases(aliases map[string]string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.aliases = make(map[string]bool)
	for alias := range aliases {
		v.aliases[strings.ToLower(alias)] = true
	}
}

// AddClient adds a client name to the grammar
func (v *VoskEngine) AddClient(name string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.clientNames[strings.ToLower(name)] = true
}

// RemoveClient removes a client name from the grammar
func (v *VoskEngine) RemoveClient(name string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.clientNames, strings.ToLower(name))
}

// getNameVariants returns name variants including number word versions
// e.g., "focused2" -> ["focused2", "focused two"]
func getNameVariants(name string) []string {
	variants := []string{name}

	// check if name ends with a digit (e.g., "focused2")
	re := regexp.MustCompile(`^(.+?)(\d+)$`)
	if matches := re.FindStringSubmatch(name); len(matches) == 3 {
		base := matches[1]
		digit := matches[2]
		if word, ok := DigitToWord[digit]; ok {
			variants = append(variants, base+" "+word)
		}
	}

	return variants
}

// BuildGrammar constructs the full grammar including clients and aliases
func (v *VoskEngine) BuildGrammar() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	grammar := make([]string, 0, len(BaseGrammar)+len(v.clientNames)*12+len(v.aliases)*4)
	grammar = append(grammar, BaseGrammar...)

	// add client names with targeting variants
	for name := range v.clientNames {
		for _, variant := range getNameVariants(name) {
			grammar = append(grammar, variant)            // bare name
			grammar = append(grammar, "target "+variant)  // target name
			grammar = append(grammar, "switch "+variant)  // switch name
			grammar = append(grammar, "control "+variant) // control name
		}
	}

	// add aliases with targeting variants
	for alias := range v.aliases {
		grammar = append(grammar, alias)
		grammar = append(grammar, "target "+alias)
		grammar = append(grammar, "switch "+alias)
		grammar = append(grammar, "control "+alias)
	}

	return grammar
}

// Transcribe audio using constrained grammar
// Returns lowercase transcription or empty string
func (v *VoskEngine) Transcribe(audioPath string) (string, error) {
	v.mu.RLock()
	if !v.available || v.recognizer == nil {
		v.mu.RUnlock()
		return "", nil
	}
	rec := v.recognizer
	v.mu.RUnlock()

	grammar := v.BuildGrammar()
	return rec.Transcribe(audioPath, grammar)
}

// Cleanup releases resources
func (v *VoskEngine) Cleanup() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.recognizer != nil {
		v.recognizer.Cleanup()
	}
	v.available = false
}
