package protocol

import (
	"encoding/json"
	"testing"
)

func TestSerializeTypeText(t *testing.T) {
	msg := NewTypeText("hello world")
	data, err := Serialize(msg)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["type"] != "type" {
		t.Errorf("type = %v, want 'type'", m["type"])
	}
	if m["text"] != "hello world" {
		t.Errorf("text = %v, want 'hello world'", m["text"])
	}
}

func TestSerializeSendKey(t *testing.T) {
	msg := NewSendKey(KeyEnter, []Modifier{ModCtrl, ModShift}, 2)
	data, err := Serialize(msg)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["type"] != "key" {
		t.Errorf("type = %v, want 'key'", m["type"])
	}
	if m["key"] != "enter" {
		t.Errorf("key = %v, want 'enter'", m["key"])
	}
	if m["repeat"] != float64(2) {
		t.Errorf("repeat = %v, want 2", m["repeat"])
	}

	mods := m["modifiers"].([]any)
	if len(mods) != 2 {
		t.Errorf("modifiers len = %d, want 2", len(mods))
	}
}

func TestSerializePing(t *testing.T) {
	msg := NewPing()
	data, err := Serialize(msg)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["type"] != "ping" {
		t.Errorf("type = %v, want 'ping'", m["type"])
	}
}

func TestSerializeTargetStatus(t *testing.T) {
	msg := NewTargetStatus(true)
	data, err := Serialize(msg)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["type"] != "target_status" {
		t.Errorf("type = %v, want 'target_status'", m["type"])
	}
	if m["is_active"] != true {
		t.Errorf("is_active = %v, want true", m["is_active"])
	}
}

func TestSerializeRegistered(t *testing.T) {
	msg := NewRegistered("client1")
	data, err := Serialize(msg)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["type"] != "registered" {
		t.Errorf("type = %v, want 'registered'", m["type"])
	}
	if m["name"] != "client1" {
		t.Errorf("name = %v, want 'client1'", m["name"])
	}
}

func TestSerializeRegister(t *testing.T) {
	msg := NewRegister("myclient", "linux")
	data, err := Serialize(msg)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["type"] != "register" {
		t.Errorf("type = %v, want 'register'", m["type"])
	}
	if m["name"] != "myclient" {
		t.Errorf("name = %v, want 'myclient'", m["name"])
	}
	if m["platform"] != "linux" {
		t.Errorf("platform = %v, want 'linux'", m["platform"])
	}
}

func TestSerializePrompt(t *testing.T) {
	msg := NewPrompt("technical terms: asyncio, websocket")
	data, err := Serialize(msg)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["type"] != "prompt" {
		t.Errorf("type = %v, want 'prompt'", m["type"])
	}
	if m["prompt"] != "technical terms: asyncio, websocket" {
		t.Errorf("prompt = %v, want 'technical terms: asyncio, websocket'", m["prompt"])
	}
}

func TestSerializePong(t *testing.T) {
	msg := NewPong()
	data, err := Serialize(msg)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["type"] != "pong" {
		t.Errorf("type = %v, want 'pong'", m["type"])
	}
}

func TestParseServerMessages(t *testing.T) {
	tests := []struct {
		name string
		json string
		want any
	}{
		{
			name: "TypeText",
			json: `{"type":"type","text":"hello"}`,
			want: &TypeText{Type: "type", Text: "hello"},
		},
		{
			name: "SendKey full",
			json: `{"type":"key","key":"enter","modifiers":["ctrl"],"repeat":2}`,
			want: &SendKey{Type: "key", Key: KeyEnter, Modifiers: []Modifier{ModCtrl}, Repeat: 2},
		},
		{
			name: "SendKey defaults",
			json: `{"type":"key","key":"escape"}`,
			want: &SendKey{Type: "key", Key: KeyEscape, Modifiers: []Modifier{}, Repeat: 1},
		},
		{
			name: "Ping",
			json: `{"type":"ping"}`,
			want: &Ping{Type: "ping"},
		},
		{
			name: "TargetStatus",
			json: `{"type":"target_status","is_active":true}`,
			want: &TargetStatus{Type: "target_status", IsActive: true},
		},
		{
			name: "Registered",
			json: `{"type":"registered","name":"client1"}`,
			want: &Registered{Type: "registered", Name: "client1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseServerMessage([]byte(tc.json))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tc.want)

			if string(gotJSON) != string(wantJSON) {
				t.Errorf("got %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestParseClientMessages(t *testing.T) {
	tests := []struct {
		name string
		json string
		want any
	}{
		{
			name: "Register",
			json: `{"type":"register","name":"client","platform":"darwin"}`,
			want: &Register{Type: "register", Name: "client", Platform: "darwin"},
		},
		{
			name: "Prompt",
			json: `{"type":"prompt","prompt":"test prompt"}`,
			want: &Prompt{Type: "prompt", Prompt: "test prompt"},
		},
		{
			name: "Pong",
			json: `{"type":"pong"}`,
			want: &Pong{Type: "pong"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseClientMessage([]byte(tc.json))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tc.want)

			if string(gotJSON) != string(wantJSON) {
				t.Errorf("got %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	messages := []any{
		NewTypeText("test text"),
		NewSendKey(KeyTab, []Modifier{ModAlt}, 3),
		NewPing(),
		NewTargetStatus(false),
		NewRegistered("myname"),
		NewRegister("client", "linux"),
		NewPrompt("vocab prompt"),
		NewPong(),
	}

	serverMsgs := messages[:5]
	clientMsgs := messages[5:]

	for _, msg := range serverMsgs {
		data, err := Serialize(msg)
		if err != nil {
			t.Fatalf("serialize %T: %v", msg, err)
		}

		parsed, err := ParseServerMessage(data)
		if err != nil {
			t.Fatalf("parse %T: %v", msg, err)
		}

		origJSON, _ := json.Marshal(msg)
		parsedJSON, _ := json.Marshal(parsed)

		if string(origJSON) != string(parsedJSON) {
			t.Errorf("round trip %T: got %s, want %s", msg, parsedJSON, origJSON)
		}
	}

	for _, msg := range clientMsgs {
		data, err := Serialize(msg)
		if err != nil {
			t.Fatalf("serialize %T: %v", msg, err)
		}

		parsed, err := ParseClientMessage(data)
		if err != nil {
			t.Fatalf("parse %T: %v", msg, err)
		}

		origJSON, _ := json.Marshal(msg)
		parsedJSON, _ := json.Marshal(parsed)

		if string(origJSON) != string(parsedJSON) {
			t.Errorf("round trip %T: got %s, want %s", msg, parsedJSON, origJSON)
		}
	}
}

func TestParseUnknownType(t *testing.T) {
	_, err := ParseServerMessage([]byte(`{"type":"unknown"}`))
	if err == nil {
		t.Error("expected error for unknown type")
	}

	_, err = ParseClientMessage([]byte(`{"type":"unknown"}`))
	if err == nil {
		t.Error("expected error for unknown type")
	}
}
