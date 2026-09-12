package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultValues(t *testing.T) {
	cfg := Default()

	// server defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want '0.0.0.0'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("Server.Port = %d, want 9999", cfg.Server.Port)
	}
	if cfg.Server.Model != "models/ggml-small.en.bin" {
		t.Errorf("Server.Model = %q, want 'models/ggml-small.en.bin'", cfg.Server.Model)
	}
	if cfg.Server.VoskModel != "models/vosk-model-small-en-us" {
		t.Errorf("Server.VoskModel = %q, want 'models/vosk-model-small-en-us'", cfg.Server.VoskModel)
	}
	if !cfg.Server.Sounds.Enabled {
		t.Error("Server.Sounds.Enabled = false, want true")
	}

	// client defaults
	if cfg.Client.ServerURL != "ws://localhost:9999" {
		t.Errorf("Client.ServerURL = %q, want 'ws://localhost:9999'", cfg.Client.ServerURL)
	}
	if cfg.Client.Executor != "auto" {
		t.Errorf("Client.Executor = %q, want 'auto'", cfg.Client.Executor)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// should return defaults with auto-detection applied
	if cfg.Server.Port != 9999 {
		t.Errorf("Server.Port = %d, want 9999", cfg.Server.Port)
	}
	if cfg.Server.WhisperCLI == nil {
		t.Error("Server.WhisperCLI should be auto-detected")
	}
	if cfg.Client.Name == "" {
		t.Error("Client.Name should be auto-generated")
	}
}

func TestLoadPartialConfig(t *testing.T) {
	// create temp config file with partial config
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// only override port, rest should use defaults
	data := []byte(`{"server": {"port": 8888}}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// port should be overridden
	if cfg.Server.Port != 8888 {
		t.Errorf("Server.Port = %d, want 8888", cfg.Server.Port)
	}

	// host should use default
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want '0.0.0.0'", cfg.Server.Host)
	}

	// client should use defaults
	if cfg.Client.ServerURL != "ws://localhost:9999" {
		t.Errorf("Client.ServerURL = %q, want 'ws://localhost:9999'", cfg.Client.ServerURL)
	}
}

func TestLoadFullConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := []byte(`{
		"server": {
			"host": "127.0.0.1",
			"port": 7777,
			"whisper_cli": "/usr/bin/whisper",
			"model": "custom-model.bin",
			"vosk_model": "custom-vosk",
			"sounds": {
				"enabled": false,
				"command_success": "/path/to/success.wav"
			},
			"client_aliases": {"foo": "bar"},
			"microphone": "Blue Yeti"
		},
		"client": {
			"name": "myclient",
			"server_url": "ws://remote:9999",
			"executor": "kitty",
			"kitty_socket": "/tmp/kitty.sock"
		}
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	// server
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %q, want '127.0.0.1'", cfg.Server.Host)
	}
	if cfg.Server.Port != 7777 {
		t.Errorf("Server.Port = %d, want 7777", cfg.Server.Port)
	}
	if cfg.Server.WhisperCLI == nil || *cfg.Server.WhisperCLI != "/usr/bin/whisper" {
		t.Errorf("Server.WhisperCLI = %v, want '/usr/bin/whisper'", cfg.Server.WhisperCLI)
	}
	if cfg.Server.Model != filepath.Join(cwd, "custom-model.bin") {
		t.Errorf("Server.Model = %q, want %q", cfg.Server.Model, filepath.Join(cwd, "custom-model.bin"))
	}
	if cfg.Server.VoskModel != filepath.Join(cwd, "custom-vosk") {
		t.Errorf("Server.VoskModel = %q, want %q", cfg.Server.VoskModel, filepath.Join(cwd, "custom-vosk"))
	}
	if cfg.Server.Sounds.Enabled {
		t.Error("Server.Sounds.Enabled = true, want false")
	}
	if cfg.Server.Sounds.CommandSuccess == nil || *cfg.Server.Sounds.CommandSuccess != "/path/to/success.wav" {
		t.Errorf("Server.Sounds.CommandSuccess = %v, want '/path/to/success.wav'", cfg.Server.Sounds.CommandSuccess)
	}
	if cfg.Server.Aliases["foo"] != "bar" {
		t.Errorf("Server.Aliases['foo'] = %q, want 'bar'", cfg.Server.Aliases["foo"])
	}
	if cfg.Server.Microphone == nil || *cfg.Server.Microphone != "Blue Yeti" {
		t.Errorf("Server.Microphone = %v, want 'Blue Yeti'", cfg.Server.Microphone)
	}

	// client
	if cfg.Client.Name != "myclient" {
		t.Errorf("Client.Name = %q, want 'myclient'", cfg.Client.Name)
	}
	if cfg.Client.ServerURL != "ws://remote:9999" {
		t.Errorf("Client.ServerURL = %q, want 'ws://remote:9999'", cfg.Client.ServerURL)
	}
	if cfg.Client.Executor != "kitty" {
		t.Errorf("Client.Executor = %q, want 'kitty'", cfg.Client.Executor)
	}
	if cfg.Client.KittySocket == nil || *cfg.Client.KittySocket != "/tmp/kitty.sock" {
		t.Errorf("Client.KittySocket = %v, want '/tmp/kitty.sock'", cfg.Client.KittySocket)
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	if got := ExpandPath(""); got != "" {
		t.Errorf("ExpandPath('') = %q, want ''", got)
	}

	if got := ExpandPath("~/models/model.bin"); got != filepath.Join(home, "models", "model.bin") {
		t.Errorf("ExpandPath('~/models/model.bin') = %q, want %q", got, filepath.Join(home, "models", "model.bin"))
	}

	if got := ExpandPath("/tmp/../tmp/model.bin"); got != "/tmp/model.bin" {
		t.Errorf("ExpandPath('/tmp/../tmp/model.bin') = %q, want '/tmp/model.bin'", got)
	}

	// existing relative paths resolve against cwd
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "model.bin"), []byte("x"), 0644); err != nil {
		t.Fatalf("write model: %v", err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldCwd)
	}()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	if got := ExpandPath("model.bin"); got != filepath.Join(cwd, "model.bin") {
		t.Errorf("ExpandPath('model.bin') = %q, want %q", got, filepath.Join(cwd, "model.bin"))
	}

	// missing relative paths still anchor to cwd
	if got := ExpandPath("missing.bin"); got != filepath.Join(cwd, "missing.bin") {
		t.Errorf("ExpandPath('missing.bin') = %q, want %q", got, filepath.Join(cwd, "missing.bin"))
	}
}

func TestLoadResolvesModelPaths(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "models"), 0755); err != nil {
		t.Fatalf("mkdir models: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "models", "ggml-small.en.bin"), []byte("x"), 0644); err != nil {
		t.Fatalf("write model: %v", err)
	}

	cfgPath := filepath.Join(dir, "config.json")
	data := []byte(`{"server": {"model": "models/ggml-small.en.bin", "vosk_model": "models/vosk"}}`)
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldCwd)
	}()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if want := filepath.Join(cwd, "models", "ggml-small.en.bin"); cfg.Server.Model != want {
		t.Errorf("Server.Model = %q, want %q", cfg.Server.Model, want)
	}
	if want := filepath.Join(cwd, "models", "vosk"); cfg.Server.VoskModel != want {
		t.Errorf("Server.VoskModel = %q, want %q", cfg.Server.VoskModel, want)
	}
}

func TestGetPlatform(t *testing.T) {
	p := GetPlatform()
	if p != "darwin" && p != "linux" {
		t.Errorf("GetPlatform() = %q, want 'darwin' or 'linux'", p)
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Error("ConfigDir() returned empty string")
	}
	// should end with .config/yapatype
	if filepath.Base(filepath.Dir(dir)) != ".config" || filepath.Base(dir) != "yapatype" {
		t.Errorf("ConfigDir() = %q, should end with .config/yapatype", dir)
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if filepath.Base(path) != "config.json" {
		t.Errorf("ConfigPath() = %q, should end with config.json", path)
	}
}

func TestStatePath(t *testing.T) {
	path := StatePath()
	if filepath.Base(path) != ".state" {
		t.Errorf("StatePath() = %q, should end with .state", path)
	}
}

func TestGenerateClientName(t *testing.T) {
	name := generateClientName("kitty")
	if name == "" {
		t.Error("generateClientName returned empty string")
	}
	// should end with -kitty
	if len(name) < 7 || name[len(name)-6:] != "-kitty" {
		t.Errorf("generateClientName('kitty') = %q, should end with '-kitty'", name)
	}
}
