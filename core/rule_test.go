package core

import (
	"strings"
	"testing"
)

func TestRuleFloatByAppID(t *testing.T) {
	m := twoOutputs()
	run(t, m, "rule", "add", "-app-id", "float*", "float")
	run(t, m, "rule", "add", "-app-id", "mpv", "-title", "*- playlist", "float")

	m.WindowAdded(10)
	m.WindowAppID(10, "floating-calculator")
	if !m.Windows[10].Floating {
		t.Errorf("glob app-id rule did not float the window")
	}

	m.WindowAdded(11)
	m.WindowAppID(11, "mpv")
	if m.Windows[11].Floating {
		t.Errorf("two-criteria rule applied with only the app-id matching")
	}
	m.WindowTitle(11, "video.mkv - playlist")
	if !m.Windows[11].Floating {
		t.Errorf("two-criteria rule did not apply once both matched")
	}

	// A window that matches nothing is unaffected.
	m.WindowAdded(12)
	m.WindowAppID(12, "foot")
	if m.Windows[12].Floating {
		t.Errorf("non-matching window floated")
	}
}

func TestRuleDoesNotApplyToDisplayedWindows(t *testing.T) {
	m := twoOutputs()
	run(t, m, "rule", "add", "-title", "*important*", "float")
	m.WindowAdded(10)
	m.WindowAppID(10, "foot")
	// The window is displayed (the compositor reported dimensions), then
	// its title changes to something matching the rule. It must not
	// suddenly float.
	m.WindowDimensions(10, 960, 1080)
	m.WindowTitle(10, "very important document")
	if m.Windows[10].Floating {
		t.Errorf("rule applied to an already-displayed window on title change")
	}
}

func TestRuleNoFloatOverridesDialogAutoFloat(t *testing.T) {
	m := twoOutputs()
	run(t, m, "rule", "add", "-app-id", "gimp", "no-float")
	m.WindowAdded(10)
	m.WindowAppID(10, "gimp")
	m.WindowAdded(11)
	m.WindowAppID(11, "gimp")
	m.WindowParent(11, 10)
	if m.Windows[11].Floating {
		t.Errorf("no-float rule did not override the dialog auto-float")
	}
}

func TestRuleWorkspaceAndDecoration(t *testing.T) {
	m := twoOutputs()
	run(t, m, "rule", "add", "-app-id", "Slack", "workspace", "9")
	run(t, m, "rule", "add", "-app-id", "emacs", "ssd")
	m.WindowAdded(10)
	m.WindowAppID(10, "Slack")
	if m.Windows[10].Workspace != "9" {
		t.Errorf("workspace rule: window on %q, want 9", m.Windows[10].Workspace)
	}
	m.WindowAdded(11)
	m.WindowAppID(11, "emacs")
	if m.Windows[11].DecorationOverride != "ssd" {
		t.Errorf("ssd rule: override = %q", m.Windows[11].DecorationOverride)
	}
}

func TestRuleListAndDel(t *testing.T) {
	m := NewModel()
	run(t, m, "rule", "add", "-app-id", "a", "float")
	run(t, m, "rule", "add", "-title", "b", "csd")
	// Adding a duplicate is a no-op.
	run(t, m, "rule", "add", "-app-id", "a", "float")
	out := run(t, m, "rule", "list")
	if len(strings.Split(out, "\n")) != 2 {
		t.Errorf("rule list:\n%s", out)
	}
	run(t, m, "rule", "del", "-app-id", "a", "float")
	out = run(t, m, "rule", "list")
	if strings.Contains(out, "app-id") {
		t.Errorf("rule not deleted:\n%s", out)
	}
	if _, err := m.Dispatch([]string{"rule", "del", "-app-id", "nope", "float"}); err == nil {
		t.Error("deleting a nonexistent rule succeeded")
	}
	// Errors.
	for _, bad := range [][]string{
		{"rule", "add", "float"},                       // no criteria
		{"rule", "add", "-app-id", "x", "explode"},     // bad action
		{"rule", "add", "-app-id", "[", "float"},       // bad glob
		{"rule", "add", "-app-id", "x", "workspace"},   // missing arg
		{"rule", "add", "-app-id", "x", "float", "yo"}, // extra arg
	} {
		if _, err := m.Dispatch(bad); err == nil {
			t.Errorf("Dispatch(%v) succeeded, want error", bad)
		}
	}
}
