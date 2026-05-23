package bridge

import "testing"

// TestUndersizedWindowIsCenteredInItsSlot checks that a window that takes
// a smaller size than its tile slot (terminal cell snapping) is positioned
// so the remainder is split evenly around it rather than left entirely at
// the bottom right.
func TestUndersizedWindowIsCenteredInItsSlot(t *testing.T) {
	f, _ := newFakeRiver(t)
	f.addOutput(0, 0, 1000, 600)
	f.addSeat()
	w1 := f.addWindow()
	f.manageCycle()

	// The slot is 1000x600 but the terminal snaps to 994x588.
	f.windowDimensions(w1, 994, 588)
	reqs := f.renderCycle()
	pos := find(reqs, "river_node_v1", nodeReqSetPosition)
	if len(pos) != 1 {
		t.Fatalf("got %d set_position, want 1", len(pos))
	}
	d := pos[0].decoder()
	x, _ := d.Int()
	y, _ := d.Int()
	if x != 3 || y != 6 {
		t.Errorf("undersized window positioned at %d,%d, want 3,6 (centered in the slot)", x, y)
	}

	// A window that takes exactly its slot stays at the slot origin.
	f.windowDimensions(w1, 1000, 600)
	reqs = f.renderCycle()
	pos = find(reqs, "river_node_v1", nodeReqSetPosition)
	if len(pos) != 1 {
		t.Fatalf("got %d set_position after exact fit, want 1", len(pos))
	}
	d = pos[0].decoder()
	x, _ = d.Int()
	y, _ = d.Int()
	if x != 0 || y != 0 {
		t.Errorf("exact-fit window positioned at %d,%d, want 0,0", x, y)
	}
}
