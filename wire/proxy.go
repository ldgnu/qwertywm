package wire

import "fmt"

// Object is the interface implemented by all protocol object proxies,
// including generated ones.
type Object interface {
	// ID returns the object's protocol ID, or 0 if the object has been
	// destroyed.
	ID() uint32
	// Interface returns the Wayland interface name, e.g. "wl_output".
	Interface() string
	// Dispatch decodes and handles the event with the given opcode. The
	// decoder is positioned at the start of the message body.
	Dispatch(opcode uint16, d *Decoder) error
}

// Proxy is the common embedded base for all protocol objects. Generated
// types embed it and the connection populates it at registration time.
type Proxy struct {
	id   uint32
	conn *Conn
}

// ID returns the object's protocol ID, or 0 after destruction.
func (p *Proxy) ID() uint32 { return p.id }

// Conn returns the connection the object belongs to.
func (p *Proxy) Conn() *Conn { return p.conn }

// init is called by the connection when the object is registered.
func (p *Proxy) init(c *Conn, id uint32) {
	p.conn = c
	p.id = id
}

// ObjectID is a convenience for marshalling nullable object arguments:
// it returns o.ID() or 0 if o is nil.
func ObjectID(o Object) uint32 {
	if o == nil {
		return 0
	}
	return o.ID()
}

// register assigns a new client-allocated ID to obj and adds it to the
// object map.
func (c *Conn) register(obj Object, p *Proxy) uint32 {
	id := c.nextID
	c.nextID++
	p.init(c, id)
	c.objects[id] = obj
	return id
}

// RegisterClient allocates a client-side ID for obj (whose embedded Proxy is
// p) and registers it. Generated request methods with new_id arguments call
// this before sending the request.
func (c *Conn) RegisterClient(obj Object, p *Proxy) uint32 {
	return c.register(obj, p)
}

// RegisterServer registers obj under a server-allocated ID received in an
// event's new_id argument. Generated event dispatch code calls this.
func (c *Conn) RegisterServer(obj Object, p *Proxy, id uint32) error {
	if _, exists := c.objects[id]; exists {
		return fmt.Errorf("wire: server allocated duplicate object id %d", id)
	}
	p.init(c, id)
	c.objects[id] = obj
	return nil
}

// Unregister removes an object from the object map. Generated destructor
// request methods call this after sending the destructor; the ID is not
// reused until the server confirms with wl_display.delete_id (we never reuse
// IDs, so confirmation just removes bookkeeping).
func (c *Conn) Unregister(obj Object) {
	delete(c.objects, obj.ID())
}

// Lookup returns the registered object with the given ID, or nil. Generated
// event dispatch code uses this to resolve object arguments.
func (c *Conn) Lookup(id uint32) Object {
	if id == 0 {
		return nil
	}
	return c.objects[id]
}
