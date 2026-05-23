package core

// pointerOp tracks an in-progress interactive move or resize driven by
// river_seat_v1.op_delta events.
type pointerOp struct {
	window WindowID
	action PointerAction
	// start is the window's float rect when the op began; deltas are
	// applied relative to it so the op is stateless with respect to event
	// coalescing (op_delta carries the total delta since the op started).
	start Rect
}

// minimum window dimensions during an interactive resize when the client
// expresses no preference.
const (
	minResizeW = 50
	minResizeH = 50
)

// StartPointerOp begins an interactive move or resize of the given window.
// A tiled window is made floating first (keeping its current geometry) so
// it can be dragged freely. Returns false if the window does not exist or
// an op is already in progress.
func (m *Model) StartPointerOp(id WindowID, action PointerAction) bool {
	if m.op != nil {
		return false
	}
	w, ok := m.Windows[id]
	if !ok || w.FullscreenOn != 0 {
		return false
	}
	if action != PointerActionMove && action != PointerActionResize {
		return false
	}
	if !w.Floating {
		// Adopt the window's current tiled geometry as its float rect so
		// the transition to floating is visually seamless.
		arr := m.Arrange()
		if p, ok := arr.Placements[id]; ok && !p.Rect.Empty() {
			w.FloatRect = p.Rect
		}
		m.setFloating(w, true)
	}
	m.focusWindow(w)
	m.op = &pointerOp{window: id, action: action, start: w.FloatRect}
	m.markChanged()
	return true
}

// PointerOpInProgress reports whether an interactive op is active.
func (m *Model) PointerOpInProgress() bool { return m.op != nil }

// PointerOpDelta applies the cumulative pointer motion since the op
// started.
func (m *Model) PointerOpDelta(dx, dy int32) {
	if m.op == nil {
		return
	}
	w, ok := m.Windows[m.op.window]
	if !ok {
		m.op = nil
		return
	}
	var r Rect
	switch m.op.action {
	case PointerActionMove:
		r = Rect{X: m.op.start.X + dx, Y: m.op.start.Y + dy, W: m.op.start.W, H: m.op.start.H}
	case PointerActionResize:
		minW, minH := int32(minResizeW), int32(minResizeH)
		if w.MinW > 0 {
			minW = w.MinW
		}
		if w.MinH > 0 {
			minH = w.MinH
		}
		r = Rect{
			X: m.op.start.X,
			Y: m.op.start.Y,
			W: max32(m.op.start.W+dx, minW),
			H: max32(m.op.start.H+dy, minH),
		}
		if w.MaxW > 0 && r.W > w.MaxW {
			r.W = w.MaxW
		}
		if w.MaxH > 0 && r.H > w.MaxH {
			r.H = w.MaxH
		}
	}
	if r != w.FloatRect {
		w.FloatRect = r
		m.markChanged()
	}
}

// EndPointerOp finishes the interactive op, if any.
func (m *Model) EndPointerOp() {
	if m.op == nil {
		return
	}
	m.op = nil
	m.markChanged()
}

// PointerOpWindow returns the window being interactively manipulated, or 0.
func (m *Model) PointerOpWindow() WindowID {
	if m.op == nil {
		return 0
	}
	return m.op.window
}
