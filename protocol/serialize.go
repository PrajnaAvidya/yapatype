package protocol

import (
	"encoding/json"
	"fmt"
)

// Serialize converts a message to JSON bytes
func Serialize(msg any) ([]byte, error) {
	return json.Marshal(msg)
}

// ParseServerMessage parses JSON into a server->client message
func ParseServerMessage(data []byte) (any, error) {
	var base struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("parse message type: %w", err)
	}

	switch base.Type {
	case "type":
		var msg TypeText
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("parse TypeText: %w", err)
		}
		return &msg, nil

	case "key":
		var msg sendKeyRaw
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("parse SendKey: %w", err)
		}
		// apply defaults
		if msg.Modifiers == nil {
			msg.Modifiers = []Modifier{}
		}
		if msg.Repeat == 0 {
			msg.Repeat = 1
		}
		return &SendKey{
			Type:      msg.Type,
			Key:       msg.Key,
			Modifiers: msg.Modifiers,
			Repeat:    msg.Repeat,
		}, nil

	case "ping":
		return &Ping{Type: "ping"}, nil

	case "target_status":
		var msg TargetStatus
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("parse TargetStatus: %w", err)
		}
		return &msg, nil

	case "registered":
		var msg Registered
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("parse Registered: %w", err)
		}
		return &msg, nil

	default:
		return nil, fmt.Errorf("unknown server message type: %s", base.Type)
	}
}

// ParseClientMessage parses JSON into a client->server message
func ParseClientMessage(data []byte) (any, error) {
	var base struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("parse message type: %w", err)
	}

	switch base.Type {
	case "register":
		var msg Register
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("parse Register: %w", err)
		}
		return &msg, nil

	case "prompt":
		var msg Prompt
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("parse Prompt: %w", err)
		}
		return &msg, nil

	case "pong":
		return &Pong{Type: "pong"}, nil

	default:
		return nil, fmt.Errorf("unknown client message type: %s", base.Type)
	}
}

// sendKeyRaw is used for parsing to handle missing/default fields
type sendKeyRaw struct {
	Type      string     `json:"type"`
	Key       Key        `json:"key"`
	Modifiers []Modifier `json:"modifiers"`
	Repeat    int        `json:"repeat"`
}
