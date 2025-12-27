package server

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// test ConvertNumberWords

func TestConvertNumberWords(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"focused two", "focused2"},
		{"desktop three", "desktop3"},
		{"desktop", "desktop"},
		{"my desktop two", "mydesktop2"},
		{"Focused Two", "focused2"},
		{"test two", "test2"},
		{"test three", "test3"},
		{"test four", "test4"},
		{"test five", "test5"},
		{"test six", "test6"},
		{"test seven", "test7"},
		{"test eight", "test8"},
		{"test nine", "test9"},
		{"test ten", "test10"},
		// homophones
		{"focused too", "focused2"},
		{"focused to", "focused2"},
		{"test won", "test1"},
		{"test for", "test4"},
		{"test fore", "test4"},
		{"test ate", "test8"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ConvertNumberWords(tt.input)
			if got != tt.want {
				t.Errorf("ConvertNumberWords(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"focused-two", "focusedtwo"},  // hyphen removed, "two" is now part of word
		{"focused two", "focused2"},    // space separated, "two" converted to 2
		{"My-Desktop", "mydesktop"},
		{"test-client-3", "testclient3"},
		{"UPPERCASE", "uppercase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Normalize(tt.input)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// test ClientManager

func TestNewClientManager(t *testing.T) {
	aliases := map[string]string{"desk": "desktop"}
	m := NewClientManager(aliases)

	if m == nil {
		t.Fatal("NewClientManager returned nil")
	}
	if m.clients == nil {
		t.Error("clients map is nil")
	}
	if m.aliases == nil {
		t.Error("aliases map is nil")
	}
	if m.aliases["desk"] != "desktop" {
		t.Errorf("aliases[desk] = %q, want 'desktop'", m.aliases["desk"])
	}
}

func TestNewClientManagerNilAliases(t *testing.T) {
	m := NewClientManager(nil)

	if m.aliases == nil {
		t.Error("aliases should be initialized even when nil passed")
	}
}

func TestGetUniqueName(t *testing.T) {
	m := NewClientManager(nil)

	// add a mock registered client
	mockConn := &websocket.Conn{}
	m.clients[mockConn] = &ConnectedClient{
		Conn:       mockConn,
		Name:       "test",
		Registered: true,
	}

	tests := []struct {
		requested string
		want      string
	}{
		{"foo", "foo"},     // unique name
		{"test", "test2"},  // duplicate of existing "test"
		{"test2", "test2"}, // "test2" doesn't exist yet, so it's unique
	}

	for _, tt := range tests {
		t.Run(tt.requested, func(t *testing.T) {
			got := m.getUniqueName(tt.requested)
			if got != tt.want {
				t.Errorf("getUniqueName(%q) = %q, want %q", tt.requested, got, tt.want)
			}
		})
	}
}

func TestGetUniqueNameWithExistingNumbers(t *testing.T) {
	m := NewClientManager(nil)

	// add test, test2, test3
	for i, name := range []string{"test", "test2", "test3"} {
		conn := &websocket.Conn{}
		m.clients[conn] = &ConnectedClient{
			Conn:       conn,
			Name:       name,
			Registered: true,
		}
		_ = i
	}

	got := m.getUniqueName("test")
	if got != "test4" {
		t.Errorf("getUniqueName('test') = %q, want 'test4'", got)
	}
}

func TestAddAndRemove(t *testing.T) {
	m := NewClientManager(nil)

	conn := &websocket.Conn{}
	client := m.Add(conn)

	if client == nil {
		t.Fatal("Add returned nil")
	}
	if client.Conn != conn {
		t.Error("client.Conn not set correctly")
	}
	if len(m.clients) != 1 {
		t.Errorf("clients count = %d, want 1", len(m.clients))
	}

	m.Remove(conn)
	if len(m.clients) != 0 {
		t.Errorf("clients count after remove = %d, want 0", len(m.clients))
	}
}

func TestListClients(t *testing.T) {
	m := NewClientManager(nil)

	// empty
	if len(m.ListClients()) != 0 {
		t.Error("ListClients should be empty initially")
	}

	// add unregistered client
	conn1 := &websocket.Conn{}
	m.clients[conn1] = &ConnectedClient{Conn: conn1, Registered: false}

	if len(m.ListClients()) != 0 {
		t.Error("ListClients should not include unregistered clients")
	}

	// add registered client
	conn2 := &websocket.Conn{}
	m.clients[conn2] = &ConnectedClient{Conn: conn2, Name: "test", Registered: true}

	names := m.ListClients()
	if len(names) != 1 || names[0] != "test" {
		t.Errorf("ListClients() = %v, want ['test']", names)
	}
}

func TestIsValidTarget(t *testing.T) {
	m := NewClientManager(map[string]string{"desk": "desktop"})

	// add registered client
	conn := &websocket.Conn{}
	m.clients[conn] = &ConnectedClient{
		Conn:       conn,
		Name:       "focused2",
		Registered: true,
	}

	tests := []struct {
		name string
		want bool
	}{
		{"focused2", true},      // exact match
		{"focused two", true},   // number word conversion
		{"focused", true},       // substring
		{"focus", true},         // prefix
		{"desk", true},          // alias
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.IsValidTarget(tt.name)
			if got != tt.want {
				t.Errorf("IsValidTarget(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestSetTargetNoClients(t *testing.T) {
	// SetTarget with no clients should return false
	m := NewClientManager(nil)

	if m.SetTarget("anything") {
		t.Error("SetTarget with no clients should return false")
	}
}

func TestSetTargetNonexistent(t *testing.T) {
	// SetTarget with nonexistent target should return false
	m := NewClientManager(nil)

	conn := &websocket.Conn{}
	m.clients[conn] = &ConnectedClient{Conn: conn, Name: "test", Registered: true}

	if m.SetTarget("nonexistent") {
		t.Error("SetTarget('nonexistent') should return false")
	}
}

func TestTextHistory(t *testing.T) {
	m := NewClientManager(nil)

	// add and set active target
	conn := &websocket.Conn{}
	m.clients[conn] = &ConnectedClient{
		Conn:       conn,
		Name:       "test",
		Registered: true,
	}
	m.activeTarget = "test"

	// record some text
	m.recordTextSentLocked(10)
	m.recordTextSentLocked(20)
	m.recordTextSentLocked(30)

	// pop last
	if got := m.PopLastTextLength(); got != 30 {
		t.Errorf("PopLastTextLength() = %d, want 30", got)
	}
	if got := m.PopLastTextLength(); got != 20 {
		t.Errorf("PopLastTextLength() = %d, want 20", got)
	}
	if got := m.PopLastTextLength(); got != 10 {
		t.Errorf("PopLastTextLength() = %d, want 10", got)
	}
	if got := m.PopLastTextLength(); got != 0 {
		t.Errorf("PopLastTextLength() on empty = %d, want 0", got)
	}
}

func TestTextHistoryMaxSize(t *testing.T) {
	m := NewClientManager(nil)

	conn := &websocket.Conn{}
	m.clients[conn] = &ConnectedClient{
		Conn:       conn,
		Name:       "test",
		Registered: true,
	}
	m.activeTarget = "test"

	// add more than MaxTextHistory items
	for i := 1; i <= MaxTextHistory+5; i++ {
		m.recordTextSentLocked(i)
	}

	// should only have last MaxTextHistory items
	client := m.clients[conn]
	if len(client.TextHistory) != MaxTextHistory {
		t.Errorf("TextHistory length = %d, want %d", len(client.TextHistory), MaxTextHistory)
	}

	// first item should be 6 (items 1-5 were pushed out)
	if client.TextHistory[0] != 6 {
		t.Errorf("TextHistory[0] = %d, want 6", client.TextHistory[0])
	}
}

func TestUpdatePong(t *testing.T) {
	m := NewClientManager(nil)

	conn := &websocket.Conn{}
	client := &ConnectedClient{
		Conn:     conn,
		LastPong: time.Now().Add(-time.Minute),
	}
	m.clients[conn] = client

	before := client.LastPong
	m.UpdatePong(conn)
	after := client.LastPong

	if !after.After(before) {
		t.Error("UpdatePong should update LastPong time")
	}
}

func TestActiveTarget(t *testing.T) {
	m := NewClientManager(nil)

	if m.ActiveTarget() != "" {
		t.Error("ActiveTarget should be empty initially")
	}

	m.activeTarget = "test"
	if m.ActiveTarget() != "test" {
		t.Errorf("ActiveTarget() = %q, want 'test'", m.ActiveTarget())
	}
}

func TestResolveAlias(t *testing.T) {
	m := NewClientManager(map[string]string{
		"desk":   "desktop",
		"laptop": "macbook",
	})

	tests := []struct {
		input string
		want  string
	}{
		{"desk", "desktop"},
		{"DESK", "desktop"},  // case insensitive
		{"laptop", "macbook"},
		{"other", "other"},   // not an alias
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := m.resolveAlias(tt.input)
			if got != tt.want {
				t.Errorf("resolveAlias(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
