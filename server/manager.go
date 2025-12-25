package server

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PrajnaAvidya/yapatype/config"
	"github.com/PrajnaAvidya/yapatype/protocol"
	"github.com/gorilla/websocket"
)

const MaxTextHistory = 10

// number word to digit conversion for fuzzy matching
var numberWords = map[string]string{
	"one": "1", "won": "1",
	"two": "2", "too": "2", "to": "2",
	"three": "3",
	"four": "4", "for": "4", "fore": "4",
	"five": "5",
	"six": "6",
	"seven": "7",
	"eight": "8", "ate": "8",
	"nine": "9",
	"ten": "10",
}

// ConvertNumberWords converts number words to digits: "focused two" -> "focused2"
func ConvertNumberWords(text string) string {
	words := strings.Fields(strings.ToLower(text))
	result := make([]string, 0, len(words))
	for _, word := range words {
		if digit, ok := numberWords[word]; ok {
			result = append(result, digit)
		} else {
			result = append(result, word)
		}
	}
	return strings.Join(result, "")
}

// Normalize converts text to normalized form for matching
func Normalize(s string) string {
	return strings.ReplaceAll(ConvertNumberWords(s), "-", "")
}

// ConnectedClient represents a connected client
type ConnectedClient struct {
	Conn        *websocket.Conn
	Name        string
	Platform    string
	Registered  bool
	TextHistory []int     // char counts of sent text for scratch
	LastPong    time.Time // for connection health checks
}

// ClientManager manages connected clients and target selection
type ClientManager struct {
	clients      map[*websocket.Conn]*ConnectedClient
	activeTarget string
	aliases      map[string]string        // alias -> actual client name
	lastTypeText *protocol.TypeText       // for repeat command
	mu           sync.RWMutex

	// callbacks (set by server)
	OnPromptReceived   func(prompt string)
	OnClientRegistered func(name string)
	OnClientRemoved    func(name string)
}

// NewClientManager creates a new client manager
func NewClientManager(aliases map[string]string) *ClientManager {
	if aliases == nil {
		aliases = make(map[string]string)
	}
	return &ClientManager{
		clients: make(map[*websocket.Conn]*ConnectedClient),
		aliases: aliases,
	}
}

// saveTarget saves active target to state file
func (m *ClientManager) saveTarget() {
	path := config.StatePath()
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)

	if m.activeTarget != "" {
		os.WriteFile(path, []byte(m.activeTarget), 0644)
	} else if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}
}

// loadSavedTarget loads saved target from state file
func (m *ClientManager) loadSavedTarget() string {
	path := config.StatePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// getUniqueName returns a unique name, appending number if name exists
func (m *ClientManager) getUniqueName(requested string) string {
	existing := make(map[string]bool)
	for _, c := range m.clients {
		if c.Registered {
			existing[c.Name] = true
		}
	}

	if !existing[requested] {
		return requested
	}

	// strip trailing digits to get base name ("focused2" -> "focused")
	re := regexp.MustCompile(`^(.+?)(\d+)$`)
	matches := re.FindStringSubmatch(requested)
	baseName := requested
	if len(matches) == 3 {
		baseName = matches[1]
	}

	// find next available number
	n := 2
	for existing[fmt.Sprintf("%s%d", baseName, n)] {
		n++
	}
	return fmt.Sprintf("%s%d", baseName, n)
}

// Add adds a new connection
func (m *ClientManager) Add(conn *websocket.Conn) *ConnectedClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	client := &ConnectedClient{
		Conn:     conn,
		LastPong: time.Now(),
	}
	m.clients[conn] = client
	return client
}

// Remove removes a connection
func (m *ClientManager) Remove(conn *websocket.Conn) {
	m.mu.Lock()
	client := m.clients[conn]
	delete(m.clients, conn)

	var wasActive bool
	if client != nil && client.Registered && client.Name != "" {
		if client.Name == m.activeTarget {
			wasActive = true
		}
	}
	m.mu.Unlock()

	if client != nil && client.Registered && client.Name != "" {
		if m.OnClientRemoved != nil {
			m.OnClientRemoved(client.Name)
		}
		if wasActive {
			m.autoSelectTarget()
		}
	}
}

// autoSelectTarget selects a target from connected clients
func (m *ClientManager) autoSelectTarget() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, client := range m.clients {
		if client.Registered {
			m.activeTarget = client.Name
			fmt.Printf("auto-selected target: %s\n", client.Name)
			return
		}
	}
	m.activeTarget = ""
}

// Register registers a client with name and platform, returns assigned name
func (m *ClientManager) Register(conn *websocket.Conn, name, platform string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[conn]
	if !ok {
		return "", fmt.Errorf("connection not found")
	}

	// probe existing clients with same name - check for dead connections
	for existingConn, existingClient := range m.clients {
		if existingClient.Registered && existingClient.Name == name {
			// try to ping - if connection is dead, this will fail
			pingData, _ := protocol.Serialize(protocol.NewPing())
			if err := existingConn.WriteMessage(websocket.TextMessage, pingData); err != nil {
				// dead connection - remove it
				fmt.Printf("removing dead connection: %s\n", name)
				delete(m.clients, existingConn)
			}
		}
	}

	// get unique name
	assignedName := m.getUniqueName(name)

	client.Name = assignedName
	client.Platform = platform
	client.Registered = true
	fmt.Printf("client registered: %s (%s)\n", assignedName, platform)

	// auto-select target: prefer saved target if matches, else first client
	if m.activeTarget == "" {
		savedTarget := m.loadSavedTarget()
		if savedTarget != "" && savedTarget == assignedName {
			// saved target just connected - restore it
			m.activeTarget = assignedName
			fmt.Printf("restored target: %s\n", assignedName)
		} else if savedTarget == "" {
			// no saved target - select first client
			m.activeTarget = assignedName
			fmt.Printf("active target: %s\n", assignedName)
		}
		// else: saved target exists but hasn't connected yet - wait for it
	}

	// callback outside lock would be better but keeping simple for now
	if m.OnClientRegistered != nil {
		m.OnClientRegistered(assignedName)
	}

	return assignedName, nil
}

// resolveAlias resolves alias to actual client name
func (m *ClientManager) resolveAlias(name string) string {
	nameLower := strings.ToLower(name)
	for alias, target := range m.aliases {
		if strings.ToLower(alias) == nameLower {
			return target
		}
	}
	return name
}

// SetTarget sets active target by name (fuzzy matching), returns true if found
func (m *ClientManager) SetTarget(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// collect registered clients
	var registered []*ConnectedClient
	for _, c := range m.clients {
		if c.Registered {
			registered = append(registered, c)
		}
	}
	if len(registered) == 0 {
		return false
	}

	// resolve alias
	resolvedName := m.resolveAlias(name)
	query := Normalize(resolvedName)

	var newTarget string
	var matchType string

	// exact match first
	for _, client := range registered {
		if client.Name == resolvedName {
			newTarget = client.Name
			break
		}
	}

	// normalized exact match
	if newTarget == "" {
		for _, client := range registered {
			if Normalize(client.Name) == query {
				newTarget = client.Name
				break
			}
		}
	}

	// query is substring of client name
	if newTarget == "" {
		for _, client := range registered {
			if strings.Contains(Normalize(client.Name), query) {
				newTarget = client.Name
				matchType = " (fuzzy)"
				break
			}
		}
	}

	// client name starts with query
	if newTarget == "" {
		for _, client := range registered {
			if strings.HasPrefix(Normalize(client.Name), query) {
				newTarget = client.Name
				matchType = " (prefix)"
				break
			}
		}
	}

	if newTarget == "" {
		return false
	}

	// send notifications if target actually changed
	oldTarget := m.activeTarget
	if oldTarget != newTarget {
		if oldTarget != "" {
			m.sendToClientLocked(oldTarget, protocol.NewTargetStatus(false))
		}
		m.sendToClientLocked(newTarget, protocol.NewTargetStatus(true))
	}

	m.activeTarget = newTarget
	m.saveTarget()
	fmt.Printf("target switched to: %s%s\n", newTarget, matchType)
	return true
}

// IsValidTarget checks if name matches a client or alias
func (m *ClientManager) IsValidTarget(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// check aliases
	nameLower := strings.ToLower(name)
	for alias := range m.aliases {
		if strings.ToLower(alias) == nameLower {
			return true
		}
	}

	// check registered clients
	query := Normalize(name)
	for _, client := range m.clients {
		if !client.Registered {
			continue
		}
		if client.Name == name {
			return true
		}
		normalized := Normalize(client.Name)
		if normalized == query {
			return true
		}
		if strings.Contains(normalized, query) {
			return true
		}
		if strings.HasPrefix(normalized, query) {
			return true
		}
	}

	return false
}

// getTargetConn returns the connection for active target (must hold lock)
func (m *ClientManager) getTargetConnLocked() *websocket.Conn {
	if m.activeTarget == "" {
		return nil
	}
	for _, client := range m.clients {
		if client.Name == m.activeTarget {
			return client.Conn
		}
	}
	return nil
}

// getClientConn returns the connection for a client by name (must hold lock)
func (m *ClientManager) getClientConnLocked(name string) *websocket.Conn {
	for _, client := range m.clients {
		if client.Name == name {
			return client.Conn
		}
	}
	return nil
}

// sendToClientLocked sends message to a specific client (must hold lock)
func (m *ClientManager) sendToClientLocked(name string, msg any) error {
	conn := m.getClientConnLocked(name)
	if conn == nil {
		return fmt.Errorf("client not found: %s", name)
	}
	data, err := protocol.Serialize(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

// SendToTarget sends message to active target
func (m *ClientManager) SendToTarget(msg any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn := m.getTargetConnLocked()
	if conn == nil {
		return fmt.Errorf("no active target")
	}

	data, err := protocol.Serialize(msg)
	if err != nil {
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}

	// track TypeText for scratch and repeat
	if tt, ok := msg.(*protocol.TypeText); ok {
		m.lastTypeText = tt
		m.recordTextSentLocked(len(tt.Text))
	}

	return nil
}

// recordTextSentLocked records text sent to active target (must hold lock)
func (m *ClientManager) recordTextSentLocked(charCount int) {
	conn := m.getTargetConnLocked()
	if conn == nil {
		return
	}
	client := m.clients[conn]
	if client == nil {
		return
	}
	client.TextHistory = append(client.TextHistory, charCount)
	if len(client.TextHistory) > MaxTextHistory {
		client.TextHistory = client.TextHistory[1:]
	}
}

// PopLastTextLength returns and removes the last text length sent to active target
func (m *ClientManager) PopLastTextLength() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn := m.getTargetConnLocked()
	if conn == nil {
		return 0
	}
	client := m.clients[conn]
	if client == nil || len(client.TextHistory) == 0 {
		return 0
	}
	length := client.TextHistory[len(client.TextHistory)-1]
	client.TextHistory = client.TextHistory[:len(client.TextHistory)-1]
	return length
}

// LastTypeText returns the last TypeText sent
func (m *ClientManager) LastTypeText() *protocol.TypeText {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastTypeText
}

// Broadcast sends message to all connected clients
func (m *ClientManager) Broadcast(msg any) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := protocol.Serialize(msg)
	if err != nil {
		return
	}

	for _, client := range m.clients {
		if client.Registered {
			client.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// ListClients returns registered client names
func (m *ClientManager) ListClients() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var names []string
	for _, client := range m.clients {
		if client.Registered {
			names = append(names, client.Name)
		}
	}
	return names
}

// PingAll sends ping to all connected clients
func (m *ClientManager) PingAll() {
	m.Broadcast(protocol.NewPing())
}

// RemoveStale removes clients that haven't responded within timeout
func (m *ClientManager) RemoveStale(timeout time.Duration) []string {
	m.mu.Lock()
	now := time.Now()
	var stale []*websocket.Conn
	var staleNames []string

	for conn, client := range m.clients {
		if client.Registered && now.Sub(client.LastPong) > timeout {
			stale = append(stale, conn)
			staleNames = append(staleNames, client.Name)
		}
	}
	m.mu.Unlock()

	// remove outside lock to avoid deadlock with callbacks
	for i, conn := range stale {
		fmt.Printf("removing stale client: %s\n", staleNames[i])
		m.Remove(conn)
	}

	return staleNames
}

// UpdatePong updates last pong time for a client
func (m *ClientManager) UpdatePong(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.clients[conn]; ok {
		client.LastPong = time.Now()
	}
}

// ActiveTarget returns the current active target name
func (m *ClientManager) ActiveTarget() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeTarget
}
