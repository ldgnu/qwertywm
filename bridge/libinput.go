package bridge

import (
	"qwertywm/core"
	"qwertywm/protocols/river"
)

// libinputDeviceState is the bridge's bookkeeping for one libinput device.
type libinputDeviceState struct {
	proxy *river.LibinputDeviceV1
	// device is the core ID of the corresponding input device, or 0 until
	// the input_device event arrives.
	device core.InputDeviceID
	// applied records the value last sent for each property so settings
	// are only re-sent when they change.
	applied map[string]string
}

// installLibinputHandlers wires up the libinput config global. Called from
// Bootstrap after the global is bound.
func (b *Bridge) installLibinputHandlers() {
	if b.libinputConfig == nil {
		return
	}
	b.libinputConfig.OnLibinputDevice = func(d *river.LibinputDeviceV1) {
		st := &libinputDeviceState{proxy: d, applied: make(map[string]string)}
		b.libinputDevices[d] = st
		d.OnInputDevice = func(dev *river.InputDeviceV1) {
			for id, ds := range b.inputDevices {
				if ds.proxy == dev {
					st.device = id
					break
				}
			}
		}
		d.OnRemoved = func() {
			delete(b.libinputDevices, d)
			d.Destroy()
		}
	}
}

// syncInputSettings applies the model's desired libinput settings to every
// device whose current configuration differs.
func (b *Bridge) syncInputSettings() {
	if b.libinputConfig == nil {
		return
	}
	for _, ld := range b.libinputDevices {
		dev, ok := b.model.InputDevices[ld.device]
		if !ok {
			continue
		}
		for property := range libinputSetters {
			setting, ok := b.model.SettingForDevice(dev.Name, property)
			if !ok || ld.applied[property] == setting.Value {
				continue
			}
			val, ok := core.InputValueIndex(property, setting.Value)
			if !ok {
				continue
			}
			result := libinputSetters[property](ld.proxy, val)
			ld.applied[property] = setting.Value
			b.watchResult(result, dev.Name, property, setting.Value)
		}
	}
}

// watchResult logs the outcome of a libinput configuration request. The
// result object is destroyed by the server after exactly one of these
// events.
func (b *Bridge) watchResult(r *river.LibinputResultV1, device, property, value string) {
	r.OnSuccess = func() {
		b.log.Debug("input setting applied", "device", device, "property", property, "value", value)
	}
	r.OnUnsupported = func() {
		b.log.Warn("input setting not supported by the device", "device", device, "property", property)
	}
	r.OnInvalid = func() {
		b.log.Warn("input setting rejected as invalid", "device", device, "property", property, "value", value)
	}
}

// libinputSetters maps property names to the protocol request that sets
// them. The uint32 value is the property's enum value as returned by
// core.InputValueIndex.
var libinputSetters = map[string]func(*river.LibinputDeviceV1, uint32) *river.LibinputResultV1{
	"natural-scroll": func(d *river.LibinputDeviceV1, v uint32) *river.LibinputResultV1 {
		return d.SetNaturalScroll(river.LibinputDeviceV1NaturalScrollState(v))
	},
	"tap": func(d *river.LibinputDeviceV1, v uint32) *river.LibinputResultV1 {
		return d.SetTap(river.LibinputDeviceV1TapState(v))
	},
	"drag": func(d *river.LibinputDeviceV1, v uint32) *river.LibinputResultV1 {
		return d.SetDrag(river.LibinputDeviceV1DragState(v))
	},
	"drag-lock": func(d *river.LibinputDeviceV1, v uint32) *river.LibinputResultV1 {
		return d.SetDragLock(river.LibinputDeviceV1DragLockState(v))
	},
	"left-handed": func(d *river.LibinputDeviceV1, v uint32) *river.LibinputResultV1 {
		return d.SetLeftHanded(river.LibinputDeviceV1LeftHandedState(v))
	},
	"middle-emulation": func(d *river.LibinputDeviceV1, v uint32) *river.LibinputResultV1 {
		return d.SetMiddleEmulation(river.LibinputDeviceV1MiddleEmulationState(v))
	},
	"dwt": func(d *river.LibinputDeviceV1, v uint32) *river.LibinputResultV1 {
		return d.SetDwt(river.LibinputDeviceV1DwtState(v))
	},
	"dwtp": func(d *river.LibinputDeviceV1, v uint32) *river.LibinputResultV1 {
		return d.SetDwtp(river.LibinputDeviceV1DwtpState(v))
	},
	"accel-profile": func(d *river.LibinputDeviceV1, v uint32) *river.LibinputResultV1 {
		return d.SetAccelProfile(river.LibinputDeviceV1AccelProfile(v))
	},
}
