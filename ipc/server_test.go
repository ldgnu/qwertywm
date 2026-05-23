package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// echoHandler records commands and returns canned responses.
type echoHandler struct {
	mu    sync.Mutex
	calls [][]string
}

func (h *echoHandler) Command(args []string) (string, error) {
	h.mu.Lock()
	h.calls = append(h.calls, args)
	h.mu.Unlock()
	if args[0] == "fail" {
		return "", errors.New("it failed")
	}
	return strings.Join(args, " "), nil
}

func newServer(t *testing.T) (*Server, *echoHandler, string) {
	t.Helper()
	// Keep the socket path short: unix socket paths are limited to ~108
	// bytes and t.TempDir can be long.
	path := filepath.Join(t.TempDir(), "w.sock")
	h := &echoHandler{}
	s, err := Listen(path, h, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s, h, path
}

func send(t *testing.T, conn net.Conn, req Request) Response {
	t.Helper()
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		t.Fatal(err)
	}
	var resp Response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	return resp
}

func TestCommandRoundTrip(t *testing.T) {
	_, h, path := newServer(t)
	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	resp := send(t, conn, Request{Command: []string{"focus", "next"}})
	if !resp.Success || resp.Output != "focus next" {
		t.Errorf("response = %+v", resp)
	}
	// Multiple commands on one connection.
	resp = send(t, conn, Request{Command: []string{"get", "state"}})
	if !resp.Success || resp.Output != "get state" {
		t.Errorf("second response = %+v", resp)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.calls) != 2 {
		t.Errorf("handler saw %d calls, want 2", len(h.calls))
	}
}

func TestCommandError(t *testing.T) {
	_, _, path := newServer(t)
	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	resp := send(t, conn, Request{Command: []string{"fail"}})
	if resp.Success || resp.Error != "it failed" {
		t.Errorf("response = %+v", resp)
	}
}

func TestInvalidRequest(t *testing.T) {
	_, _, path := newServer(t)
	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	fmt.Fprintln(conn, "this is not json")
	var resp Response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Success {
		t.Errorf("invalid request succeeded: %+v", resp)
	}
}

func TestSubscribe(t *testing.T) {
	s, _, path := newServer(t)
	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(Request{Subscribe: true}); err != nil {
		t.Fatal(err)
	}
	// Wait for the subscription to register before broadcasting.
	deadline := time.Now().Add(5 * time.Second)
	for !s.HasSubscribers() {
		if time.Now().After(deadline) {
			t.Fatal("subscriber never registered")
		}
		time.Sleep(time.Millisecond)
	}

	s.Broadcast([]byte(`{"event":"state","state":{"n":1}}` + "\n"))
	r := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := r.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	var ev Event
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		t.Fatalf("event line %q: %v", line, err)
	}
	if ev.Event != "state" {
		t.Errorf("event = %+v", ev)
	}

	// A second broadcast arrives as a second line.
	s.Broadcast([]byte(`{"event":"state","state":{"n":2}}` + "\n"))
	line, err = r.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(line, `"n":2`) {
		t.Errorf("second event = %q", line)
	}
}

func TestSubscriberDisconnectUnregisters(t *testing.T) {
	s, _, path := newServer(t)
	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(conn).Encode(Request{Subscribe: true}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for !s.HasSubscribers() {
		if time.Now().After(deadline) {
			t.Fatal("subscriber never registered")
		}
		time.Sleep(time.Millisecond)
	}
	conn.Close()
	for s.HasSubscribers() {
		if time.Now().After(deadline) {
			t.Fatal("subscriber never unregistered after disconnect")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestSlowSubscriberGetsLatestState(t *testing.T) {
	s, _, path := newServer(t)
	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(Request{Subscribe: true}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for !s.HasSubscribers() {
		if time.Now().After(deadline) {
			t.Fatal("subscriber never registered")
		}
		time.Sleep(time.Millisecond)
	}
	// Broadcast many events without the client reading. The single-slot
	// mailbox means intermediate states are dropped; the client must
	// eventually observe the final state.
	for i := 1; i <= 100; i++ {
		s.Broadcast([]byte(fmt.Sprintf(`{"event":"state","state":{"n":%d}}`+"\n", i)))
	}
	r := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var last string
	for {
		line, err := r.ReadString('\n')
		if line != "" {
			last = line
		}
		if strings.Contains(last, `"n":100`) {
			break
		}
		if err != nil {
			t.Fatalf("never saw the final state; last = %q, err = %v", last, err)
		}
	}
}

func TestStaleSocketIsReplaced(t *testing.T) {
	// A leftover path that nothing is listening on (here: a plain file,
	// which fails to dial just like an orphaned socket inode does) must be
	// removed and replaced rather than aborting startup.
	path := filepath.Join(t.TempDir(), "w.sock")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	s, err := Listen(path, &echoHandler{}, nil)
	if err != nil {
		t.Fatalf("Listen over a stale socket: %v", err)
	}
	defer s.Close()
	// And it must actually work.
	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestLiveSocketIsNotStolen(t *testing.T) {
	_, _, path := newServer(t)
	if _, err := Listen(path, &echoHandler{}, nil); err == nil {
		t.Fatal("second Listen on a live socket succeeded")
	}
}
