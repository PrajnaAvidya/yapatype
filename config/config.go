package config

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
)

// SoundsConfig holds audio feedback settings
type SoundsConfig struct {
	Enabled               bool    `json:"enabled"`
	VoiceAcknowledgements bool    `json:"voice_acknowledgements"`
	CommandSuccess        *string `json:"command_success"`
	CommandWarning        *string `json:"command_warning"`
	Ready                 *string `json:"ready"`
}

// ServerConfig holds server settings
type ServerConfig struct {
	Host       string            `json:"host"`
	Port       int               `json:"port"`
	WhisperCLI *string           `json:"whisper_cli"`
	Model      string            `json:"model"`
	VoskModel  string            `json:"vosk_model"`
	Sounds     SoundsConfig      `json:"sounds"`
	Aliases    map[string]string `json:"client_aliases"`
	Microphone *string           `json:"microphone"`
}

// ClientConfig holds client settings
type ClientConfig struct {
	Name        string  `json:"name"`
	ServerURL   string  `json:"server_url"`
	Executor    string  `json:"executor"`
	KittySocket *string `json:"kitty_socket"`
}

// Config holds full configuration
type Config struct {
	Server ServerConfig `json:"server"`
	Client ClientConfig `json:"client"`
}

// DefaultSounds returns default sounds config
func DefaultSounds() SoundsConfig {
	return SoundsConfig{
		Enabled:               true,
		VoiceAcknowledgements: true,
		CommandSuccess:        nil,
		CommandWarning:        nil,
		Ready:                 nil,
	}
}

// DefaultServer returns default server config
func DefaultServer() ServerConfig {
	return ServerConfig{
		Host:       "0.0.0.0",
		Port:       9999,
		WhisperCLI: nil, // auto-detect
		Model:      "models/ggml-tiny.en.bin",
		VoskModel:  "models/vosk-model-small-en-us",
		Sounds:     DefaultSounds(),
		Aliases:    make(map[string]string),
		Microphone: nil,
	}
}

// DefaultClient returns default client config
func DefaultClient() ClientConfig {
	return ClientConfig{
		Name:        "", // auto-generate
		ServerURL:   "ws://localhost:9999",
		Executor:    "auto",
		KittySocket: nil,
	}
}

// Default returns a config with all default values
func Default() *Config {
	return &Config{
		Server: DefaultServer(),
		Client: DefaultClient(),
	}
}

// Load loads config from file, merging with defaults
func Load(path string) (*Config, error) {
	if path == "" {
		path = ConfigPath()
	}

	cfg := Default()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// apply post-init defaults and return
		cfg.applyDefaults()
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	// parse into raw structure to preserve existing defaults
	var raw struct {
		Server *json.RawMessage `json:"server"`
		Client *json.RawMessage `json:"client"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// merge server config
	if raw.Server != nil {
		if err := json.Unmarshal(*raw.Server, &cfg.Server); err != nil {
			return nil, err
		}
	}

	// merge client config
	if raw.Client != nil {
		if err := json.Unmarshal(*raw.Client, &cfg.Client); err != nil {
			return nil, err
		}
	}

	cfg.applyDefaults()
	return cfg, nil
}

// applyDefaults applies auto-detection and auto-generation after loading
func (c *Config) applyDefaults() {
	// auto-detect whisper-cli
	if c.Server.WhisperCLI == nil {
		cli := detectWhisperCLI()
		c.Server.WhisperCLI = &cli
	}

	// ensure aliases map is initialized
	if c.Server.Aliases == nil {
		c.Server.Aliases = make(map[string]string)
	}

	// auto-generate client name
	if c.Client.Name == "" {
		c.Client.Name = generateClientName(c.Client.Executor)
	}
}

// detectWhisperCLI finds the whisper-cli binary
func detectWhisperCLI() string {
	if GetPlatform() == "darwin" {
		return "/opt/homebrew/bin/whisper-cli"
	}
	// try to find in PATH
	if path, err := exec.LookPath("whisper-cli"); err == nil {
		return path
	}
	return "whisper-cli"
}

// generateClientName creates a default client name
func generateClientName(executor string) string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	// take first part of FQDN
	hostname = strings.Split(hostname, ".")[0]
	hostname = strings.ToLower(hostname)
	return hostname + "-" + executor
}
