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
	"path/filepath"
	"time"

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
	configPath := flag.String("config", "", "config file path for hot-reload (default: $HOME/.config/qwertywm/config)")
	flag.Parse()

	if *showVersion {
		fmt.Println("weir", version)
		return
	}

	cfg := *configPath
	if cfg == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			cfg = filepath.Join(home, ".config", "qwertywm", "config")
		}
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(*logLevel)); err != nil {
		fmt.Fprintf(os.Stderr, "weir: invalid log level %q\n", *logLevel)
		os.Exit(2)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	if err := run(logger, *socket, cfg); err != nil {
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

func run(logger *slog.Logger, socketOverride, configPath string) error {
	conn, err := wire.Connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	model := core.NewModel()
	b := bridge.New(conn, model, logger)
	b.ConfigPath = configPath
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
	cmds := make(chan bridge.Command, 16)
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

	// Source the config on startup so it is applied even without the
	// session init script, and start watching it for hot-reload.
	if configPath != "" {
		// Queue initial config source via the reload-config command.
		reply := make(chan bridge.CommandResult, 1)
		cmds <- bridge.Command{Args: []string{"reload-config"}, Reply: reply}
		go func() {
			<-reply
			logger.Info("config loaded", "path", configPath)
		}()
		go watchConfigFile(configPath, cmds, logger)
	}

	logger.Info("weir started", "version", version)
	return b.Run(cmds)
}

// watchConfigFile polls the config file every 2 seconds and sends a
// reload-config command to the bridge when its mtime changes.
func watchConfigFile(path string, cmds chan<- bridge.Command, log *slog.Logger) {
	var lastMod time.Time
	for {
		fi, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) && !lastMod.IsZero() {
				// Config deleted; forget mtime so creation triggers reload.
				lastMod = time.Time{}
			}
			time.Sleep(2 * time.Second)
			continue
		}
		if fi.ModTime().After(lastMod) {
			lastMod = fi.ModTime()
			reply := make(chan bridge.CommandResult, 1)
			cmds <- bridge.Command{Args: []string{"reload-config"}, Reply: reply}
			<-reply
			log.Debug("config reloaded", "path", path)
		}
		time.Sleep(2 * time.Second)
	}
}
