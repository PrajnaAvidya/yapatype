package client

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/PrajnaAvidya/yapatype/executor"
	"github.com/PrajnaAvidya/yapatype/protocol"
	"github.com/gorilla/websocket"
)

// Client connects to the yapatype server and executes commands
type Client struct {
	serverURL      string
	originalName   string // for reconnection (never changes)
	assignedName   string // from server Registered msg
	platform       string
	executor       executor.Executor
	reconnectDelay time.Duration
	conn           *websocket.Conn
	running        bool
	mu             sync.Mutex
}

// New creates a new yapatype client
func New(serverURL, name, platform string, exec executor.Executor) *Client {
	return &Client{
		serverURL:      serverURL,
		originalName:   name,
		platform:       platform,
		executor:       exec,
		reconnectDelay: 2 * time.Second,
	}
}

// Run starts the client connection loop
// blocks until Stop() is called or context is cancelled
func (c *Client) Run(ctx context.Context) error {
	if err := c.executor.Setup(ctx); err != nil {
		return fmt.Errorf("executor setup: %w", err)
	}
	defer c.executor.Cleanup(ctx)

	c.mu.Lock()
	c.running = true
	c.mu.Unlock()

	for c.isRunning() {
		if err := c.connect(ctx); err != nil {
			if ctx.Err() != nil {
				// context cancelled, exit cleanly
				return nil
			}
			fmt.Printf("connection error: %v\n", err)
		}

		if c.isRunning() {
			fmt.Printf("reconnecting in %v...\n", c.reconnectDelay)
			select {
			case <-time.After(c.reconnectDelay):
			case <-ctx.Done():
				return nil
			}
		}
	}

	return nil
}

// Stop signals the client to stop and close the connection
func (c *Client) Stop() {
	c.mu.Lock()
	c.running = false
	conn := c.conn
	c.mu.Unlock()

	if conn != nil {
		conn.Close()
	}
}

// SendPrompt sends a whisper prompt to the server
func (c *Client) SendPrompt(ctx context.Context, prompt string) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := protocol.Serialize(protocol.NewPrompt(prompt))
	if err != nil {
		return fmt.Errorf("serialize prompt: %w", err)
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// isRunning returns whether the client should keep running
func (c *Client) isRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

// connect establishes a connection and handles messages until disconnect
func (c *Client) connect(ctx context.Context) error {
	u, err := url.Parse(c.serverURL)
	if err != nil {
		return fmt.Errorf("parse server url: %w", err)
	}

	fmt.Printf("connecting to %s...\n", c.serverURL)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		conn.Close()
	}()

	fmt.Println("connected!")

	// register with server
	registerMsg := protocol.NewRegister(c.originalName, c.platform)
	data, err := protocol.Serialize(registerMsg)
	if err != nil {
		return fmt.Errorf("serialize register: %w", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("send register: %w", err)
	}

	// handle messages
	return c.handleMessages(ctx, conn)
}

// handleMessages reads and processes messages from the server
func (c *Client) handleMessages(ctx context.Context, conn *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println("connection closed")
				return nil
			}
			return fmt.Errorf("read message: %w", err)
		}

		msg, err := protocol.ParseServerMessage(data)
		if err != nil {
			fmt.Printf("error parsing message: %v\n", err)
			continue
		}

		if err := c.handleMessage(ctx, conn, msg); err != nil {
			fmt.Printf("error handling message: %v\n", err)
		}
	}
}

// handleMessage processes a single server message
func (c *Client) handleMessage(ctx context.Context, conn *websocket.Conn, msg any) error {
	switch m := msg.(type) {
	case *protocol.TypeText:
		return c.executor.TypeText(ctx, m.Text)

	case *protocol.SendKey:
		return c.executor.SendKey(ctx, m.Key, m.Modifiers, m.Repeat)

	case *protocol.Ping:
		data, err := protocol.Serialize(protocol.NewPong())
		if err != nil {
			return fmt.Errorf("serialize pong: %w", err)
		}
		return conn.WriteMessage(websocket.TextMessage, data)

	case *protocol.TargetStatus:
		if m.IsActive {
			fmt.Println("[active]")
		} else {
			fmt.Println("[inactive]")
		}

	case *protocol.Registered:
		c.mu.Lock()
		if m.Name != c.originalName {
			fmt.Printf("registered as: %s\n", m.Name)
		}
		c.assignedName = m.Name
		c.mu.Unlock()
	}

	return nil
}
