// Package weir is the repository root; see PLAN.md for the layout.
//
// The directives below regenerate the protocol bindings from the vendored
// XML in protocol/. Run "go generate ./..." after changing the generator or
// updating a protocol file.
package weir

// References from a river protocol to a core wayland interface (e.g.
// wl_surface) are intentionally untyped (wire.Object): the two namespaces
// are generated into separate packages.
//
//go:generate go run ./internal/gen -package wl -out protocols/wl/wl.go protocol/wayland.xml
//go:generate go run ./internal/gen -package river -out protocols/river/river.go protocol/river-window-management-v1.xml protocol/river-xkb-bindings-v1.xml protocol/river-xkb-config-v1.xml protocol/river-input-management-v1.xml protocol/river-libinput-config-v1.xml protocol/river-layer-shell-v1.xml
