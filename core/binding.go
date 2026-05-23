package core

import (
	"fmt"
	"sort"
	"strings"
)

// Binding is a key binding: a chord and the command it runs.
type Binding struct {
	Keysym  Keysym
	Mods    Modifiers
	Command []string
}

// Chord renders the binding's chord for display, e.g. "Super+Shift+Return".
func (b Binding) Chord() string {
	if b.Mods == 0 {
		return "None+" + KeysymName(b.Keysym)
	}
	return b.Mods.String() + "+" + KeysymName(b.Keysym)
}

// PointerAction is what a pointer binding does when triggered.
type PointerAction string

const (
	// PointerActionMove starts an interactive move of the window under the
	// pointer.
	PointerActionMove PointerAction = "move"
	// PointerActionResize starts an interactive resize of the window under
	// the pointer.
	PointerActionResize PointerAction = "resize"
	// PointerActionCommand runs a command.
	PointerActionCommand PointerAction = "command"
)

// PointerBinding is a pointer-button binding.
type PointerBinding struct {
	Button uint32 // Linux input event code
	Mods   Modifiers
	Action PointerAction
	// Command is the command to run when Action is PointerActionCommand.
	Command []string
}

// Chord renders the pointer binding's chord for display.
func (b PointerBinding) Chord() string {
	if b.Mods == 0 {
		return "None+" + ButtonName(b.Button)
	}
	return b.Mods.String() + "+" + ButtonName(b.Button)
}

// bindingKey uniquely identifies a key binding within the model.
type bindingKey struct {
	sym  Keysym
	mods Modifiers
}

// pointerBindingKey uniquely identifies a pointer binding within the model.
type pointerBindingKey struct {
	button uint32
	mods   Modifiers
}

// cmdBind implements: bind <chord> <command...>
func cmdBind(m *Model, args []string) (string, error) {
	if len(args) < 2 {
		return "", cmdErrf("usage: bind <mods+key> <command...>")
	}
	mods, sym, err := ParseChord(args[0])
	if err != nil {
		return "", err
	}
	if err := validateBindingCommand(args[1:]); err != nil {
		return "", err
	}
	m.Bindings[bindingKey{sym, mods}] = Binding{Keysym: sym, Mods: mods, Command: append([]string(nil), args[1:]...)}
	m.markChanged()
	return "", nil
}

// cmdUnbind implements: unbind <chord>
func cmdUnbind(m *Model, args []string) (string, error) {
	if len(args) != 1 {
		return "", cmdErrf("usage: unbind <mods+key>")
	}
	mods, sym, err := ParseChord(args[0])
	if err != nil {
		return "", err
	}
	key := bindingKey{sym, mods}
	if _, ok := m.Bindings[key]; !ok {
		return "", cmdErrf("no binding for %q", args[0])
	}
	delete(m.Bindings, key)
	m.markChanged()
	return "", nil
}

// cmdBindPointer implements: bind-pointer <chord> move|resize|<command...>
func cmdBindPointer(m *Model, args []string) (string, error) {
	if len(args) < 2 {
		return "", cmdErrf("usage: bind-pointer <mods+button> move|resize|<command...>")
	}
	mods, btn, err := ParsePointerChord(args[0])
	if err != nil {
		return "", err
	}
	pb := PointerBinding{Button: btn, Mods: mods}
	switch {
	case len(args) == 2 && args[1] == "move":
		pb.Action = PointerActionMove
	case len(args) == 2 && args[1] == "resize":
		pb.Action = PointerActionResize
	default:
		if err := validateBindingCommand(args[1:]); err != nil {
			return "", err
		}
		pb.Action = PointerActionCommand
		pb.Command = append([]string(nil), args[1:]...)
	}
	m.PointerBindings[pointerBindingKey{btn, mods}] = pb
	m.markChanged()
	return "", nil
}

// cmdUnbindPointer implements: unbind-pointer <chord>
func cmdUnbindPointer(m *Model, args []string) (string, error) {
	if len(args) != 1 {
		return "", cmdErrf("usage: unbind-pointer <mods+button>")
	}
	mods, btn, err := ParsePointerChord(args[0])
	if err != nil {
		return "", err
	}
	key := pointerBindingKey{btn, mods}
	if _, ok := m.PointerBindings[key]; !ok {
		return "", cmdErrf("no pointer binding for %q", args[0])
	}
	delete(m.PointerBindings, key)
	m.markChanged()
	return "", nil
}

// validateBindingCommand rejects commands that would never work when bound:
// unknown commands fail at bind time rather than silently at key-press time.
func validateBindingCommand(args []string) error {
	if len(args) == 0 || args[0] == "" {
		return cmdErrf("empty command")
	}
	for i := range commands {
		if commands[i].name == args[0] {
			return nil
		}
	}
	return cmdErrf("unknown command %q (try \"help\")", args[0])
}

// LookupBinding returns the key binding for a chord, if any. The bridge
// uses this to resolve a binding's command at press time so that rebinding
// a chord takes effect without recreating the protocol object.
func (m *Model) LookupBinding(sym Keysym, mods Modifiers) (Binding, bool) {
	b, ok := m.Bindings[bindingKey{sym, mods}]
	return b, ok
}

// LookupPointerBinding returns the pointer binding for a chord, if any.
func (m *Model) LookupPointerBinding(button uint32, mods Modifiers) (PointerBinding, bool) {
	b, ok := m.PointerBindings[pointerBindingKey{button, mods}]
	return b, ok
}

// cmdListBindings implements: list-bindings
func cmdListBindings(m *Model, _ []string) (string, error) {
	var lines []string
	for _, b := range m.Bindings {
		lines = append(lines, fmt.Sprintf("%-28s %s", b.Chord(), strings.Join(b.Command, " ")))
	}
	for _, b := range m.PointerBindings {
		action := string(b.Action)
		if b.Action == PointerActionCommand {
			action = strings.Join(b.Command, " ")
		}
		lines = append(lines, fmt.Sprintf("%-28s %s", "(pointer) "+b.Chord(), action))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n"), nil
}

// cmdSpawn implements: spawn <command...>
// The command is run with /bin/sh -c by the process that owns the model
// (the bridge); the model only queues the request.
func cmdSpawn(m *Model, args []string) (string, error) {
	if len(args) == 0 {
		return "", cmdErrf("usage: spawn <command...>")
	}
	// Note: deliberately not markChanged — spawning a process does not
	// affect window management state, so no manage sequence is needed.
	m.SpawnRequests = append(m.SpawnRequests, strings.Join(args, " "))
	return "", nil
}
