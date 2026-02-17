package server

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/PrajnaAvidya/yapatype/config"
	"github.com/PrajnaAvidya/yapatype/protocol"
	"github.com/PrajnaAvidya/yapatype/server/audio"
	"github.com/PrajnaAvidya/yapatype/server/commands"
	"github.com/PrajnaAvidya/yapatype/server/sounds"
	"github.com/PrajnaAvidya/yapatype/server/transcribe"
)

// target switching pattern
var targetPattern = regexp.MustCompile(`(?i)^(?:target|switch|control)\s+(.+)$`)

// resume commands pattern for dictation mode exit
var resumePattern = regexp.MustCompile(`^(resume\s?commands?|start\s?commands?)$`)

// focus tab pattern for kitty tab switching
var focusTabPattern = regexp.MustCompile(`(?i)^focus\s+tab\s+(\w+)[.,!?]*$`)

// click sound pattern to skip
var clickPattern = regexp.MustCompile(`(?i)^click[, ]*\.?$`)

// default whisper prompt
const defaultPrompt = "Commands: newline, enter, send it, escape, tab, slash, backspace."

// audioChunk holds a recorded audio segment for processing
type audioChunk struct {
	path     string
	duration float64
}

// MainServer orchestrates audio capture, transcription, and dispatch
type MainServer struct {
	config        *config.ServerConfig
	audio         *audio.AudioCapture
	whisper       *transcribe.WhisperEngine
	vosk          *transcribe.VoskEngine
	sounds        *sounds.Player
	wsServer      *Server
	dictationMode bool
	prompt        string
	running       bool
	mu            sync.Mutex
}

// NewMainServer creates a new main server
// device is the resolved device name to use
// requiredDevice is the original configured device pattern (if set, device must remain available)
func NewMainServer(cfg *config.ServerConfig, device string, requiredDevice string) *MainServer {
	// create audio capture
	audioCapture := audio.NewAudioCapture(audio.DefaultAudioConfig(), device, requiredDevice)

	// create whisper engine
	whisperCLI := "whisper-cli"
	if cfg.WhisperCLI != nil {
		whisperCLI = *cfg.WhisperCLI
	}
	whisper := transcribe.NewWhisperEngine(whisperCLI, cfg.Model, 4)

	// create vosk engine if model configured
	var vosk *transcribe.VoskEngine
	if cfg.VoskModel != "" {
		vosk = transcribe.NewVoskEngine(cfg.VoskModel)
	}

	// create sounds player
	soundsCfg := sounds.DefaultConfig()
	soundsCfg.Enabled = cfg.Sounds.Enabled
	if cfg.Sounds.Ready != nil {
		soundsCfg.Ready = *cfg.Sounds.Ready
	}
	if cfg.Sounds.CommandSuccess != nil {
		soundsCfg.CommandSuccess = *cfg.Sounds.CommandSuccess
	}
	if cfg.Sounds.CommandWarning != nil {
		soundsCfg.CommandWarning = *cfg.Sounds.CommandWarning
	}
	soundsCfg.VoiceAcknowledgements = cfg.Sounds.VoiceAcknowledgements
	player := sounds.NewPlayer(soundsCfg)

	// create websocket server
	wsServer := NewServer(cfg.Host, cfg.Port, cfg.Aliases)

	return &MainServer{
		config:   cfg,
		audio:    audioCapture,
		whisper:  whisper,
		vosk:     vosk,
		sounds:   player,
		wsServer: wsServer,
		prompt:   defaultPrompt,
	}
}

// Run starts the main server
func (s *MainServer) Run(ctx context.Context) error {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	// setup vosk if available
	if s.vosk != nil {
		s.vosk.SetAliases(s.config.Aliases)
		if err := s.vosk.Setup(); err != nil {
			fmt.Printf("vosk not available: %v\n", err)
			s.vosk = nil
		} else {
			fmt.Println("vosk ready for fast command recognition")
		}
	}

	// setup whisper prompt
	s.whisper.SetPrompt(s.prompt)

	// register callbacks
	s.setupCallbacks()

	// start websocket server in background
	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()
	go func() {
		if err := s.wsServer.Run(wsCtx); err != nil {
			fmt.Printf("websocket server error: %v\n", err)
		}
	}()

	// play ready sound
	s.sounds.PlayReady()

	// get device info for display
	deviceInfo := s.audio.Device
	if deviceInfo == "" {
		deviceInfo, _ = audio.GetDefaultInputDevice(ctx)
	}
	fmt.Printf("audio: %s\n", deviceInfo)
	fmt.Println("listening... (pause for 0.7s to transcribe)")
	fmt.Println("---")

	// main audio loop
	return s.audioLoop(ctx)
}

// Stop stops the server
func (s *MainServer) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	s.audio.Cancel()
	s.wsServer.Stop()

	if s.vosk != nil {
		s.vosk.Cleanup()
	}
}

// setupCallbacks registers callbacks with websocket server
func (s *MainServer) setupCallbacks() {
	s.wsServer.Manager.OnPromptReceived = func(prompt string) {
		s.mu.Lock()
		s.prompt = prompt
		s.mu.Unlock()
		s.whisper.SetPrompt(prompt)
		fmt.Printf("whisper prompt updated (%d chars)\n", len(prompt))
	}

	if s.vosk != nil {
		s.wsServer.Manager.OnClientRegistered = func(name string) {
			s.vosk.AddClient(name)
		}
		s.wsServer.Manager.OnClientRemoved = func(name string) {
			s.vosk.RemoveClient(name)
		}
	}
}

// audioLoop orchestrates parallel recording and transcription
func (s *MainServer) audioLoop(ctx context.Context) error {
	audioChan := make(chan audioChunk, 5) // buffer for bursts

	// start recording goroutine - runs continuously
	go s.recordLoop(ctx, audioChan)

	// run transcription in main goroutine
	s.transcribeLoop(ctx, audioChan)
	return nil
}

// recordLoop continuously captures audio, sending chunks to the channel
func (s *MainServer) recordLoop(ctx context.Context, audioChan chan<- audioChunk) {
	defer close(audioChan)

	for {
		s.mu.Lock()
		running := s.running
		s.mu.Unlock()

		if !running {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		audioPath, duration, err := s.audio.Record(ctx)
		if err != nil {
			// if required device is unavailable, wait for it to come back
			if err == audio.ErrDeviceUnavailable {
				s.sounds.PlayCommandWarning()
				fmt.Printf("configured microphone unavailable, waiting...\n")
				device, waitErr := WaitForMicrophone(ctx, s.audio.RequiredDevice)
				if waitErr != nil {
					return
				}
				s.audio.Device = device
				fmt.Printf("microphone reconnected: %s\n", device)
				s.sounds.PlayReady()
				continue
			}
			fmt.Printf("audio error: %v\n", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if audioPath != "" {
			select {
			case audioChan <- audioChunk{path: audioPath, duration: duration}:
			case <-ctx.Done():
				s.audio.Cleanup(audioPath)
				return
			}
		}
	}
}

// transcribeLoop processes audio chunks from the channel
func (s *MainServer) transcribeLoop(ctx context.Context, audioChan <-chan audioChunk) {
	for chunk := range audioChan {
		select {
		case <-ctx.Done():
			s.audio.Cleanup(chunk.path)
			return
		default:
		}

		s.processAudio(ctx, chunk.path, chunk.duration)
	}
}

// processAudio handles transcription and dispatch for recorded audio
func (s *MainServer) processAudio(ctx context.Context, audioPath string, duration float64) {
	// skip very short audio
	if duration < 0.1 {
		s.audio.Cleanup(audioPath)
		return
	}

	fmt.Printf("[audio] %.2fs captured", duration)

	var transcription, modelUsed string
	voskHeardSpeech := false

	// for short utterances, try vosk first (fast command check)
	if duration < 1.2 && s.vosk != nil && s.vosk.Available() {
		voskText, _ := s.vosk.Transcribe(audioPath)
		if voskText != "" {
			voskHeardSpeech = true
			// accept if it's a known command OR a valid target
			if commands.IsKnownCommand(voskText) || s.wsServer.Manager.IsValidTarget(voskText) {
				transcription = voskText
				modelUsed = "vosk"
			}
		}
	}

	// continuation check if vosk didn't hear speech
	if transcription == "" && duration < 1.2 && !voskHeardSpeech {
		quickText, _ := s.whisper.QuickTranscribe(ctx, audioPath)
		if quickText != "" {
			isTarget := s.wsServer.Manager.IsValidTarget(quickText)
			if !commands.IsKnownCommand(quickText) && !isTarget {
				// not a command or target, wait for possible continuation
				contPath, _ := s.audio.RecordContinuation(ctx, 1.2)
				if contPath != "" {
					combinedPath := "/tmp/yapatype-combined.wav"
					if err := s.audio.Concatenate(audioPath, contPath, combinedPath); err == nil {
						s.audio.Cleanup(audioPath)
						s.audio.Cleanup(contPath)
						audioPath = combinedPath
					} else {
						s.audio.Cleanup(contPath)
					}
				}
			}
		}
	}

	// full whisper transcription if needed
	if transcription == "" {
		var err error
		transcription, err = s.whisper.Transcribe(ctx, audioPath)
		if err != nil {
			fmt.Printf(" - transcription error: %v\n", err)
		}
		modelUsed = "whisper"
	}

	s.audio.Cleanup(audioPath)

	if transcription == "" {
		fmt.Println(" - no speech detected")
		return
	}

	// skip single letters
	if len(transcription) == 1 && unicode.IsLetter(rune(transcription[0])) {
		fmt.Printf(" - skipped single letter: %s\n", transcription)
		return
	}

	// skip keyboard clicks
	if clickPattern.MatchString(transcription) {
		fmt.Println(" - skipped keyboard clicks")
		return
	}

	// skip whisper hallucinations (only short segments)
	if transcribe.IsHallucination(transcription, 6) {
		fmt.Println(" - skipped whisper hallucination")
		return
	}

	fmt.Println() // newline after "captured"
	modeIndicator := ""
	if s.dictationMode {
		modeIndicator = "[dictation] "
	}
	fmt.Printf(">> %s[%s] %s\n", modeIndicator, modelUsed, transcription)

	s.handleTranscription(transcription)
}

// handleTranscription processes transcribed text
func (s *MainServer) handleTranscription(text string) {
	// dictation mode check first
	if s.dictationMode {
		matchText := strings.ToLower(strings.TrimSpace(text))
		matchText = strings.TrimRight(matchText, ".,!?")
		if resumePattern.MatchString(matchText) {
			// exit dictation mode
			s.dictationMode = false
			fmt.Println("   [dictation mode OFF]")
			s.sounds.Say("commands active")
			return
		}
		// type as-is with trailing space
		s.sendToTarget(protocol.NewTypeText(text + " "))
		return
	}

	// normalize for matching: remove commas, collapse spaces
	normalized := strings.ReplaceAll(text, ",", "")
	normalized = strings.TrimSpace(normalized)
	normalized = strings.Join(strings.Fields(normalized), " ")

	// check for target switching command
	if match := targetPattern.FindStringSubmatch(normalized); len(match) > 1 {
		// remove spaces/punctuation from target name
		targetName := strings.ReplaceAll(match[1], " ", "")
		targetName = strings.TrimRight(targetName, ".,!?")
		if s.wsServer.Manager.SetTarget(targetName) {
			s.sounds.Say("targeting " + targetName)
		} else {
			fmt.Printf("   unknown target: %s\n", targetName)
			fmt.Printf("   available: %v\n", s.wsServer.Manager.ListClients())
			s.sounds.PlayCommandWarning()
		}
		return
	}

	// two-word fallback: "X client-name" -> target client-name
	words := strings.Fields(normalized)
	if len(words) == 2 {
		secondWord := strings.TrimRight(words[1], ".,!?")
		if s.wsServer.Manager.IsValidTarget(secondWord) {
			if s.wsServer.Manager.SetTarget(secondWord) {
				s.sounds.Say("targeting " + secondWord)
				return
			}
		}
	}

	// bare client name or alias
	bareName := strings.ReplaceAll(normalized, " ", "")
	bareName = strings.TrimRight(bareName, ".,!?")
	if s.wsServer.Manager.IsValidTarget(bareName) {
		if s.wsServer.Manager.SetTarget(bareName) {
			s.sounds.Say("targeting " + bareName)
			return
		}
	}

	// check for focus tab command
	if match := focusTabPattern.FindStringSubmatch(normalized); len(match) > 1 {
		tabStr := ConvertNumberWords(match[1])
		if tabNum, err := strconv.Atoi(tabStr); err == nil && tabNum > 0 {
			if err := s.sendToTarget(protocol.NewFocusTab(tabNum)); err != nil {
				fmt.Println("   (no target connected)")
				s.sounds.PlayCommandWarning()
			} else {
				fmt.Printf("   focus tab %d\n", tabNum)
				s.sounds.Say(fmt.Sprintf("tab %d", tabNum))
			}
			return
		}
	}

	// parse and dispatch
	result := commands.ParseText(text)
	s.dispatch(result)
}

// dispatch sends parsed actions to target
func (s *MainServer) dispatch(result commands.ParseResult) {
	if len(result.Actions) == 0 {
		return
	}

	// handle special actions first
	for _, action := range result.Actions {
		switch action.(type) {
		case commands.ScratchAction:
			count := s.wsServer.Manager.PopLastTextLength()
			if count > 0 {
				s.sendToTarget(protocol.NewSendKey(protocol.KeyBackspace, nil, count))
				s.sounds.PlayCommandSuccess()
			} else {
				fmt.Println("   (nothing to scratch)")
				s.sounds.PlayCommandWarning()
			}
			return

		case commands.RepeatAction:
			if last := s.wsServer.Manager.LastTypeText(); last != nil {
				s.sendToTarget(last)
				s.sounds.PlayCommandSuccess()
			} else {
				fmt.Println("   (nothing to repeat)")
				s.sounds.PlayCommandWarning()
			}
			return

		case commands.PauseCommandsAction:
			s.dictationMode = true
			fmt.Println("   [dictation mode ON]")
			s.sounds.Say("dictation mode")
			return

		case commands.ResumeCommandsAction:
			s.dictationMode = false
			fmt.Println("   [dictation mode OFF]")
			s.sounds.Say("commands active")
			return
		}
	}

	// play sound based on result type
	isShort := len(result.OriginalText) <= 15
	if result.WasCommand {
		s.sounds.PlayCommandSuccess()
	} else if isShort {
		// short text that wasn't a command
		s.sounds.PlayCommandWarning()
	}

	// send actions to target
	sentText := false
	for _, action := range result.Actions {
		// 50ms delay before enter if we just sent text (ydotool timing issue)
		if sk, ok := action.(*protocol.SendKey); ok && sk.Key == protocol.KeyEnter && sentText {
			time.Sleep(50 * time.Millisecond)
		}

		if err := s.sendToTarget(action); err != nil {
			fmt.Println("   (no target connected)")
			break
		}

		if _, ok := action.(*protocol.TypeText); ok {
			sentText = true
		}
	}
}

// sendToTarget sends a message to the active target
func (s *MainServer) sendToTarget(msg any) error {
	return s.wsServer.SendToTarget(msg)
}

// WaitForMicrophone waits for a microphone matching target to become available
func WaitForMicrophone(ctx context.Context, target string) (string, error) {
	printed := false
	for {
		device, _ := audio.FindMatchingDevice(ctx, target)
		if device != "" {
			return device, nil
		}

		if !printed {
			fmt.Printf("waiting for microphone matching '%s'...\n", target)
			printed = true
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
