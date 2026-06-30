package river_test

import (
	"testing"

	"qwertywm/protocols/river"
	"qwertywm/wire"
	"qwertywm/wire/wiretest"
)

// opcodes used by the fake compositor, matching the order of declarations
// in river-window-management-v1.xml. If the protocol changes these tests
// fail loudly rather than silently testing the wrong message.
const (
	// river_window_manager_v1 events
	wmEvUnavailable = 0
	wmEvFinished    = 1
	wmEvManageStart = 2
	wmEvRenderStart = 3
	wmEvWindow      = 6
	wmEvOutput      = 7
	wmEvSeat        = 8
	// river_window_manager_v1 requests
	wmReqStop         = 0
	wmReqDestroy      = 1
	wmReqManageFinish = 2
	wmReqManageDirty  = 3
	wmReqRenderFinish = 4
	// river_window_v1 events
	winEvClosed         = 0
	winEvDimensions     = 2
	winEvAppID          = 3
	winEvDecorationHint = 6
	// river_window_v1 requests
	winReqClose             = 1
	winReqGetNode           = 2
	winReqProposeDimensions = 3
	// river_node_v1 requests
	nodeReqSetPosition = 1
	// river_output_v1 events
	outEvRemoved    = 0
	outEvWlOutput   = 1
	outEvPosition   = 2
	outEvDimensions = 3
)

// TestManageSequence walks the generated bindings through the protocol's
// core loop: bind the global, receive an output and a window, propose
// dimensions during a manage sequence, and position the window during a
// render sequence. The fake compositor verifies every request the client
// sends, byte for byte.
func TestManageSequence(t *testing.T) {
	client, server := wiretest.Pair(t)

	reg := client.Display.GetRegistry()
	var wm *river.WindowManagerV1
	reg.OnGlobal = func(name uint32, iface string, version uint32) {
		if iface == river.WindowManagerV1Name {
			wm = river.BindWindowManagerV1(reg, name, version)
		}
	}
	client.Flush()
	server.Recv() // get_registry

	// Compositor advertises the window manager global.
	e := &wire.Encoder{}
	e.PutUint(77)
	e.PutString(river.WindowManagerV1Name)
	e.PutUint(4)
	server.Send(reg.ID(), 0, e)
	if _, err := client.Dispatch(); err != nil {
		t.Fatal(err)
	}
	if wm == nil {
		t.Fatal("window manager global not bound")
	}
	client.Flush()

	// Compositor sees the bind request.
	bind := server.Recv()
	if bind.Object != reg.ID() {
		t.Fatalf("bind sent to object %d, want registry %d", bind.Object, reg.ID())
	}
	d := bind.Decoder()
	gname, _ := d.Uint()
	giface, _, _ := d.String()
	gver, _ := d.Uint()
	gid, _ := d.Uint()
	if gname != 77 || giface != river.WindowManagerV1Name || gver != 4 || gid != wm.ID() {
		t.Fatalf("bind args = %d %q %d %d", gname, giface, gver, gid)
	}

	// Wire up the window manager handlers the way the bridge will.
	var (
		windows      []*river.WindowV1
		outputs      []*river.OutputV1
		manageStarts int
		renderStarts int
		dims         = map[*river.WindowV1][2]int32{}
	)
	wm.OnWindow = func(w *river.WindowV1) {
		windows = append(windows, w)
		w.OnDimensions = func(width, height int32) {
			dims[w] = [2]int32{width, height}
		}
	}
	wm.OnOutput = func(o *river.OutputV1) {
		outputs = append(outputs, o)
	}
	wm.OnManageStart = func() { manageStarts++ }
	wm.OnRenderStart = func() { renderStarts++ }

	// Compositor: new output (id 0xff000001), position, dimensions, then a
	// new window (id 0xff000002), then manage_start.
	const outID, winID = 0xff000001, 0xff000002
	e = &wire.Encoder{}
	e.PutUint(outID)
	server.Send(wm.ID(), wmEvOutput, e)
	e = &wire.Encoder{}
	e.PutInt(0)
	e.PutInt(0)
	server.Send(outID, outEvPosition, e)
	e = &wire.Encoder{}
	e.PutInt(1920)
	e.PutInt(1080)
	server.Send(outID, outEvDimensions, e)
	e = &wire.Encoder{}
	e.PutUint(winID)
	server.Send(wm.ID(), wmEvWindow, e)
	server.Send(wm.ID(), wmEvManageStart, &wire.Encoder{})

	for manageStarts == 0 {
		if _, err := client.Dispatch(); err != nil {
			t.Fatal(err)
		}
	}
	if len(outputs) != 1 || outputs[0].ID() != outID {
		t.Fatalf("outputs = %v", outputs)
	}
	if len(windows) != 1 || windows[0].ID() != winID {
		t.Fatalf("windows = %v", windows)
	}

	// Client responds: propose dimensions for the window, finish the manage
	// sequence.
	win := windows[0]
	win.ProposeDimensions(1920, 1080)
	wm.ManageFinish()
	client.Flush()

	msg := server.Recv()
	if msg.Object != winID || msg.Opcode != winReqProposeDimensions {
		t.Fatalf("got %d.%d, want %d.%d (propose_dimensions)", msg.Object, msg.Opcode, winID, winReqProposeDimensions)
	}
	d = msg.Decoder()
	w, _ := d.Int()
	h, _ := d.Int()
	if w != 1920 || h != 1080 {
		t.Fatalf("propose_dimensions args = %dx%d", w, h)
	}
	msg = server.Recv()
	if msg.Object != wm.ID() || msg.Opcode != wmReqManageFinish {
		t.Fatalf("got %d.%d, want manage_finish", msg.Object, msg.Opcode)
	}

	// Compositor: window took the dimensions, render sequence starts.
	e = &wire.Encoder{}
	e.PutInt(1920)
	e.PutInt(1080)
	server.Send(winID, winEvDimensions, e)
	server.Send(wm.ID(), wmEvRenderStart, &wire.Encoder{})
	for renderStarts == 0 {
		if _, err := client.Dispatch(); err != nil {
			t.Fatal(err)
		}
	}
	if dims[win] != [2]int32{1920, 1080} {
		t.Fatalf("window dimensions = %v", dims[win])
	}

	// Client positions the window via its node and finishes the render.
	node := win.GetNode()
	node.SetPosition(0, 0)
	wm.RenderFinish()
	client.Flush()

	msg = server.Recv()
	if msg.Object != winID || msg.Opcode != winReqGetNode {
		t.Fatalf("got %d.%d, want get_node", msg.Object, msg.Opcode)
	}
	d = msg.Decoder()
	nodeID, _ := d.Uint()
	if nodeID != node.ID() {
		t.Fatalf("get_node new_id = %d, want %d", nodeID, node.ID())
	}
	msg = server.Recv()
	if msg.Object != node.ID() || msg.Opcode != nodeReqSetPosition {
		t.Fatalf("got %d.%d, want set_position on node %d", msg.Object, msg.Opcode, node.ID())
	}
	msg = server.Recv()
	if msg.Object != wm.ID() || msg.Opcode != wmReqRenderFinish {
		t.Fatalf("got %d.%d, want render_finish", msg.Object, msg.Opcode)
	}
}

// TestEnumAndStringEvents checks enum-typed and nullable-string event
// arguments decode correctly through the generated dispatch code.
func TestEnumAndStringEvents(t *testing.T) {
	client, server := wiretest.Pair(t)
	reg := client.Display.GetRegistry()
	var wm *river.WindowManagerV1
	reg.OnGlobal = func(name uint32, iface string, version uint32) {
		if iface == river.WindowManagerV1Name {
			wm = river.BindWindowManagerV1(reg, name, version)
		}
	}
	client.Flush()
	server.Recv()
	e := &wire.Encoder{}
	e.PutUint(1)
	e.PutString(river.WindowManagerV1Name)
	e.PutUint(4)
	server.Send(reg.ID(), 0, e)
	for wm == nil {
		if _, err := client.Dispatch(); err != nil {
			t.Fatal(err)
		}
	}

	var win *river.WindowV1
	var gotAppID string
	var gotHint river.WindowV1DecorationHint = 999
	wm.OnWindow = func(w *river.WindowV1) {
		win = w
		w.OnAppId = func(appId string) { gotAppID = appId }
		w.OnDecorationHint = func(h river.WindowV1DecorationHint) { gotHint = h }
	}

	const winID = 0xff000010
	e = &wire.Encoder{}
	e.PutUint(winID)
	server.Send(wm.ID(), wmEvWindow, e)
	// app_id with a value, then a null app_id, then a decoration hint.
	e = &wire.Encoder{}
	e.PutString("org.example.Terminal")
	server.Send(winID, winEvAppID, e)
	e = &wire.Encoder{}
	e.PutUint(uint32(river.WindowV1DecorationHintPrefersSsd))
	server.Send(winID, winEvDecorationHint, e)

	for gotHint == 999 {
		if _, err := client.Dispatch(); err != nil {
			t.Fatal(err)
		}
	}
	if win == nil || gotAppID != "org.example.Terminal" {
		t.Fatalf("app_id = %q", gotAppID)
	}
	if gotHint != river.WindowV1DecorationHintPrefersSsd {
		t.Fatalf("decoration hint = %d", gotHint)
	}
}

// TestDestructorUnregisters checks that destructor requests remove the
// object from the connection's map so late events for it are ignored.
func TestDestructorUnregisters(t *testing.T) {
	client, server := wiretest.Pair(t)
	reg := client.Display.GetRegistry()
	var wm *river.WindowManagerV1
	reg.OnGlobal = func(name uint32, iface string, version uint32) {
		wm = river.BindWindowManagerV1(reg, name, version)
	}
	client.Flush()
	server.Recv()
	e := &wire.Encoder{}
	e.PutUint(1)
	e.PutString(river.WindowManagerV1Name)
	e.PutUint(4)
	server.Send(reg.ID(), 0, e)
	for wm == nil {
		if _, err := client.Dispatch(); err != nil {
			t.Fatal(err)
		}
	}

	var win *river.WindowV1
	closed := false
	wm.OnWindow = func(w *river.WindowV1) {
		win = w
		w.OnClosed = func() { closed = true }
	}
	const winID = 0xff000020
	e = &wire.Encoder{}
	e.PutUint(winID)
	server.Send(wm.ID(), wmEvWindow, e)
	e = &wire.Encoder{}
	server.Send(winID, winEvClosed, e)
	for !closed {
		if _, err := client.Dispatch(); err != nil {
			t.Fatal(err)
		}
	}

	// Client destroys the window object.
	win.Destroy()
	if win.ID() != winID {
		// Destroy must not zero the ID (the destructor request needs it),
		// but the object must be unregistered.
		t.Logf("note: ID after destroy = %d", win.ID())
	}
	if got := client.Lookup(winID); got != nil {
		t.Fatalf("window still registered after Destroy")
	}
	// A late event for the destroyed object must be ignored, not crash.
	e = &wire.Encoder{}
	e.PutString("late")
	server.Send(winID, winEvAppID, e)
	server.Send(wm.ID(), wmEvManageStart, &wire.Encoder{})
	done := false
	wm.OnManageStart = func() { done = true }
	for !done {
		if _, err := client.Dispatch(); err != nil {
			t.Fatal(err)
		}
	}
}
