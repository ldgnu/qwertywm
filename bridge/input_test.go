package bridge

import (
	"os"
	"testing"
	"time"

	"github.com/psanford/weir/core"
	"github.com/psanford/weir/wire"
)

// waitForFile polls for a file to exist, failing the test after a timeout.
func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("file %s never appeared", path)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Opcodes for the binding interfaces, from declaration order in the XML.
const (
	xkbBindingsReqDestroy       = 0
	xkbBindingsReqGetXkbBinding = 1
	xkbBindingsReqGetSeat       = 2

	xkbBindingReqDestroy           = 0
	xkbBindingReqSetLayoutOverride = 1
	xkbBindingReqEnable            = 2
	xkbBindingReqDisable           = 3

	xkbBindingEvPressed  = 0
	xkbBindingEvReleased = 1

	seatReqOpStartPointer    = 4
	seatReqOpEnd             = 5
	seatReqGetPointerBinding = 6

	seatEvOpDelta   = 6
	seatEvOpRelease = 7

	pointerBindingReqDestroy = 0
	pointerBindingReqEnable  = 1
	pointerBindingReqDisable = 2

	pointerBindingEvPressed  = 0
	pointerBindingEvReleased = 1
)

// TestKeyBindingLifecycle checks that binding a chord creates and enables a
// protocol binding object inside a manage sequence, that a pressed event
// runs the bound command, and that unbinding destroys the object.
func TestKeyBindingLifecycle(t *testing.T) {
	f, b := newFakeRiver(t)
	f.addOutput(0, 0, 1000, 600)
	f.addSeat()
	w1 := f.addWindow()
	w2 := f.addWindow()
	f.manageCycle()
	f.windowDimensions(w1, 600, 600)
	f.windowDimensions(w2, 400, 600)
	f.renderCycle()

	// Bind Super+j to focus next.
	if _, err := b.runCommand([]string{"bind", "Super+j", "focus", "next"}); err != nil {
		t.Fatal(err)
	}
	b.Dirty()
	f.collect()
	reqs := f.manageCycle()

	// The binding object is created and enabled during the manage
	// sequence.
	gets := find(reqs, "river_xkb_bindings_v1", xkbBindingsReqGetXkbBinding)
	if len(gets) != 1 {
		t.Fatalf("got %d get_xkb_binding requests, want 1: %v", len(gets), reqs)
	}
	d := gets[0].decoder()
	d.Object() // seat
	bindID, _ := d.Uint()
	keysym, _ := d.Uint()
	mods, _ := d.Uint()
	if keysym != 'j' || mods != uint32(core.ModSuper) {
		t.Errorf("binding keysym=%#x mods=%#x, want j/Super", keysym, mods)
	}
	if got := find(reqs, "river_xkb_binding_v1", xkbBindingReqEnable); len(got) != 1 {
		t.Fatalf("got %d enable requests, want 1", len(got))
	}
	f.renderCycle()

	// Focus is currently on window 2 (newest). Press Super+j: focus wraps
	// to window 1.
	f.server.Send(bindID, xkbBindingEvPressed, &wire.Encoder{})
	reqs = f.manageCycle()
	focus := find(reqs, "river_seat_v1", seatReqFocusWindow)
	if len(focus) != 1 || objectArg(t, focus[0]) != w1 {
		t.Fatalf("after pressing Super+j focus = %v, want window %d", focus, w1)
	}
	f.renderCycle()

	// Unbind: the protocol object is disabled and destroyed.
	if _, err := b.runCommand([]string{"unbind", "Super+j"}); err != nil {
		t.Fatal(err)
	}
	b.Dirty()
	f.collect()
	reqs = f.manageCycle()
	if got := find(reqs, "river_xkb_binding_v1", xkbBindingReqDisable); len(got) != 1 {
		t.Errorf("got %d disable requests, want 1", len(got))
	}
	if got := find(reqs, "river_xkb_binding_v1", xkbBindingReqDestroy); len(got) != 1 {
		t.Errorf("got %d destroy requests, want 1", len(got))
	}
}

// TestKeyBindingRebindKeepsObject checks that rebinding the same chord to a
// different command reuses the protocol object and runs the new command.
func TestKeyBindingRebindKeepsObject(t *testing.T) {
	f, b := newFakeRiver(t)
	f.addOutput(0, 0, 1000, 600)
	f.addSeat()
	f.manageCycle()
	f.renderCycle()

	b.runCommand([]string{"bind", "Super+x", "view", "3"})
	b.Dirty()
	f.collect()
	reqs := f.manageCycle()
	gets := find(reqs, "river_xkb_bindings_v1", xkbBindingsReqGetXkbBinding)
	if len(gets) != 1 {
		t.Fatalf("got %d get_xkb_binding, want 1", len(gets))
	}
	d := gets[0].decoder()
	d.Object()
	bindID, _ := d.Uint()
	f.renderCycle()

	b.runCommand([]string{"bind", "Super+x", "view", "7"})
	b.Dirty()
	f.collect()
	reqs = f.manageCycle()
	if got := find(reqs, "river_xkb_bindings_v1", xkbBindingsReqGetXkbBinding); len(got) != 0 {
		t.Errorf("rebinding created a new protocol object")
	}
	f.renderCycle()

	// Pressing it runs the new command.
	f.server.Send(bindID, xkbBindingEvPressed, &wire.Encoder{})
	f.manageCycle()
	if got := b.Model().Outputs[1].Workspace; got != "7" {
		t.Errorf("workspace after rebound press = %q, want 7", got)
	}
}

// TestPointerBindingMoveOp checks the full interactive move flow: press
// starts the op and floats the window, deltas move it, release ends the op.
func TestPointerBindingMoveOp(t *testing.T) {
	f, b := newFakeRiver(t)
	f.addOutput(0, 0, 1000, 600)
	f.addSeat()
	w1 := f.addWindow()
	w2 := f.addWindow()
	f.manageCycle()
	// w2 (the newer window) holds the main slot, w1 the stack slot.
	f.windowDimensions(w1, 400, 600)
	f.windowDimensions(w2, 600, 600)
	f.renderCycle()

	b.runCommand([]string{"bind-pointer", "Super+Left", "move"})
	b.Dirty()
	f.collect()
	reqs := f.manageCycle()
	gets := find(reqs, "river_seat_v1", seatReqGetPointerBinding)
	if len(gets) != 1 {
		t.Fatalf("got %d get_pointer_binding, want 1: %v", len(gets), reqs)
	}
	d := gets[0].decoder()
	pbID, _ := d.Uint()
	btn, _ := d.Uint()
	if btn != 0x110 {
		t.Errorf("bound button %#x, want BTN_LEFT", btn)
	}
	if got := find(reqs, "river_pointer_binding_v1", pointerBindingReqEnable); len(got) != 1 {
		t.Fatalf("pointer binding not enabled")
	}
	f.renderCycle()

	// Pointer enters the main window (w2), then the binding fires.
	e := &wire.Encoder{}
	e.PutObject(w2)
	f.server.Send(f.seatID, seatEvPointerEnter, e)
	f.server.Send(pbID, pointerBindingEvPressed, &wire.Encoder{})
	reqs = f.manageCycle()
	if got := find(reqs, "river_seat_v1", seatReqOpStartPointer); len(got) != 1 {
		t.Fatalf("op_start_pointer not sent: %v", reqs)
	}
	// The window became floating and keeps its tiled geometry.
	mw := b.Model().Windows[2]
	if !mw.Floating || mw.FloatRect != (core.Rect{X: 0, Y: 0, W: 600, H: 600}) {
		t.Fatalf("window not floating at its tiled rect: floating=%v rect=%v", mw.Floating, mw.FloatRect)
	}
	f.renderCycle()

	// Drag by (50, 30).
	e = &wire.Encoder{}
	e.PutInt(50)
	e.PutInt(30)
	f.server.Send(f.seatID, seatEvOpDelta, e)
	f.manageCycle()
	if mw.FloatRect.X != 50 || mw.FloatRect.Y != 30 {
		t.Errorf("after drag: %v", mw.FloatRect)
	}
	reqs = f.renderCycle()
	pos := find(reqs, "river_node_v1", nodeReqSetPosition)
	if len(pos) != 1 {
		t.Fatalf("got %d set_position during drag, want 1 (only the dragged window moved)", len(pos))
	}
	d = pos[0].decoder()
	x, _ := d.Int()
	y, _ := d.Int()
	if x != 50 || y != 30 {
		t.Errorf("dragged window positioned at %d,%d, want 50,30", x, y)
	}

	// Release ends the op.
	f.server.Send(f.seatID, seatEvOpRelease, &wire.Encoder{})
	reqs = f.manageCycle()
	if got := find(reqs, "river_seat_v1", seatReqOpEnd); len(got) != 1 {
		t.Fatalf("op_end not sent on release: %v", reqs)
	}
	if b.Model().PointerOpInProgress() {
		t.Error("model op still in progress after release")
	}
}

// TestSpawnExecutes checks that the spawn command actually runs a process.
func TestSpawnExecutes(t *testing.T) {
	f, b := newFakeRiver(t)
	f.addOutput(0, 0, 1000, 600)
	f.addSeat()
	f.manageCycle()
	f.renderCycle()

	marker := t.TempDir() + "/spawned"
	if _, err := b.runCommand([]string{"spawn", "touch " + marker}); err != nil {
		t.Fatal(err)
	}
	if len(b.Model().SpawnRequests) != 0 {
		t.Fatal("spawn request not drained")
	}
	waitForFile(t, marker)
}
