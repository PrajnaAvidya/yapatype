package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PrajnaAvidya/yapatype/protocol"
	"github.com/gorilla/websocket"
)

func TestNewServer(t *testing.T) {
	srv := NewServer("localhost", 9999, map[string]string{"test": "alias"})

	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
	if srv.host != "localhost" {
		t.Errorf("host = %q, want 'localhost'", srv.host)
	}
	if srv.port != 9999 {
		t.Errorf("port = %d, want 9999", srv.port)
	}
	if srv.Manager == nil {
		t.Error("Manager should not be nil")
	}
}

// helper to create test server and connect websocket client
func setupTestServer(t *testing.T) (*Server, *httptest.Server, *websocket.Conn, func()) {
	srv := NewServer("", 0, nil)

	// create http test server
	ts := httptest.NewServer(http.HandlerFunc(srv.handleConnection))

	// connect websocket client
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}

	cleanup := func() {
		conn.Close()
		ts.Close()
	}

	return srv, ts, conn, cleanup
}

func TestServerAcceptsConnection(t *testing.T) {
	srv, _, conn, cleanup := setupTestServer(t)
	defer cleanup()

	// give server time to process connection
	time.Sleep(10 * time.Millisecond)

	// should have one client
	srv.Manager.mu.RLock()
	count := len(srv.Manager.clients)
	srv.Manager.mu.RUnlock()

	if count != 1 {
		t.Errorf("expected 1 client, got %d", count)
	}

	_ = conn // used in defer cleanup
}

func TestServerHandlesRegister(t *testing.T) {
	srv, _, conn, cleanup := setupTestServer(t)
	defer cleanup()

	// send register message
	regMsg := protocol.NewRegister("testclient", "linux")
	data, _ := protocol.Serialize(regMsg)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// read response
	_, respData, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	msg, err := protocol.ParseServerMessage(respData)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	reg, ok := msg.(*protocol.Registered)
	if !ok {
		t.Fatalf("expected Registered message, got %T", msg)
	}
	if reg.Name != "testclient" {
		t.Errorf("registered name = %q, want 'testclient'", reg.Name)
	}

	// check server state
	if srv.Manager.ActiveTarget() != "testclient" {
		t.Errorf("active target = %q, want 'testclient'", srv.Manager.ActiveTarget())
	}
}

func TestServerHandlesPong(t *testing.T) {
	srv, _, conn, cleanup := setupTestServer(t)
	defer cleanup()

	// register first
	regMsg := protocol.NewRegister("test", "linux")
	data, _ := protocol.Serialize(regMsg)
	conn.WriteMessage(websocket.TextMessage, data)
	conn.ReadMessage() // read Registered response

	// get initial pong time
	srv.Manager.mu.RLock()
	var client *ConnectedClient
	for _, c := range srv.Manager.clients {
		client = c
		break
	}
	initialPong := client.LastPong
	srv.Manager.mu.RUnlock()

	// wait a bit
	time.Sleep(10 * time.Millisecond)

	// send pong
	pongMsg := protocol.NewPong()
	data, _ = protocol.Serialize(pongMsg)
	conn.WriteMessage(websocket.TextMessage, data)

	// give server time to process
	time.Sleep(10 * time.Millisecond)

	// check pong time was updated
	srv.Manager.mu.RLock()
	newPong := client.LastPong
	srv.Manager.mu.RUnlock()

	if !newPong.After(initialPong) {
		t.Error("Pong should update LastPong time")
	}
}

func TestServerHandlesPrompt(t *testing.T) {
	srv, _, conn, cleanup := setupTestServer(t)
	defer cleanup()

	// set up callback
	promptReceived := make(chan string, 1)
	srv.Manager.OnPromptReceived = func(prompt string) {
		promptReceived <- prompt
	}

	// register first
	regMsg := protocol.NewRegister("test", "linux")
	data, _ := protocol.Serialize(regMsg)
	conn.WriteMessage(websocket.TextMessage, data)
	conn.ReadMessage() // read Registered response

	// send prompt
	promptMsg := protocol.NewPrompt("test prompt content")
	data, _ = protocol.Serialize(promptMsg)
	conn.WriteMessage(websocket.TextMessage, data)

	// wait for callback
	select {
	case prompt := <-promptReceived:
		if prompt != "test prompt content" {
			t.Errorf("prompt = %q, want 'test prompt content'", prompt)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for prompt callback")
	}
}

func TestServerUniqueNames(t *testing.T) {
	srv := NewServer("", 0, nil)
	ts := httptest.NewServer(http.HandlerFunc(srv.handleConnection))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	// connect first client
	conn1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer conn1.Close()

	regMsg := protocol.NewRegister("test", "linux")
	data, _ := protocol.Serialize(regMsg)
	conn1.WriteMessage(websocket.TextMessage, data)

	_, respData, _ := conn1.ReadMessage()
	msg1, _ := protocol.ParseServerMessage(respData)
	reg1 := msg1.(*protocol.Registered)

	if reg1.Name != "test" {
		t.Errorf("first client name = %q, want 'test'", reg1.Name)
	}

	// connect second client with same name
	conn2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer conn2.Close()

	conn2.WriteMessage(websocket.TextMessage, data) // same register message

	_, respData, _ = conn2.ReadMessage()
	msg2, _ := protocol.ParseServerMessage(respData)
	reg2 := msg2.(*protocol.Registered)

	if reg2.Name != "test2" {
		t.Errorf("second client name = %q, want 'test2'", reg2.Name)
	}
}

func TestServerStop(t *testing.T) {
	srv := NewServer("localhost", 19998, nil)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx)
	}()

	// wait for server to start
	time.Sleep(50 * time.Millisecond)

	// stop server
	cancel()
	srv.Stop()

	// should exit cleanly
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for server to stop")
	}
}
