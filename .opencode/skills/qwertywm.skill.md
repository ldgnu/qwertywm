---
name: qwertywm
description: Window manager for river (dynamic tiling, Wayland, Go)
---

# qwertywm Skill

## Build & Install

```sh
# Build both binaries
cd ~/Proyectos/qwertywm
go build ./cmd/qwertywm
go build ./cmd/qwertywmctl
sudo cp qwertywm qwertywmctl /usr/local/bin/

# Or use the build script
./build.sh
sudo cp qwertywm qwertywmctl /usr/local/bin/
```

## Test

```sh
go test ./...
go vet ./...
go run ./cmd/wmsim example/two-outputs.txt
```

## Run against real river (nested)

```sh
go build ./cmd/qwertywm && river -c ./qwertywm
```

## Key files

| File | Purpose |
|------|---------|
| `core/model.go` | Core state machine (outputs, workspaces, windows) |
| `core/layout.go` | Layout interface + tile/monocle implementations |
| `core/commands.go` | Command dispatch table |
| `bridge/bridge.go` | River protocol adapter |
| `bridge/bindings.go` | Key/pointer binding handling |
| `ipc/handler.go` | Unix socket command handler |
| `cmd/qwertywm/main.go` | WM entry point |
| `cmd/qwertywmctl/main.go` | CLI entry point |

## Architecture rules

1. `core/` never imports Wayland — pure Go deterministic state machine
2. `bridge/` never makes policy decisions — translates events ↔ commands
3. All commands route through the same table (keybinds, IPC, pointer binds)

## Adding a command

1. Add function in `core/commands.go` or relevant file
2. Register in the command table
3. Add IPC handler in `ipc/handler.go` if needed
4. Add tests

## Adding a layout

1. Implement `core/layout.go`'s `Layout` interface
2. Register in `core/layout.go`'s `layouts` map
3. Test in `core/` and optionally `example/`

## User config

The user's qwertywm config is at `~/.config/qwertywm/config`.
River init is at `~/.config/river/init`.
Themes system is at `~/.config/themes/`.
