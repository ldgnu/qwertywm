package core

import "testing"

// TestAbsoluteSetAppliesToEveryWorkspace reproduces the "workspace 1 has
// gaps but the others don't" bug: an init script's absolute set commands
// must configure every workspace and the default for new ones, not just
// whichever workspace happens to be focused.
func TestAbsoluteSetAppliesToEveryWorkspace(t *testing.T) {
	m := twoOutputs()
	run(t, m, "set", "gaps", "2", "2")
	run(t, m, "set", "main-ratio", "0.55")
	run(t, m, "set", "smart-gaps", "on")
	run(t, m, "set", "main-location", "top")

	// Only the workspaces shown on the two outputs exist at this point;
	// the rest are created on demand and must inherit via DefaultParams.
	for _, name := range []string{"1", "2"} {
		ws := m.Workspaces[name]
		if ws == nil {
			t.Fatalf("workspace %q does not exist", name)
		}
		if ws.Params.InnerGap != 2 || ws.Params.OuterGap != 2 {
			t.Errorf("workspace %q gaps = %d,%d, want 2,2", name, ws.Params.InnerGap, ws.Params.OuterGap)
		}
		if ws.Params.MainRatio != 0.55 {
			t.Errorf("workspace %q main ratio = %v, want 0.55", name, ws.Params.MainRatio)
		}
		if !ws.Params.SmartGaps {
			t.Errorf("workspace %q smart gaps not set", name)
		}
		if ws.Params.MainLocation != MainTop {
			t.Errorf("workspace %q main location = %v, want top", name, ws.Params.MainLocation)
		}
	}

	// Workspaces created after the configuration (including the default
	// 3-9, which only come into existence when first viewed) inherit it.
	for _, name := range []string{"9", "scratch"} {
		run(t, m, "view", name)
		if got := m.Workspaces[name].Params.InnerGap; got != 2 {
			t.Errorf("workspace %q inner gap = %d, want 2 (inherited from the default)", name, got)
		}
		if got := m.Workspaces[name].Params.MainRatio; got != 0.55 {
			t.Errorf("workspace %q main ratio = %v, want 0.55", name, got)
		}
	}
}

// TestRelativeSetAppliesToFocusedWorkspaceOnly checks that +/- adjustments
// (the keybinding form) keep their per-workspace behavior.
func TestRelativeSetAppliesToFocusedWorkspaceOnly(t *testing.T) {
	m := twoOutputs()
	run(t, m, "set", "main-ratio", "0.5")
	// Adjust only the focused workspace (workspace 1 on DP-1).
	run(t, m, "set", "main-ratio", "+0.1")
	run(t, m, "set", "main-count", "+1")
	if got := m.Workspaces["1"].Params.MainRatio; got != 0.6 {
		t.Errorf("focused workspace ratio = %v, want 0.6", got)
	}
	if got := m.Workspaces["2"].Params.MainRatio; got != 0.5 {
		t.Errorf("other workspace ratio = %v, want 0.5 (unaffected by the relative adjustment)", got)
	}
	if got := m.Workspaces["2"].Params.MainCount; got != 1 {
		t.Errorf("other workspace main count = %d, want 1", got)
	}
	// The default for new workspaces is also unaffected by relative
	// adjustments.
	if got := m.DefaultParams.MainRatio; got != 0.5 {
		t.Errorf("default ratio = %v, want 0.5", got)
	}
}

// TestBorderInset checks that tiled windows are shrunk by the border width
// so the compositor-drawn border (which extends outward from the content)
// stays within the usable area instead of landing off screen or under a
// bar's exclusive zone.
func TestBorderInset(t *testing.T) {
	m := twoOutputs()
	m.Borders.Width = 2
	m.WindowAdded(10)
	arr := m.Arrange()
	if got := arr.Placements[10].Rect; got != (Rect{X: 2, Y: 2, W: 1916, H: 1076}) {
		t.Errorf("single window with a 2px border = %v, want 1916x1076 at 2,2", got)
	}

	// Two windows: each is inset within its own slot, so the borders
	// between them and at every screen edge all fit on screen. The newer
	// window (11) takes the main slot.
	m.WindowAdded(11)
	arr = m.Arrange()
	if got := arr.Placements[11].Rect; got != (Rect{X: 2, Y: 2, W: 1148, H: 1076}) {
		t.Errorf("master = %v, want 1148x1076 at 2,2", got)
	}
	if got := arr.Placements[10].Rect; got != (Rect{X: 1154, Y: 2, W: 764, H: 1076}) {
		t.Errorf("stack = %v, want 764x1076 at 1154,2", got)
	}

	// A reduced usable area (a bar reserving the top 30px) keeps the top
	// border below the bar.
	m.OutputUsableArea(1, Rect{X: 0, Y: 30, W: 1920, H: 1050})
	m.WindowClosed(11)
	arr = m.Arrange()
	if got := arr.Placements[10].Rect; got != (Rect{X: 2, Y: 32, W: 1916, H: 1046}) {
		t.Errorf("window under a bar = %v, want 1916x1046 at 2,32 (top border at y=30..32)", got)
	}

	// Smart borders with a single window: no border, no inset.
	m.Borders.SmartBorders = true
	arr = m.Arrange()
	if got := arr.Placements[10].Rect; got != (Rect{X: 0, Y: 30, W: 1920, H: 1050}) {
		t.Errorf("smart-borderless window = %v, want the full usable area", got)
	}
	m.Borders.SmartBorders = false

	// Border width 0: no inset.
	m.Borders.Width = 0
	arr = m.Arrange()
	if got := arr.Placements[10].Rect; got != (Rect{X: 0, Y: 30, W: 1920, H: 1050}) {
		t.Errorf("borderless window = %v, want the full usable area", got)
	}
}
