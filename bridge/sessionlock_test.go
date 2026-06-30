package bridge

import (
	"testing"

	"qwertywm/wire"
)

// TestSessionUnlockReassertsFocus checks that keyboard focus is re-sent to
// the focused window after the session lock (swaylock) releases it, and
// that no focus requests are wasted while the lock holds the keyboard.
func TestSessionUnlockReassertsFocus(t *testing.T) {
	f, _ := newFakeRiver(t)
	f.addOutput(0, 0, 1000, 600)
	f.addSeat()
	w1 := f.addWindow()
	reqs := f.manageCycle()
	if got := find(reqs, "river_seat_v1", seatReqFocusWindow); len(got) != 1 {
		t.Fatalf("setup: expected an initial focus_window")
	}
	f.windowDimensions(w1, 1000, 600)
	f.renderCycle()

	// The session locks. No focus requests while locked, even across
	// manage sequences.
	f.server.Send(f.wmID, wmEvSessionLocked, &wire.Encoder{})
	reqs = f.manageCycle()
	if got := find(reqs, "river_seat_v1", seatReqFocusWindow); len(got) != 0 {
		t.Errorf("focus_window sent while the session is locked")
	}
	f.renderCycle()

	// Unlock: the manage sequence that follows must re-assert focus on
	// the same window even though the model's focus never changed.
	f.server.Send(f.wmID, wmEvSessionUnlocked, &wire.Encoder{})
	reqs = f.manageCycle()
	focus := find(reqs, "river_seat_v1", seatReqFocusWindow)
	if len(focus) != 1 || objectArg(t, focus[0]) != w1 {
		t.Fatalf("focus not re-asserted after session unlock: %v", reqs)
	}
}
