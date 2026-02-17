package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/PrajnaAvidya/yapatype/protocol"
	"github.com/gorilla/websocket"
)

// mockExecutor records calls for testing
type mockExecutor struct {
	name      string
	typeCalls []string
	keyCalls  []keyCall
	setupErr  error
	typeErr   error
	keyErr    error
}

type keyCall struct {
	key       protocol.Key
	modifiers []protocol.Modifier
	repeat    int
}

func (m *mockExecutor) Name() string { return m.name }

func (m *mockExecutor) TypeText(ctx context.Context, text string) error {
	m.typeCalls = append(m.typeCalls, text)
	return m.typeErr
}

func (m *mockExecutor) SendKey(ctx context.Context, key protocol.Key, mods []protocol.Modifier, repeat int) error {
	m.keyCalls = append(m.keyCalls, keyCall{key, mods, repeat})
	return m.keyErr
}

func (m *mockExecutor) FocusTab(ctx context.Context, index int) error { return nil }
func (m *mockExecutor) Setup(ctx context.Context) error               { return m.setupErr }
func (m *mockExecutor) Cleanup(ctx context.Context) error             { return nil }

// test New()

func TestNew(t *testing.T) {
	exec := &mockExecutor{name: "mock"}
	c := New("ws://localhost:9999", "testclient", "linux", exec)

	if c.serverURL != "ws://localhost:9999" {
		t.Errorf("serverURL = %q, want %q", c.serverURL, "ws://localhost:9999")
	}
	if c.originalName != "testclient" {
		t.Errorf("originalName = %q, want %q", c.originalName, "testclient")
	}
	if c.platform != "linux" {
		t.Errorf("platform = %q, want %q", c.platform, "linux")
	}
	if c.executor != exec {
		t.Error("executor not set correctly")
	}
	if c.reconnectDelay != 2*time.Second {
		t.Errorf("reconnectDelay = %v, want %v", c.reconnectDelay, 2*time.Second)
	}
	if c.running {
		t.Error("running should be false initially")
	}
}

// test handleMessage with different message types

func TestHandleMessageTypeText(t *testing.T) {
	exec := &mockExecutor{name: "mock"}
	c := New("ws://localhost:9999", "test", "linux", exec)

	msg := protocol.NewTypeText("hello world")
	err := c.handleMessage(context.Background(), nil, msg)

	if err != nil {
		t.Fatalf("handleMessage error: %v", err)
	}
	if len(exec.typeCalls) != 1 {
		t.Fatalf("typeCalls = %d, want 1", len(exec.typeCalls))
	}
	if exec.typeCalls[0] != "hello world" {
		t.Errorf("typeCalls[0] = %q, want %q", exec.typeCalls[0], "hello world")
	}
}

func TestHandleMessageSendKey(t *testing.T) {
	exec := &mockExecutor{name: "mock"}
	c := New("ws://localhost:9999", "test", "linux", exec)

	msg := protocol.NewSendKey(protocol.KeyEnter, []protocol.Modifier{protocol.ModCtrl}, 2)
	err := c.handleMessage(context.Background(), nil, msg)

	if err != nil {
		t.Fatalf("handleMessage error: %v", err)
	}
	if len(exec.keyCalls) != 1 {
		t.Fatalf("keyCalls = %d, want 1", len(exec.keyCalls))
	}
	call := exec.keyCalls[0]
	if call.key != protocol.KeyEnter {
		t.Errorf("key = %q, want %q", call.key, protocol.KeyEnter)
	}
	if !reflect.DeepEqual(call.modifiers, []protocol.Modifier{protocol.ModCtrl}) {
		t.Errorf("modifiers = %v, want [ctrl]", call.modifiers)
	}
	if call.repeat != 2 {
		t.Errorf("repeat = %d, want 2", call.repeat)
	}
}

func TestHandleMessageTargetStatus(t *testing.T) {
	exec := &mockExecutor{name: "mock"}
	c := New("ws://localhost:9999", "test", "linux", exec)

	// active status
	msg := protocol.NewTargetStatus(true)
	err := c.handleMessage(context.Background(), nil, msg)
	if err != nil {
		t.Fatalf("handleMessage error: %v", err)
	}

	// inactive status
	msg = protocol.NewTargetStatus(false)
	err = c.handleMessage(context.Background(), nil, msg)
	if err != nil {
		t.Fatalf("handleMessage error: %v", err)
	}
}

func TestHandleMessageRegistered(t *testing.T) {
	exec := &mockExecutor{name: "mock"}
	c := New("ws://localhost:9999", "test", "linux", exec)

	// same name (no rename)
	msg := protocol.NewRegistered("test")
	err := c.handleMessage(context.Background(), nil, msg)
	if err != nil {
		t.Fatalf("handleMessage error: %v", err)
	}
	if c.assignedName != "test" {
		t.Errorf("assignedName = %q, want %q", c.assignedName, "test")
	}

	// different name (renamed by server)
	msg = protocol.NewRegistered("test2")
	err = c.handleMessage(context.Background(), nil, msg)
	if err != nil {
		t.Fatalf("handleMessage error: %v", err)
	}
	if c.assignedName != "test2" {
		t.Errorf("assignedName = %q, want %q", c.assignedName, "test2")
	}
}

// test Stop()

func TestStop(t *testing.T) {
	exec := &mockExecutor{name: "mock"}
	c := New("ws://localhost:9999", "test", "linux", exec)

	// set running to true first
	c.mu.Lock()
	c.running = true
	c.mu.Unlock()

	c.Stop()

	if c.isRunning() {
		t.Error("running should be false after Stop()")
	}
}

func TestSendPromptNotConnected(t *testing.T) {
	exec := &mockExecutor{name: "mock"}
	c := New("ws://localhost:9999", "test", "linux", exec)

	err := c.SendPrompt(context.Background(), "test prompt")
	if err == nil {
		t.Error("expected error when not connected")
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Errorf("error = %q, want 'not connected'", err.Error())
	}
}

// test with mock WebSocket server

var upgrader = websocket.Upgrader{}

func TestClientConnectsAndRegisters(t *testing.T) {
	// channel to receive register message
	registerCh := make(chan *protocol.Register, 1)

	// mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// read register message
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Logf("read error: %v", err)
			return
		}

		msg, err := protocol.ParseClientMessage(data)
		if err != nil {
			t.Logf("parse error: %v", err)
			return
		}

		if reg, ok := msg.(*protocol.Register); ok {
			registerCh <- reg
		}

		// send registered response
		resp, _ := protocol.Serialize(protocol.NewRegistered("testclient"))
		conn.WriteMessage(websocket.TextMessage, resp)

		// keep connection open briefly
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	// convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	exec := &mockExecutor{name: "mock"}
	c := New(wsURL, "testclient", "linux", exec)

	// run client in background
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go func() {
		c.Run(ctx)
	}()

	// wait for register message
	select {
	case reg := <-registerCh:
		if reg.Name != "testclient" {
			t.Errorf("register name = %q, want %q", reg.Name, "testclient")
		}
		if reg.Platform != "linux" {
			t.Errorf("register platform = %q, want %q", reg.Platform, "linux")
		}
	case <-time.After(300 * time.Millisecond):
		t.Error("timeout waiting for register message")
	}
}

func TestClientHandlesPing(t *testing.T) {
	// channel to confirm pong received
	pongCh := make(chan bool, 1)

	// mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// read register
		conn.ReadMessage()

		// send registered
		resp, _ := protocol.Serialize(protocol.NewRegistered("test"))
		conn.WriteMessage(websocket.TextMessage, resp)

		// send ping
		ping, _ := protocol.Serialize(protocol.NewPing())
		conn.WriteMessage(websocket.TextMessage, ping)

		// wait for pong
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		msg, err := protocol.ParseClientMessage(data)
		if err != nil {
			return
		}

		if _, ok := msg.(*protocol.Pong); ok {
			pongCh <- true
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	exec := &mockExecutor{name: "mock"}
	c := New(wsURL, "test", "linux", exec)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go func() {
		c.Run(ctx)
	}()

	select {
	case <-pongCh:
		// success
	case <-time.After(300 * time.Millisecond):
		t.Error("timeout waiting for pong response")
	}
}
