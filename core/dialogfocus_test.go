package core

import "testing"

// TestDialogCloseReturnsFocusToParent reproduces the file-dialog flow: a
// browser in monocle opens a dialog (which floats above it and takes
// focus); when the dialog closes, focus — and therefore the top of the
// render order — must return to the browser, not to whichever window
// happens to occupy the vacated stack index.
func TestDialogCloseReturnsFocusToParent(t *testing.T) {
	m := twoOutputs()
	run(t, m, "set-layout", "monocle")

	m.WindowAdded(10) // the browser
	m.WindowAppID(10, "firefox")
	m.WindowAdded(11) // some other window, opened later
	m.WindowAppID(11, "foot")

	// The user is working in the browser.
	run(t, m, "focus", "main")
	if fw := m.FocusedWindow(); fw == nil || fw.ID != 10 {
		t.Fatalf("setup: focused = %v, want the browser", fw)
	}

	// The browser opens a file dialog: new window with the browser as its
	// parent. It floats and takes focus.
	m.WindowAdded(20)
	m.WindowParent(20, 10)
	if fw := m.FocusedWindow(); fw == nil || fw.ID != 20 {
		t.Fatalf("dialog did not take focus: %v", fw)
	}
	order := m.Arrange().Order
	if order[len(order)-1] != 20 {
		t.Fatalf("dialog is not on top: %v", order)
	}

	// The user picks a file; the dialog closes.
	m.WindowClosed(20)
	if fw := m.FocusedWindow(); fw == nil || fw.ID != 10 {
		t.Errorf("focus after the dialog closed = %v, want the browser (its parent)", fw)
	}
	order = m.Arrange().Order
	if order[len(order)-1] != 10 {
		t.Errorf("top of the render order after the dialog closed = %v, want the browser", order)
	}
}

// TestParentlessWindowCloseKeepsExistingBehavior documents that closing a
// focused window with no parent falls back to the existing stack-index
// behavior.
func TestParentlessWindowCloseKeepsExistingBehavior(t *testing.T) {
	m := twoOutputs()
	m.WindowAdded(10)
	m.WindowAdded(11)
	m.WindowAdded(12) // focused, no parent
	m.WindowClosed(12)
	if fw := m.FocusedWindow(); fw == nil || fw.ID != 11 {
		t.Errorf("focus = %v, want 11 (the last remaining window)", fw)
	}
}

// TestDialogCloseParentOnOtherWorkspace checks that focus does not jump
// across workspaces when a dialog's parent has been sent elsewhere.
func TestDialogCloseParentOnOtherWorkspace(t *testing.T) {
	m := twoOutputs()
	m.WindowAdded(10)
	m.WindowAdded(20)
	m.WindowParent(20, 10)
	// The parent gets sent to another workspace while its dialog is open.
	run(t, m, "focus", "main")
	run(t, m, "send", "5")
	// The dialog (still on workspace 1) closes; focus stays on workspace
	// 1's remaining content rather than following the parent to 5.
	m.WindowClosed(20)
	if m.Outputs[1].Workspace != "1" {
		t.Fatalf("workspace changed unexpectedly")
	}
	if fw := m.FocusedWindow(); fw != nil {
		t.Errorf("focused window = %v, want none (workspace 1 is empty)", fw.ID)
	}
}
