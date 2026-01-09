package transcribe

import (
	"reflect"
	"sort"
	"testing"
)

func TestNewVoskEngine(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")

	if v.modelPath != "/models/vosk-small" {
		t.Errorf("modelPath = %q, want /models/vosk-small", v.modelPath)
	}
	if v.available {
		t.Error("available should be false before Setup")
	}
	if v.clientNames == nil {
		t.Error("clientNames should be initialized")
	}
	if v.aliases == nil {
		t.Error("aliases should be initialized")
	}
}

func TestVoskEngine_Available(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")

	if v.Available() {
		t.Error("Available() should return false before Setup")
	}
}

func TestVoskEngine_AddRemoveClient(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")

	v.AddClient("focused")
	v.AddClient("Desktop")

	if !v.clientNames["focused"] {
		t.Error("focused should be in clientNames")
	}
	if !v.clientNames["desktop"] {
		t.Error("desktop (lowercase) should be in clientNames")
	}

	v.RemoveClient("Focused")
	if v.clientNames["focused"] {
		t.Error("focused should be removed")
	}
	if !v.clientNames["desktop"] {
		t.Error("desktop should still be in clientNames")
	}
}

func TestVoskEngine_SetAliases(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")

	aliases := map[string]string{
		"work":   "desktop",
		"Claude": "focused",
	}
	v.SetAliases(aliases)

	if !v.aliases["work"] {
		t.Error("work should be in aliases")
	}
	if !v.aliases["claude"] {
		t.Error("claude (lowercase) should be in aliases")
	}
}

func TestGetNameVariants(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"focused", []string{"focused"}},
		{"desktop", []string{"desktop"}},
		{"focused2", []string{"focused2", "focused two"}},
		{"focused3", []string{"focused3", "focused three"}},
		{"focused10", []string{"focused10", "focused ten"}},
		{"focused1", []string{"focused1"}}, // 1 not in DigitToWord
		{"focused11", []string{"focused11"}}, // 11 not in DigitToWord
		{"test5name", []string{"test5name"}}, // digit not at end
	}

	for _, tt := range tests {
		got := getNameVariants(tt.name)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("getNameVariants(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestVoskEngine_BuildGrammar_BaseOnly(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")
	grammar := v.BuildGrammar()

	// should contain all base grammar
	grammarSet := make(map[string]bool)
	for _, g := range grammar {
		grammarSet[g] = true
	}

	for _, base := range BaseGrammar {
		if !grammarSet[base] {
			t.Errorf("base grammar missing: %q", base)
		}
	}
}

func TestVoskEngine_BuildGrammar_WithClients(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")
	v.AddClient("focused")
	v.AddClient("desktop")

	grammar := v.BuildGrammar()
	grammarSet := make(map[string]bool)
	for _, g := range grammar {
		grammarSet[g] = true
	}

	// check client targeting variants
	expected := []string{
		"focused", "target focused", "switch focused", "control focused",
		"desktop", "target desktop", "switch desktop", "control desktop",
	}

	for _, exp := range expected {
		if !grammarSet[exp] {
			t.Errorf("grammar missing: %q", exp)
		}
	}
}

func TestVoskEngine_BuildGrammar_WithNumberedClient(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")
	v.AddClient("focused2")

	grammar := v.BuildGrammar()
	grammarSet := make(map[string]bool)
	for _, g := range grammar {
		grammarSet[g] = true
	}

	// check numbered client variants
	expected := []string{
		"focused2", "target focused2", "switch focused2", "control focused2",
		"focused two", "target focused two", "switch focused two", "control focused two",
	}

	for _, exp := range expected {
		if !grammarSet[exp] {
			t.Errorf("grammar missing: %q", exp)
		}
	}
}

func TestVoskEngine_BuildGrammar_WithAliases(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")
	v.SetAliases(map[string]string{
		"work":   "desktop",
		"coding": "focused",
	})

	grammar := v.BuildGrammar()
	grammarSet := make(map[string]bool)
	for _, g := range grammar {
		grammarSet[g] = true
	}

	expected := []string{
		"work", "target work", "switch work", "control work",
		"coding", "target coding", "switch coding", "control coding",
	}

	for _, exp := range expected {
		if !grammarSet[exp] {
			t.Errorf("grammar missing: %q", exp)
		}
	}
}

func TestVoskEngine_TranscribeNotAvailable(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")

	// without setup, should return empty
	result, err := v.Transcribe("/tmp/test.wav")
	if err != nil {
		t.Errorf("Transcribe error = %v, want nil", err)
	}
	if result != "" {
		t.Errorf("Transcribe result = %q, want empty", result)
	}
}

func TestDigitToWord(t *testing.T) {
	expected := map[string]string{
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

	if !reflect.DeepEqual(DigitToWord, expected) {
		t.Errorf("DigitToWord = %v, want %v", DigitToWord, expected)
	}
}

func TestBaseGrammar_ContainsEssentials(t *testing.T) {
	essentials := []string{
		"enter", "escape", "tab", "up", "down", "left", "right",
		"send", "submit", "scratch", "repeat",
		"pause commands", "resume commands",
	}

	grammarSet := make(map[string]bool)
	for _, g := range BaseGrammar {
		grammarSet[g] = true
	}

	for _, e := range essentials {
		if !grammarSet[e] {
			t.Errorf("BaseGrammar missing essential: %q", e)
		}
	}
}

func TestVoskEngine_ConcurrentAccess(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")

	// test concurrent add/remove
	done := make(chan bool, 4)

	go func() {
		for i := 0; i < 100; i++ {
			v.AddClient("client1")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			v.RemoveClient("client1")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			v.Available()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			v.BuildGrammar()
		}
		done <- true
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}

func TestVoskEngine_BuildGrammar_NoDuplicates(t *testing.T) {
	v := NewVoskEngine("/models/vosk-small")
	v.AddClient("focused")
	v.AddClient("focused") // add twice

	grammar := v.BuildGrammar()

	// count occurrences of "focused"
	count := 0
	for _, g := range grammar {
		if g == "focused" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("'focused' appears %d times, want 1", count)
	}
}

func TestVoskEngine_BuildGrammar_Sorted(t *testing.T) {
	// just verify grammar is valid (no need to be sorted, but should be consistent)
	v := NewVoskEngine("/models/vosk-small")
	v.AddClient("zebra")
	v.AddClient("alpha")

	grammar1 := v.BuildGrammar()
	grammar2 := v.BuildGrammar()

	sort.Strings(grammar1)
	sort.Strings(grammar2)

	if !reflect.DeepEqual(grammar1, grammar2) {
		t.Error("BuildGrammar should return consistent results")
	}
}
