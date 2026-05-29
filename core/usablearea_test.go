package core

import "testing"

// TestScaleChangeKeepsWindowsBelowTheBar reproduces the reported bug: a
// scale change shrinks the output's logical size, and river delivers the
// updated layer-shell usable area *before* the new output dimensions in the
// same manage sequence. weir used to discard the freshly received area when
// the geometry event arrived, tiling windows underneath the bar until the
// bar was restarted.
func TestScaleChangeKeepsWindowsBelowTheBar(t *testing.T) {
	m := NewModel()
	m.Borders.Width = 0
	m.OutputAdded(1, "eDP-1", Rect{X: 0, Y: 0, W: 2560, H: 1600})
	m.WindowAdded(10)

	// waybar reserves a 30px strip at the top.
	m.OutputUsableArea(1, Rect{X: 0, Y: 30, W: 2560, H: 1570})
	if got := m.Arrange().Placements[10].Rect; got != (Rect{X: 0, Y: 30, W: 2560, H: 1570}) {
		t.Fatalf("setup: window = %v, want it below the bar", got)
	}

	// Scale change to 1.5: logical size becomes 1707x1067. River sends the
	// recomputed usable area first, then the new dimensions.
	m.OutputUsableArea(1, Rect{X: 0, Y: 30, W: 1707, H: 1037})
	m.OutputGeometry(1, Rect{X: 0, Y: 0, W: 1707, H: 1067})

	got := m.Arrange().Placements[10].Rect
	if got != (Rect{X: 0, Y: 30, W: 1707, H: 1037}) {
		t.Errorf("after the scale change: window = %v, want 1707x1037 at 0,30 (below the bar)", got)
	}

	// The reverse ordering (dimensions first, area second) must end up in
	// the same place.
	m.OutputGeometry(1, Rect{X: 0, Y: 0, W: 2560, H: 1600})
	m.OutputUsableArea(1, Rect{X: 0, Y: 30, W: 2560, H: 1570})
	got = m.Arrange().Placements[10].Rect
	if got != (Rect{X: 0, Y: 30, W: 2560, H: 1570}) {
		t.Errorf("after scaling back (geometry first): window = %v, want 2560x1570 at 0,30", got)
	}
}

// TestUsableAreaIsClampedToTheOutput checks that a stale usable area larger
// than the current output never lets windows extend beyond the output, and
// that an empty area resets to the full output.
func TestUsableAreaIsClampedToTheOutput(t *testing.T) {
	m := NewModel()
	m.Borders.Width = 0
	m.OutputAdded(1, "eDP-1", Rect{X: 0, Y: 0, W: 2560, H: 1600})
	m.WindowAdded(10)
	m.OutputUsableArea(1, Rect{X: 0, Y: 30, W: 2560, H: 1570})

	// The output shrinks with no fresh usable-area event at all (the
	// degenerate case): the stored area is clamped, the bar strip stays
	// reserved, and nothing extends beyond the output.
	m.OutputGeometry(1, Rect{X: 0, Y: 0, W: 1707, H: 1067})
	got := m.Arrange().Placements[10].Rect
	if got != (Rect{X: 0, Y: 30, W: 1707, H: 1037}) {
		t.Errorf("clamped usable area: window = %v, want 1707x1037 at 0,30", got)
	}

	// The bar goes away: an empty area means the full output is usable.
	m.OutputUsableArea(1, Rect{})
	got = m.Arrange().Placements[10].Rect
	if got != (Rect{X: 0, Y: 0, W: 1707, H: 1067}) {
		t.Errorf("after the bar left: window = %v, want the full output", got)
	}
}
