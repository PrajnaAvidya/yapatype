# yapatype

Voice-driven development interface.

## What is this?

A tool for hands-free coding sessions with Claude and other LLMs. Modern AI-assisted development is increasingly conversational - you describe intent, discuss tradeoffs, and iterate through dialogue. Instead of typing all day, speak naturally and let the LLM do the heavy lifting.

Runs as a server/client architecture over WebSocket. The server captures audio, transcribes speech, parses commands, and broadcasts to connected clients. Clients execute keystrokes and text input.

## Installation

```bash
go build -o yapatype .
```

### System Dependencies

- **sox**: audio capture
- **whisper-cli**: speech transcription (whisper.cpp)
- **vosk** (optional): fast recognition for short commands
- **ydotool** (linux): keystroke injection
- **kitty** (linux, optional): kitty terminal remote control

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

**Dictation mode:** "pause commands" to enter pure dictation, "resume commands" to exit

## Configuration

Optional config at `~/.config/yapatype/config.json`:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 9999,
    "whisper_cli": "/usr/bin/whisper-cli",
    "model": "models/ggml-tiny.en.bin",
    "vosk_model": "models/vosk-model-small-en-us",
    "sounds": { "enabled": true },
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
go test ./...

# build
go build -o yapatype .
```
