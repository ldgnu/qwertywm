package wire

import "fmt"

// Display is the wl_display singleton, always object 1. It is hand-written
// rather than generated because it bootstraps the object system and because
// wl_registry.bind has a unique wire encoding.
type Display struct {
	Proxy
	// OnError is called when the compositor reports a fatal protocol
	// error. The connection is unusable afterwards. If nil, the error is
	// still recorded on the connection and returned from the next call.
	OnError func(objectID uint32, code uint32, message string)
	// OnDeleteID is called when the server confirms destruction of an
	// object ID. Most clients have no use for this.
	OnDeleteID func(id uint32)
}

func (*Display) Interface() string { return "wl_display" }

// Sync sends wl_display.sync. The returned callback's OnDone fires once the
// compositor has processed all prior requests.
func (d *Display) Sync() *Callback {
	cb := &Callback{}
	d.conn.RegisterClient(cb, &cb.Proxy)
	e := &Encoder{}
	e.PutUint(cb.ID())
	d.conn.SendRequest(d, 0, e)
	return cb
}

// GetRegistry sends wl_display.get_registry and returns the new registry.
func (d *Display) GetRegistry() *Registry {
	r := &Registry{}
	d.conn.RegisterClient(r, &r.Proxy)
	e := &Encoder{}
	e.PutUint(r.ID())
	d.conn.SendRequest(d, 1, e)
	return r
}

func (d *Display) Dispatch(opcode uint16, dec *Decoder) error {
	switch opcode {
	case 0: // error
		objID, err := dec.Object()
		if err != nil {
			return err
		}
		code, err := dec.Uint()
		if err != nil {
			return err
		}
		msg, _, err := dec.String()
		if err != nil {
			return err
		}
		iface := "unknown"
		if obj := d.conn.Lookup(objID); obj != nil {
			iface = obj.Interface()
		}
		d.conn.fatal(fmt.Errorf("wire: compositor protocol error on %s@%d code %d: %s", iface, objID, code, msg))
		if d.OnError != nil {
			d.OnError(objID, code, msg)
		}
		return nil
	case 1: // delete_id
		id, err := dec.Uint()
		if err != nil {
			return err
		}
		// The object was already unregistered when the client sent the
		// destructor; delete_id for a server-created object (e.g. a
		// wl_callback after done) needs the cleanup here.
		delete(d.conn.objects, id)
		if d.OnDeleteID != nil {
			d.OnDeleteID(id)
		}
		return nil
	default:
		return fmt.Errorf("unknown opcode %d", opcode)
	}
}

// Registry is the wl_registry global listing object.
type Registry struct {
	Proxy
	// OnGlobal announces a global object available for binding.
	OnGlobal func(name uint32, iface string, version uint32)
	// OnGlobalRemove announces that a global has been removed.
	OnGlobalRemove func(name uint32)
}

func (*Registry) Interface() string { return "wl_registry" }

// Bind binds the global with the given numeric name to obj, registering obj
// under a fresh client ID. iface and version must match what the client
// actually implements (they are not taken from the global event so the
// caller can bind a lower version than advertised).
func (r *Registry) Bind(name uint32, iface string, version uint32, obj Object, p *Proxy) {
	r.conn.RegisterClient(obj, p)
	e := &Encoder{}
	e.PutUint(name)
	// new_id without a fixed interface: string interface, uint version,
	// uint id.
	e.PutString(iface)
	e.PutUint(version)
	e.PutUint(obj.ID())
	r.conn.SendRequest(r, 0, e)
}

func (r *Registry) Dispatch(opcode uint16, dec *Decoder) error {
	switch opcode {
	case 0: // global
		name, err := dec.Uint()
		if err != nil {
			return err
		}
		iface, _, err := dec.String()
		if err != nil {
			return err
		}
		version, err := dec.Uint()
		if err != nil {
			return err
		}
		if r.OnGlobal != nil {
			r.OnGlobal(name, iface, version)
		}
		return nil
	case 1: // global_remove
		name, err := dec.Uint()
		if err != nil {
			return err
		}
		if r.OnGlobalRemove != nil {
			r.OnGlobalRemove(name)
		}
		return nil
	default:
		return fmt.Errorf("unknown opcode %d", opcode)
	}
}

// Callback is a wl_callback. The server destroys it after firing done.
type Callback struct {
	Proxy
	// OnDone is called with the event's callback data (for wl_display.sync
	// this is the event serial).
	OnDone func(data uint32)
}

func (*Callback) Interface() string { return "wl_callback" }

func (c *Callback) Dispatch(opcode uint16, dec *Decoder) error {
	switch opcode {
	case 0: // done
		data, err := dec.Uint()
		if err != nil {
			return err
		}
		if c.OnDone != nil {
			c.OnDone(data)
		}
		// wl_callback is destroyed by the server after done; the
		// delete_id event will remove it from the object map.
		return nil
	default:
		return fmt.Errorf("unknown opcode %d", opcode)
	}
}
