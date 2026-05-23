package core

import "testing"

func TestParseChord(t *testing.T) {
	cases := []struct {
		in      string
		mods    Modifiers
		sym     Keysym
		wantErr bool
	}{
		{"Super+j", ModSuper, 'j', false},
		{"Super+J", ModSuper, 'j', false}, // uppercase letters normalize to the unshifted keysym
		{"Super+Shift+J", ModSuper | ModShift, 'j', false},
		{"Super+Return", ModSuper, 0xff0d, false},
		{"super+return", ModSuper, 0xff0d, false}, // case-insensitive
		{"Ctrl+Alt+Delete", ModCtrl | ModAlt, 0xffff, false},
		{"Control+Mod1+t", ModCtrl | ModAlt, 't', false}, // alternate spellings
		{"Logo+1", ModSuper, '1', false},
		{"None+XF86AudioMute", 0, 0x1008ff12, false},
		{"Super+comma", ModSuper, ',', false},
		{"Super+F11", ModSuper, 0xffc8, false},
		{"Super+Page_Up", ModSuper, 0xff55, false},
		{"Super+é", ModSuper, 0xe9, false},       // latin-1
		{"Super+ж", ModSuper, 0x01000436, false}, // unicode rule
		{"Bogus+x", 0, 0, true},
		{"Super+NotAKey", 0, 0, true},
		{"Super+", 0, 0, true},
		{"", 0, 0, true},
	}
	for _, tc := range cases {
		mods, sym, err := ParseChord(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseChord(%q) succeeded, want error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseChord(%q): %v", tc.in, err)
			continue
		}
		if mods != tc.mods || sym != tc.sym {
			t.Errorf("ParseChord(%q) = %v, %#x; want %v, %#x", tc.in, mods, sym, tc.mods, tc.sym)
		}
	}
}

func TestParsePointerChord(t *testing.T) {
	mods, btn, err := ParsePointerChord("Super+Left")
	if err != nil || mods != ModSuper || btn != 0x110 {
		t.Errorf("ParsePointerChord(Super+Left) = %v, %#x, %v", mods, btn, err)
	}
	if _, _, err := ParsePointerChord("Super+NotAButton"); err == nil {
		t.Error("expected error for unknown button")
	}
}

func TestModifiersString(t *testing.T) {
	if got := (ModSuper | ModShift).String(); got != "Super+Shift" {
		t.Errorf("String() = %q", got)
	}
	if got := Modifiers(0).String(); got != "None" {
		t.Errorf("String() = %q", got)
	}
}

func TestBindCommands(t *testing.T) {
	m := twoOutputs()
	run(t, m, "bind", "Super+j", "focus", "next")
	run(t, m, "bind", "Super+Return", "spawn", "foot")
	if len(m.Bindings) != 2 {
		t.Fatalf("got %d bindings", len(m.Bindings))
	}
	// Rebinding the same chord replaces it.
	run(t, m, "bind", "Super+j", "focus", "prev")
	if len(m.Bindings) != 2 {
		t.Fatalf("rebind created a duplicate: %d bindings", len(m.Bindings))
	}
	b := m.Bindings[bindingKey{'j', ModSuper}]
	if len(b.Command) != 2 || b.Command[1] != "prev" {
		t.Errorf("rebind did not replace the command: %v", b.Command)
	}
	// Binding an unknown command fails at bind time.
	if _, err := m.Dispatch([]string{"bind", "Super+k", "frobnicate"}); err == nil {
		t.Error("binding an unknown command succeeded")
	}
	run(t, m, "unbind", "Super+j")
	if len(m.Bindings) != 1 {
		t.Fatalf("unbind left %d bindings", len(m.Bindings))
	}
	if _, err := m.Dispatch([]string{"unbind", "Super+j"}); err == nil {
		t.Error("unbinding a nonexistent chord succeeded")
	}
}

func TestSpawnQueues(t *testing.T) {
	m := NewModel()
	run(t, m, "spawn", "foot", "-e", "htop")
	if len(m.SpawnRequests) != 1 || m.SpawnRequests[0] != "foot -e htop" {
		t.Errorf("spawn requests = %q", m.SpawnRequests)
	}
	if m.Changed() {
		t.Error("spawn marked the model changed; it should not trigger a manage sequence")
	}
}

func TestPointerOpMove(t *testing.T) {
	m := twoOutputs()
	m.WindowAdded(10)
	m.WindowAdded(11)
	// Window 10 is tiled as the master at 0,0 1152x1080 (60% of 1920).
	if !m.StartPointerOp(10, PointerActionMove) {
		t.Fatal("StartPointerOp failed")
	}
	w := m.Windows[10]
	if !w.Floating {
		t.Fatal("window did not become floating for the move")
	}
	if w.FloatRect != (Rect{X: 0, Y: 0, W: 1152, H: 1080}) {
		t.Fatalf("float rect did not adopt the tiled geometry: %v", w.FloatRect)
	}
	m.PointerOpDelta(100, 50)
	if w.FloatRect.X != 100 || w.FloatRect.Y != 50 {
		t.Errorf("after delta: %v", w.FloatRect)
	}
	// Deltas are cumulative from the start, not incremental.
	m.PointerOpDelta(10, 10)
	if w.FloatRect.X != 10 || w.FloatRect.Y != 10 {
		t.Errorf("cumulative delta not applied from the start rect: %v", w.FloatRect)
	}
	m.EndPointerOp()
	if m.PointerOpInProgress() {
		t.Error("op still in progress after EndPointerOp")
	}
	// A second op cannot start while one is active.
	m.StartPointerOp(10, PointerActionMove)
	if m.StartPointerOp(11, PointerActionMove) {
		t.Error("second concurrent op started")
	}
}

func TestPointerOpResizeRespectsMinimum(t *testing.T) {
	m := twoOutputs()
	m.WindowAdded(10)
	run(t, m, "toggle-float")
	m.Windows[10].FloatRect = Rect{X: 100, Y: 100, W: 400, H: 300}
	m.WindowDimensionsHint(10, 200, 150, 0, 0)
	m.StartPointerOp(10, PointerActionResize)
	m.PointerOpDelta(-1000, -1000)
	w := m.Windows[10]
	if w.FloatRect.W != 200 || w.FloatRect.H != 150 {
		t.Errorf("resize below minimum: %v", w.FloatRect)
	}
	m.PointerOpDelta(100, 100)
	if w.FloatRect.W != 500 || w.FloatRect.H != 400 {
		t.Errorf("resize: %v", w.FloatRect)
	}
}
