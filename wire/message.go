// Package wire implements the client side of the Wayland wire protocol:
// connection management, message marshalling, file descriptor passing, and
// object lifetime. It contains no knowledge of any specific protocol beyond
// the wl_display, wl_registry, and wl_callback bootstrap interfaces;
// everything else is generated from protocol XML by internal/gen.
package wire

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Wayland messages use the native byte order of the machine.
var order = binary.NativeEndian

// headerSize is the size of a message header in bytes: object ID followed by
// (size << 16 | opcode).
const headerSize = 8

// maxMessageSize is the largest message the protocol can express: the size
// field is 16 bits.
const maxMessageSize = 1<<16 - 1

// Fixed is a Wayland 24.8 signed fixed-point number.
type Fixed int32

// FixedFromFloat converts a float64 to a Fixed, truncating toward zero.
func FixedFromFloat(f float64) Fixed { return Fixed(math.Round(f * 256)) }

// FixedFromInt converts an integer to a Fixed.
func FixedFromInt(i int32) Fixed { return Fixed(i << 8) }

// Float returns the value as a float64.
func (f Fixed) Float() float64 { return float64(f) / 256 }

// Int returns the integer part of the value, truncating toward negative
// infinity.
func (f Fixed) Int() int32 { return int32(f) >> 8 }

// pad returns n rounded up to the next multiple of 4.
func pad(n int) int { return (n + 3) &^ 3 }

// Encoder builds the body of a single message. The zero value is ready to
// use.
type Encoder struct {
	buf []byte
	fds []int
}

// Bytes returns the encoded message body. The slice aliases the encoder's
// internal buffer.
func (e *Encoder) Bytes() []byte { return e.buf }

// Fds returns the file descriptors queued by PutFd.
func (e *Encoder) Fds() []int { return e.fds }

// PutUint appends a uint argument.
func (e *Encoder) PutUint(v uint32) {
	e.buf = order.AppendUint32(e.buf, v)
}

// PutInt appends an int argument.
func (e *Encoder) PutInt(v int32) { e.PutUint(uint32(v)) }

// PutFixed appends a fixed argument.
func (e *Encoder) PutFixed(v Fixed) { e.PutUint(uint32(v)) }

// PutObject appends an object argument. Pass 0 for a null object.
func (e *Encoder) PutObject(id uint32) { e.PutUint(id) }

// PutString appends a string argument. Wayland strings may not contain NUL
// bytes; the encoding includes a NUL terminator and the length counts it.
func (e *Encoder) PutString(s string) {
	e.PutUint(uint32(len(s) + 1))
	e.buf = append(e.buf, s...)
	e.buf = append(e.buf, 0)
	for len(e.buf)%4 != 0 {
		e.buf = append(e.buf, 0)
	}
}

// PutNullString appends a null string argument (for allow-null strings).
func (e *Encoder) PutNullString() { e.PutUint(0) }

// PutArray appends an array argument.
func (e *Encoder) PutArray(b []byte) {
	e.PutUint(uint32(len(b)))
	e.buf = append(e.buf, b...)
	for len(e.buf)%4 != 0 {
		e.buf = append(e.buf, 0)
	}
}

// PutFd queues a file descriptor to be sent with the message as ancillary
// data. Fds occupy no space in the message body.
func (e *Encoder) PutFd(fd int) { e.fds = append(e.fds, fd) }

// Decoder reads arguments from the body of a single received message.
type Decoder struct {
	buf []byte
	off int
	// fds is the connection's queue of received file descriptors. Fd
	// arguments are consumed from the front in order.
	conn *Conn
}

// NewDecoder returns a Decoder over a raw message body. Decoders created
// this way cannot decode fd arguments; the connection creates its own
// decoders for dispatch. Intended for tests.
func NewDecoder(body []byte) *Decoder { return &Decoder{buf: body} }

func (d *Decoder) remaining() int { return len(d.buf) - d.off }

func (d *Decoder) take(n int) ([]byte, error) {
	if d.remaining() < n {
		return nil, fmt.Errorf("wire: message truncated: need %d bytes, have %d", n, d.remaining())
	}
	b := d.buf[d.off : d.off+n]
	d.off += n
	return b, nil
}

// Uint reads a uint argument.
func (d *Decoder) Uint() (uint32, error) {
	b, err := d.take(4)
	if err != nil {
		return 0, err
	}
	return order.Uint32(b), nil
}

// Int reads an int argument.
func (d *Decoder) Int() (int32, error) {
	v, err := d.Uint()
	return int32(v), err
}

// Fixed reads a fixed argument.
func (d *Decoder) Fixed() (Fixed, error) {
	v, err := d.Uint()
	return Fixed(v), err
}

// Object reads an object or new_id argument. 0 means null.
func (d *Decoder) Object() (uint32, error) { return d.Uint() }

// String reads a string argument. A null string is returned as ("", false).
func (d *Decoder) String() (string, bool, error) {
	n, err := d.Uint()
	if err != nil {
		return "", false, err
	}
	if n == 0 {
		return "", false, nil
	}
	b, err := d.take(pad(int(n)))
	if err != nil {
		return "", false, err
	}
	// n includes the NUL terminator.
	return string(b[:n-1]), true, nil
}

// Array reads an array argument. The returned slice aliases the read buffer
// and must be copied if retained.
func (d *Decoder) Array() ([]byte, error) {
	n, err := d.Uint()
	if err != nil {
		return nil, err
	}
	b, err := d.take(pad(int(n)))
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}

// Fd consumes the next received file descriptor from the connection's fd
// queue. The caller owns the returned fd and must close it.
func (d *Decoder) Fd() (int, error) {
	if d.conn == nil || len(d.conn.recvFds) == 0 {
		return -1, fmt.Errorf("wire: message expects a file descriptor but none was received")
	}
	fd := d.conn.recvFds[0]
	d.conn.recvFds = d.conn.recvFds[1:]
	return fd, nil
}
