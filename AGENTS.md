# qwertywm — AI Agent Guide

## Overview

qwertywm is an xmonad-style dynamic tiling window manager for the river Wayland
compositor (≥ 0.4). Written in Go. Forked from [weir](https://github.com/psanford/weir).

## Build

```sh
go build ./cmd/qwertywm       # WM binary
go build ./cmd/qwertywmctl    # CLI controller
go build ./cmd/wmsim          # ASCII simulator
./build.sh                    # both binaries at once
```

## Test

```sh
go test ./...                    # unit + property + protocol tests
go vet ./...                     # static analysis
go run ./cmd/wmsim example/two-outputs.txt  # ASCII sim
```

Integration tests (requires river headless):
```sh
eval "$(scripts/fetch-river.sh)"
scripts/smoke-test.sh
```

## Architecture

```
core/      Pure Go state machine (model + layout + commands). No Wayland imports.
bridge/    River protocol adapter. Translates protocol events ↔ core commands.
ipc/       Unix socket server. JSON command/query/subscribe.
cmd/qwertywm/    Main binary entry point.
cmd/qwertywmctl/ CLI binary entry point.
wire/      Pure-Go Wayland wire protocol (no cgo).
protocol/  Vendored protocol XML + generated bindings.
```

### Rules

1. `core/` never imports Wayland — deterministic state machine.
2. `bridge/` never makes policy decisions — only translates.
3. All actions dispatch through the same command table (keybinds, IPC, pointer).

## Key Types & Concepts

- **Output**: A physical or virtual monitor. Has a name, geometry, and shows one workspace.
- **Workspace**: A named, ordered stack of windows. Created on first reference, never destroyed.
- **Window**: Has app_id, title, parent, float/fullscreen flags, position in workspace stack.
- **Arrangement**: Complete desired state of every window (produced by core, consumed by bridge).
- **Workspace modes**: `independent` (each output has own workspaces) / `locked` (all outputs switch together).

## Commands

The command table is in `core/commands.go`. Add new commands there.
The binding system is in `bridge/bindings.go` and `ipc/handler.go`.

Key commands:
- `focus next|prev|main`
- `view <ws>` / `send <ws>` / `pull <ws>`
- `focus-output <dir|name>` / `send-to-output <dir|name>`
- `toggle-float` / `toggle-fullscreen`
- `cycle-layout <l1,l2,...>`
- `spawn <cmd>`
- `bind <mods+key> <cmd>` / `rule add <filter> <action>`
- `get state|outputs|windows|workspaces|settings`
- `subscribe`

## Adding a Layout

1. Define a struct implementing `core/layout.go`'s `Layout` interface.
2. Register it in `core/layout.go`'s `layouts` map.
3. Add tests in `core/` and optionally in `example/`.

## Codegen

```sh
go generate ./...   # regenerates protocol bindings from XML
```

Protocol XML is in `protocol/`. Generated code goes to `protocols/`.

## Dependencies

Minimal: Go 1.21, wayland, wayland-protocols, libxkbcommon.
No cgo, no external Go deps. Module path is `qwertywm`.

## Developing

```sh
# Quick feedback loop
go run ./cmd/wmsim        # interactive REPL
go run ./cmd/wmsim example/two-outputs.txt  # scripted sim

# Against real river (nested in your current session)
go build ./cmd/qwertywm && river -c ./qwertywm
```

## Project Standards

- No cgo, no external dependencies.
- Property tests in `core/invariants_test.go` enforce structural invariants.
- Protocol sequence tests in `bridge/` verify manage/render ordering.
- New features must include tests.
- Pure functions in core; side effects only in bridge/ipc.
