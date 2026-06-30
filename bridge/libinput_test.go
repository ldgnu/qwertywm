package bridge

import (
	"strings"
	"testing"

	"qwertywm/wire"
)

const (
	libinputConfigEvDevice = 1

	libinputDeviceEvRemoved     = 0
	libinputDeviceEvInputDevice = 1

	libinputDeviceReqSetNaturalScroll = 11

	libinputResultEvSuccess     = 0
	libinputResultEvUnsupported = 1
)

// addLibinputDevice announces a libinput device linked to an existing input
// device and returns its object ID.
func (f *fakeRiver) addLibinputDevice(inputDevID uint32) uint32 {
	id := f.allocServerID("river_libinput_device_v1")
	e := &wire.Encoder{}
	e.PutUint(id)
	f.server.Send(f.libinputConfigID, libinputConfigEvDevice, e)
	e = &wire.Encoder{}
	e.PutObject(inputDevID)
	f.server.Send(id, libinputDeviceEvInputDevice, e)
	return id
}

// TestInputNaturalScroll checks that an input setting is translated into
// the corresponding libinput request for matching devices only, and is not
// re-sent once applied.
func TestInputNaturalScroll(t *testing.T) {
	f, b := newFakeRiver(t)
	f.addOutput(0, 0, 1000, 600)
	f.addSeat()
	// Two pointers: an Apple trackpad and a normal mouse.
	trackpadDev := f.allocServerID("river_input_device_v1")
	e := &wire.Encoder{}
	e.PutUint(trackpadDev)
	f.server.Send(f.inputManagerID, inputManagerEvInputDevice, e)
	e = &wire.Encoder{}
	e.PutUint(1) // pointer
	f.server.Send(trackpadDev, inputDeviceEvType, e)
	e = &wire.Encoder{}
	e.PutString("Apple MTP multi-touch")
	f.server.Send(trackpadDev, inputDeviceEvName, e)
	trackpad := f.addLibinputDevice(trackpadDev)

	mouseDev := f.allocServerID("river_input_device_v1")
	e = &wire.Encoder{}
	e.PutUint(mouseDev)
	f.server.Send(f.inputManagerID, inputManagerEvInputDevice, e)
	e = &wire.Encoder{}
	e.PutString("Logitech MX Master")
	f.server.Send(mouseDev, inputDeviceEvName, e)
	f.addLibinputDevice(mouseDev)

	f.manageCycle()
	f.renderCycle()

	// Validation happens at command time.
	if _, err := b.runCommand([]string{"input", "*Apple*", "natural-scroll", "sideways"}); err == nil {
		t.Fatal("invalid value accepted")
	}
	if _, err := b.runCommand([]string{"input", "*Apple*", "warp-speed", "enabled"}); err == nil || !strings.Contains(err.Error(), "natural-scroll") {
		t.Fatalf("unknown property error should list the valid ones, got: %v", err)
	}

	if _, err := b.runCommand([]string{"input", "*Apple*", "natural-scroll", "disabled"}); err != nil {
		t.Fatal(err)
	}
	b.Dirty()
	f.collect()
	reqs := f.manageCycle()
	sets := find(reqs, "river_libinput_device_v1", libinputDeviceReqSetNaturalScroll)
	if len(sets) != 1 {
		t.Fatalf("got %d set_natural_scroll requests, want 1 (only the matching device): %v", len(sets), reqs)
	}
	if sets[0].object != trackpad {
		t.Errorf("set_natural_scroll sent to object %d, want the trackpad %d", sets[0].object, trackpad)
	}
	d := sets[0].decoder()
	resultID, _ := d.Uint()
	state, _ := d.Uint()
	if state != 0 {
		t.Errorf("natural scroll state = %d, want 0 (disabled)", state)
	}
	f.renderCycle()

	// The device acknowledges; nothing is re-sent on the next cycle.
	f.server.Send(resultID, libinputResultEvSuccess, &wire.Encoder{})
	reqs = f.manageCycle()
	if got := find(reqs, "river_libinput_device_v1", libinputDeviceReqSetNaturalScroll); len(got) != 0 {
		t.Errorf("set_natural_scroll re-sent after success")
	}

	// Changing the value re-sends it.
	b.runCommand([]string{"input", "*Apple*", "natural-scroll", "enabled"})
	b.Dirty()
	f.collect()
	reqs = f.manageCycle()
	sets = find(reqs, "river_libinput_device_v1", libinputDeviceReqSetNaturalScroll)
	if len(sets) != 1 {
		t.Fatalf("got %d set_natural_scroll after changing the value, want 1", len(sets))
	}
	d = sets[0].decoder()
	d.Uint()
	state, _ = d.Uint()
	if state != 1 {
		t.Errorf("natural scroll state = %d, want 1 (enabled)", state)
	}
}
