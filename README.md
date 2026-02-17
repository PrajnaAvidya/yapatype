# yapatype

Voice-driven development interface.

## What is this?

A tool for hands-free coding sessions with Claude and other LLMs. Modern AI-assisted development is increasingly conversational - you describe intent, discuss tradeoffs, and iterate through dialogue. Instead of typing all day, speak naturally and let the LLM do the heavy lifting.

Runs as a server/client architecture over WebSocket. The server captures audio, transcribes speech, parses commands, and broadcasts to connected clients. Clients execute keystrokes and text input.

## Installation

Download the latest archive for your platform from [GitHub Releases](https://github.com/PrajnaAvidya/yapatype/releases):

- `yapatype-linux-amd64.tar.gz` - Linux
- `yapatype-windows-amd64.zip` - Windows
- `yapatype-darwin-arm64.tar.gz` - macOS (Apple Silicon)

Extract and run. Or build from source: `go build -o yapatype .`

### macOS Gatekeeper

macOS blocks unsigned binaries. After extracting, remove the quarantine attribute:

```bash
xattr -d com.apple.quarantine yapatype
```

Or right-click the binary → Open → Open Anyway.

### System Dependencies

#### macOS

```bash
# audio capture
brew install sox

# whisper.cpp (speech transcription)
brew install whisper-cpp
```

#### Linux

```bash
# audio capture + keystroke injection
sudo apt install sox ydotool   # Ubuntu/Debian
sudo dnf install sox ydotool   # Fedora

# whisper.cpp (build from source)
git clone https://github.com/ggerganov/whisper.cpp
cd whisper.cpp
make
sudo cp main /usr/local/bin/whisper-cli
```

#### Optional: kitty terminal

If you want to send keystrokes to an unfocused [kitty](https://sw.kovidgoyal.net/kitty/) terminal, start kitty with remote control enabled:

```bash
kitty -o allow_remote_control=yes --listen-on unix:/tmp/mykitty.sock
```

### Models

yapatype requires speech recognition models. Use the Makefile for easy setup:

```bash
# Download whisper model (required) - options: tiny, small, base, medium
make setup-whisper WHISPER_MODEL=small

# Download vosk library + model (optional, macOS only)
make setup-vosk

# Or download both at once
make setup WHISPER_MODEL=small
```

**Whisper models** (English-only, from [whisper.cpp](https://huggingface.co/ggerganov/whisper.cpp)):

| Model | Size | Notes |
|-------|------|-------|
| tiny.en | 75MB | Fastest, lower accuracy |
| base.en | 142MB | Balanced |
| small.en | 466MB | Recommended |
| medium.en | 1.5GB | Best accuracy, slower |

**Vosk** enables fast command recognition for short utterances (<1.2s). Without vosk, all transcription uses whisper (slightly slower for commands but works fine).

## Usage

### Server

Captures audio, transcribes speech, and broadcasts to clients.

```bash
yapatype server [flags]

Flags:
  -p, --port int       websocket port (default: 9999)
      --host string    websocket host (default: 0.0.0.0)
      --mic string     microphone to use (substring match)
      --no-sound       disable audio feedback
      --no-voice       disable voice acknowledgements (chime only, macOS)
  -c, --config string  config file path
```

### Client

Connects to server and executes commands via keystroke injection.

```bash
yapatype client [flags]

Flags:
  -n, --name string          client name (default: hostname-executor)
  -s, --server string        server url (default: ws://localhost:9999)
  -e, --executor string      input method: auto, ydotool, kitty, osascript
  -k, --kitty-socket string  kitty socket path (implies kitty executor)
  -c, --config string        config file path
```

### Multiple Clients

Run multiple clients on the same machine with different executors:

```bash
# terminal 1: typing to focused window
yapatype client --name desktop-ydotool --executor ydotool

# terminal 2: typing to unfocused kitty terminal
yapatype client --name desktop-kitty --kitty-socket /tmp/mykitty.sock
```

Switch targets with voice: "target desktop-kitty" or "switch desktop-ydotool"

### Vocabulary Generation

Generate project-specific whisper prompts for better transcription:

```bash
yapatype vocab [project_path]
```

Creates a `.whisper-prompt` file that clients automatically load and send to the server.

## Voice Commands

**Short utterances (fuzzy matched):** enter, escape, tab, up, down, left, right, cancel, scratch, repeat

**Target switching:** "target [name]" or "switch [name]"

**Inline:** slash, backslash

**End-of-text:** send, send it, hit enter, press enter, submit

**Tab switching (kitty only):** "focus tab 1", "focus tab 2", etc. Switches which kitty tab receives keystrokes. Supports number words ("focus tab two"). Can be combined with target switching: "target focused tab 2". All subsequent input goes to that tab regardless of visual focus.

**Dictation mode:** "pause commands" to enter pure dictation, "resume commands" to exit

## Configuration

Optional config at `~/.config/yapatype/config.json`:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 9999,
    "whisper_cli": "/usr/bin/whisper-cli",
    "model": "models/ggml-small.en.bin",
    "vosk_model": "models/vosk-model-small-en-us",
    "sounds": { "enabled": true, "voice_acknowledgements": true },
    "client_aliases": { "focused": "desktop-ydotool" }
  },
  "client": {
    "name": "desktop",
    "server_url": "ws://localhost:9999",
    "executor": "auto"
  }
}
```

## Development

```bash
# run tests
make test

# build without vosk (whisper only)
make build

# build with vosk support (macOS)
# requires: make setup-vosk first
make build-vosk
```

The vosk build handles CGO flags and fixes the runtime library path automatically. If you build manually with `go build -tags vosk`, you'll need to set `CGO_CFLAGS`, `CGO_LDFLAGS`, and fix the dylib rpath with `install_name_tool`.
