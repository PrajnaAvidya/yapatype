package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/PrajnaAvidya/yapatype/protocol"
	"github.com/gorilla/websocket"
)

const (
	PingInterval = 30 * time.Second
	PongTimeout  = 40 * time.Second
)

// Server is the WebSocket server for client connections
type Server struct {
	host     string
	port     int
	Manager  *ClientManager
	upgrader websocket.Upgrader
	server   *http.Server
	running  bool
	mu       sync.Mutex
}

// NewServer creates a new WebSocket server
func NewServer(host string, port int, aliases map[string]string) *Server {
	return &Server{
		host:    host,
		port:    port,
		Manager: NewClientManager(aliases),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // allow all connections
			},
		},
	}
}

// Run starts the WebSocket server and blocks until stopped
func (s *Server) Run(ctx context.Context) error {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleConnection)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// start ping loop
	go s.pingLoop(ctx)

	fmt.Printf("websocket server listening on %s\n", addr)

	// run server
	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// wait for context cancellation or error
	select {
	case <-ctx.Done():
		s.Stop()
		return nil
	case err := <-errCh:
		return err
	}
}

// Stop gracefully stops the server
func (s *Server) Stop() {
	s.mu.Lock()
	s.running = false
	server := s.server
	s.mu.Unlock()

	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}
}

// handleConnection handles a new WebSocket connection
func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("upgrade error: %v\n", err)
		return
	}

	client := s.Manager.Add(conn)
	remote := r.RemoteAddr
	fmt.Printf("client connected: %s\n", remote)

	defer func() {
		name := client.Name
		if name == "" {
			name = remote
		}
		fmt.Printf("client disconnected: %s\n", name)
		s.Manager.Remove(conn)
		conn.Close()
	}()

	s.handleMessages(conn, client)
}

// handleMessages reads and processes messages from a client
func (s *Server) handleMessages(conn *websocket.Conn, client *ConnectedClient) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return
			}
			if websocket.IsUnexpectedCloseError(err) {
				return
			}
			fmt.Printf("read error: %v\n", err)
			return
		}

		msg, err := protocol.ParseClientMessage(data)
		if err != nil {
			fmt.Printf("parse error: %v\n", err)
			continue
		}

		if err := s.handleMessage(conn, client, msg); err != nil {
			fmt.Printf("handle error: %v\n", err)
		}
	}
}

// handleMessage processes a single client message
func (s *Server) handleMessage(conn *websocket.Conn, client *ConnectedClient, msg any) error {
	switch m := msg.(type) {
	case *protocol.Register:
		assignedName, err := s.Manager.Register(conn, m.Name, m.Platform)
		if err != nil {
			return err
		}
		// send Registered response
		resp := protocol.NewRegistered(assignedName)
		data, err := protocol.Serialize(resp)
		if err != nil {
			return err
		}
		return conn.WriteMessage(websocket.TextMessage, data)

	case *protocol.Prompt:
		fmt.Printf("received prompt from client (%d chars)\n", len(m.Prompt))
		if s.Manager.OnPromptReceived != nil {
			s.Manager.OnPromptReceived(m.Prompt)
		}

	case *protocol.Pong:
		s.Manager.UpdatePong(conn)
	}

	return nil
}

// pingLoop periodically pings clients and removes stale ones
func (s *Server) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()

			if !running {
				return
			}

			s.Manager.PingAll()
			s.Manager.RemoveStale(PongTimeout)
		}
	}
}

// SendToTarget sends a message to the active target client
func (s *Server) SendToTarget(msg any) error {
	return s.Manager.SendToTarget(msg)
}
