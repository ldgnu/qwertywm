// Package ipc implements weir's control socket: a unix socket speaking
// newline-delimited JSON over which every weir command can be executed,
// state can be queried, and state changes can be subscribed to.
//
// The protocol is a single request object per connection for commands and
// queries, or a long-lived stream for subscriptions:
//
//	-> {"command": ["focus", "next"]}
//	<- {"success": true}
//
//	-> {"command": ["get", "state"]}
//	<- {"success": true, "output": "{ ... }"}
//
//	-> {"subscribe": true}
//	<- {"event": "state", "state": { ... }}
//	<- {"event": "state", "state": { ... }}   (one line per change)
//
// A connection may send multiple command requests sequentially; each gets
// exactly one response line. Once a connection subscribes it receives only
// event lines and any further requests on it are ignored.
package ipc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Request is one line sent by a client.
type Request struct {
	// Command is an argv-style weir command, e.g. ["focus", "next"].
	Command []string `json:"command,omitempty"`
	// Subscribe switches the connection into event streaming mode.
	Subscribe bool `json:"subscribe,omitempty"`
}

// Response is one line sent by the server in reply to a command request.
type Response struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Event is one line sent by the server on a subscribed connection.
type Event struct {
	Event string `json:"event"`
	// State is the full weir state snapshot, present for "state" events.
	// It is kept as a pre-marshalled raw message so the snapshot is
	// serialized exactly once per change regardless of subscriber count.
	State any `json:"state,omitempty"`
}

// SocketPath returns the path of the weir control socket for the current
// environment. Resolution order:
//
//  1. $WEIRSOCK if set.
//  2. $XDG_RUNTIME_DIR/weir.$WAYLAND_DISPLAY.sock
//
// weir (the server) and weirctl (the client) must agree on this, so both
// call this function.
func SocketPath() (string, error) {
	if p := os.Getenv("WEIRSOCK"); p != "" {
		return p, nil
	}
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		return "", fmt.Errorf("XDG_RUNTIME_DIR is not set")
	}
	display := os.Getenv("WAYLAND_DISPLAY")
	if display == "" {
		display = "wayland-0"
	}
	// WAYLAND_DISPLAY may be an absolute path; only the final component
	// matters for distinguishing sessions.
	display = filepath.Base(display)
	// Defensive: a display name should never contain a path separator or
	// start with a dot after Base, but never build a hidden or traversing
	// path from environment input.
	display = strings.TrimLeft(display, ".")
	if display == "" {
		display = "wayland-0"
	}
	return filepath.Join(dir, "weir."+display+".sock"), nil
}
