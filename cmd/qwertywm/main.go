// Command weir is a window manager for the river Wayland compositor.
//
// weir must be started by river (or another compositor implementing
// river-window-management-v1), typically from the river init script:
//
//	exec weir &
//
// It connects to the Wayland display named by the environment, takes the
// window manager role, and manages windows until the compositor exits.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"qwertywm/bridge"
	"qwertywm/core"
	"qwertywm/ipc"
	"qwertywm/wire"
)

var version = "0.1.0-dev"

func main() {
	logLevel := flag.String("log-level", "info", "log level: debug, info, warn, error")
	showVersion := flag.Bool("version", false, "print the version and exit")
	socket := flag.String("socket", "", "control socket path (default: derived from the environment)")
	flag.Parse()

	if *showVersion {
		fmt.Println("weir", version)
		return
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(*logLevel)); err != nil {
		fmt.Fprintf(os.Stderr, "weir: invalid log level %q\n", *logLevel)
		os.Exit(2)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	if err := run(logger, *socket); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

// commandClient adapts the IPC server's Handler interface to the bridge's
// command channel: each command is shipped to the bridge goroutine and the
// caller blocks until it has been executed there.
type commandClient struct {
	cmds chan<- bridge.Command
}

func (c *commandClient) Command(args []string) (string, error) {
	reply := make(chan bridge.CommandResult, 1)
	c.cmds <- bridge.Command{Args: args, Reply: reply}
	res := <-reply
	return res.Output, res.Err
}

func run(logger *slog.Logger, socketOverride string) error {
	conn, err := wire.Connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	model := core.NewModel()
	b := bridge.New(conn, model, logger)
	if err := b.Bootstrap(); err != nil {
		if errors.Is(err, bridge.ErrUnavailable) {
			return err
		}
		return fmt.Errorf("bootstrap: %w", err)
	}

	// Control socket. Commands arriving on it are executed on the bridge
	// goroutine via the command channel; state changes are broadcast to
	// subscribers from the bridge goroutine after each manage sequence.
	socketPath := socketOverride
	if socketPath == "" {
		socketPath, err = ipc.SocketPath()
		if err != nil {
			return err
		}
	}
	cmds := make(chan bridge.Command)
	srv, err := ipc.Listen(socketPath, &commandClient{cmds: cmds}, logger)
	if err != nil {
		return err
	}
	defer srv.Close()
	b.OnStateChange = func() {
		if !srv.HasSubscribers() {
			return
		}
		line, err := json.Marshal(ipc.Event{Event: "state", State: model.Snapshot()})
		if err != nil {
			logger.Error("marshal state event", "err", err)
			return
		}
		srv.Broadcast(append(line, '\n'))
	}

	logger.Info("weir started", "version", version)
	return b.Run(cmds)
}
