package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
)

// Handler executes a command on behalf of an IPC client. Implementations
// must be safe to call from multiple goroutines; the bridge's command
// channel provides that by serializing execution onto its own goroutine.
type Handler interface {
	// Command runs an argv-style weir command and returns its output.
	Command(args []string) (string, error)
}

// Server is the weir control socket server.
type Server struct {
	path     string
	listener net.Listener
	handler  Handler
	log      *slog.Logger

	mu          sync.Mutex
	subscribers map[*subscriber]struct{}
	closed      bool
}

// subscriber is a connection in event-streaming mode. Each subscriber has a
// single-slot mailbox holding the latest pending event line: a slow reader
// skips intermediate states rather than lagging further and further behind.
type subscriber struct {
	ch chan []byte
}

// Listen creates the control socket at path. A stale socket left behind by
// a crashed weir is detected (nothing accepts connections on it) and
// replaced; a live socket is an error so two weirs never fight over one
// session.
func Listen(path string, handler Handler, logger *slog.Logger) (*Server, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if _, err := os.Stat(path); err == nil {
		// Something is already there. If a weir is listening, refuse to
		// start; if not, it is a stale socket from a crashed instance.
		if c, err := net.Dial("unix", path); err == nil {
			c.Close()
			return nil, fmt.Errorf("ipc: %s is already in use by a running weir", path)
		}
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("ipc: removing stale socket: %w", err)
		}
		logger.Warn("removed stale control socket", "path", path)
	}
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("ipc: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		l.Close()
		return nil, fmt.Errorf("ipc: chmod socket: %w", err)
	}
	s := &Server{
		path:        path,
		listener:    l,
		handler:     handler,
		log:         logger,
		subscribers: make(map[*subscriber]struct{}),
	}
	go s.acceptLoop()
	logger.Info("control socket listening", "path", path)
	return s, nil
}

// Path returns the socket path the server is listening on.
func (s *Server) Path() string { return s.path }

// Close stops the server and removes the socket.
func (s *Server) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	err := s.listener.Close()
	os.Remove(s.path)
	return err
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if !closed {
				s.log.Error("ipc accept", "err", err)
			}
			return
		}
		go s.serve(conn)
	}
}

func (s *Server) serve(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	enc := json.NewEncoder(conn)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) == 0 {
			if err != nil && !errors.Is(err, io.EOF) {
				s.log.Debug("ipc read", "err", err)
			}
			return
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			enc.Encode(Response{Success: false, Error: "invalid request: " + err.Error()})
			return
		}
		switch {
		case req.Subscribe:
			s.streamEvents(conn)
			return
		case len(req.Command) > 0:
			out, err := s.handler.Command(req.Command)
			resp := Response{Success: err == nil, Output: out}
			if err != nil {
				resp.Error = err.Error()
			}
			if err := enc.Encode(resp); err != nil {
				return
			}
		default:
			enc.Encode(Response{Success: false, Error: "request must contain a command or subscribe"})
			return
		}
		if err != nil {
			// The read that produced this line also hit EOF.
			return
		}
	}
}

// streamEvents registers the connection as a subscriber and writes event
// lines until the client disconnects.
func (s *Server) streamEvents(conn net.Conn) {
	sub := &subscriber{ch: make(chan []byte, 1)}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.subscribers[sub] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.subscribers, sub)
		s.mu.Unlock()
	}()

	// Detect client disconnect: subscribers send nothing further, so any
	// read completing (EOF or data) means the connection is done. Without
	// this, a subscriber that connects and never receives an event would
	// leak its goroutine and fd forever after disconnecting.
	gone := make(chan struct{})
	go func() {
		io.Copy(io.Discard, conn)
		close(gone)
	}()

	for {
		select {
		case line := <-sub.ch:
			if _, err := conn.Write(line); err != nil {
				return
			}
		case <-gone:
			return
		}
	}
}

// Broadcast sends a pre-marshalled event line to every subscriber. The
// line must already end in a newline. Slow subscribers skip intermediate
// events: only the most recent unread line is retained per subscriber.
//
// Broadcast may be called from any goroutine.
func (s *Server) Broadcast(line []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sub := range s.subscribers {
		// Drain the stale event, if any, then deposit the new one. The
		// channel has capacity 1 and this is the only sender (the mutex is
		// held), so the send cannot block.
		select {
		case <-sub.ch:
		default:
		}
		sub.ch <- line
	}
}

// HasSubscribers reports whether any connection is currently subscribed.
// Callers can use this to skip marshalling events nobody will read.
func (s *Server) HasSubscribers() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subscribers) > 0
}
