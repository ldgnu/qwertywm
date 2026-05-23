package wire

import (
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

// fakeServer is the compositor end of a socketpair. A nearly identical
// helper exists in the wiretest package for tests outside this package;
// this one is duplicated here because an internal test file cannot import
// a package that imports wire.
type fakeServer struct {
	t    *testing.T
	Sock *net.UnixConn
}

// newPair returns a connected client Conn and fake server.
func newPair(t *testing.T) (*Conn, *fakeServer) {
	t.Helper()
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	if err != nil {
		t.Fatalf("socketpair: %v", err)
	}
	toConn := func(fd int, name string) *net.UnixConn {
		f := os.NewFile(uintptr(fd), name)
		defer f.Close()
		c, err := net.FileConn(f)
		if err != nil {
			t.Fatalf("FileConn: %v", err)
		}
		return c.(*net.UnixConn)
	}
	client := NewConn(toConn(fds[0], "client"))
	server := &fakeServer{t: t, Sock: toConn(fds[1], "server")}
	t.Cleanup(func() {
		client.Close()
		server.Sock.Close()
	})
	return client, server
}

// send writes a raw event to the client.
func send(s *fakeServer, objID uint32, opcode uint16, body []byte, fds ...int) {
	s.t.Helper()
	size := headerSize + len(body)
	msg := make([]byte, 0, size)
	msg = order.AppendUint32(msg, objID)
	msg = order.AppendUint32(msg, uint32(size)<<16|uint32(opcode))
	msg = append(msg, body...)
	var oob []byte
	if len(fds) > 0 {
		oob = syscall.UnixRights(fds...)
	}
	if _, _, err := s.Sock.WriteMsgUnix(msg, oob, nil); err != nil {
		s.t.Fatalf("server send: %v", err)
	}
}

// recv reads one message from the client and returns its header and body.
func recv(s *fakeServer) (objID uint32, opcode uint16, body []byte) {
	s.t.Helper()
	s.Sock.SetReadDeadline(time.Now().Add(5 * time.Second))
	hdr := make([]byte, headerSize)
	if _, err := readFull(s.Sock, hdr); err != nil {
		s.t.Fatalf("server read header: %v", err)
	}
	objID = order.Uint32(hdr)
	sizeOp := order.Uint32(hdr[4:])
	size := int(sizeOp >> 16)
	body = make([]byte, size-headerSize)
	if _, err := readFull(s.Sock, body); err != nil {
		s.t.Fatalf("server read body: %v", err)
	}
	return objID, uint16(sizeOp & 0xffff), body
}

func readFull(c *net.UnixConn, b []byte) (int, error) {
	n := 0
	for n < len(b) {
		m, err := c.Read(b[n:])
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func TestGetRegistryAndGlobals(t *testing.T) {
	client, server := newPair(t)

	reg := client.Display.GetRegistry()
	if reg.ID() != 2 {
		t.Fatalf("registry id = %d, want 2", reg.ID())
	}
	var globals []string
	reg.OnGlobal = func(name uint32, iface string, version uint32) {
		globals = append(globals, iface)
	}
	if err := client.Flush(); err != nil {
		t.Fatal(err)
	}

	// Server sees get_registry (opcode 1 on object 1).
	objID, opcode, body := recv(server)
	if objID != 1 || opcode != 1 {
		t.Fatalf("server got %d.%d, want 1.1", objID, opcode)
	}
	d := Decoder{buf: body}
	if id, _ := d.Uint(); id != 2 {
		t.Fatalf("get_registry new_id = %d, want 2", id)
	}

	// Server announces two globals.
	for i, iface := range []string{"river_window_manager_v1", "wl_output"} {
		e := &Encoder{}
		e.PutUint(uint32(i + 1))
		e.PutString(iface)
		e.PutUint(4)
		send(server, 2, 0, e.buf)
	}
	if _, err := client.Dispatch(); err != nil {
		t.Fatal(err)
	}
	for len(globals) < 2 {
		if _, err := client.Dispatch(); err != nil {
			t.Fatal(err)
		}
	}
	if globals[0] != "river_window_manager_v1" || globals[1] != "wl_output" {
		t.Errorf("globals = %v", globals)
	}
}

func TestRoundTrip(t *testing.T) {
	client, server := newPair(t)

	// The server side: respond to sync with done + delete_id.
	go func() {
		objID, opcode, body := recv(server)
		if objID != 1 || opcode != 0 {
			t.Errorf("server got %d.%d, want 1.0 (sync)", objID, opcode)
			return
		}
		d := Decoder{buf: body}
		cbID, _ := d.Uint()
		// done
		e := &Encoder{}
		e.PutUint(12345)
		send(server, cbID, 0, e.buf)
		// delete_id
		e = &Encoder{}
		e.PutUint(cbID)
		send(server, 1, 1, e.buf)
	}()

	if err := client.RoundTrip(); err != nil {
		t.Fatal(err)
	}
	// The callback object must have been cleaned up by delete_id.
	for len(client.objects) > 1 {
		if _, err := client.DispatchPending(); err != nil {
			t.Fatal(err)
		}
		if len(client.objects) > 1 {
			if _, err := client.Dispatch(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestServerError(t *testing.T) {
	client, server := newPair(t)
	e := &Encoder{}
	e.PutObject(1)
	e.PutUint(2)
	e.PutString("you broke it")
	send(server, 1, 0, e.buf)
	_, err := client.Dispatch()
	if err == nil {
		t.Fatal("dispatching wl_display.error did not return an error")
	}
	if client.Err() == nil {
		t.Fatal("connection error not recorded")
	}
	// All further operations fail fast.
	if err := client.Flush(); err == nil {
		t.Error("Flush after fatal error did not fail")
	}
}

func TestFdPassing(t *testing.T) {
	client, server := newPair(t)

	// A test object that expects an event carrying an fd.
	var gotFd int = -1
	obj := &testObject{
		iface: "test_fd_receiver",
		dispatch: func(opcode uint16, d *Decoder) error {
			fd, err := d.Fd()
			if err != nil {
				return err
			}
			gotFd = fd
			return nil
		},
	}
	client.RegisterClient(obj, &obj.Proxy)

	// Server sends an event with an fd pointing at a pipe with known
	// contents.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	w.WriteString("hello from the fd")
	send(server, obj.ID(), 0, nil, int(r.Fd()))

	if _, err := client.Dispatch(); err != nil {
		t.Fatal(err)
	}
	if gotFd < 0 {
		t.Fatal("no fd received")
	}
	defer syscall.Close(gotFd)
	buf := make([]byte, 64)
	n, err := syscall.Read(gotFd, buf)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf[:n]) != "hello from the fd" {
		t.Errorf("read %q through the passed fd", buf[:n])
	}
}

func TestSendFd(t *testing.T) {
	client, server := newPair(t)
	obj := &testObject{iface: "test_fd_sender"}
	client.RegisterClient(obj, &obj.Proxy)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	e := &Encoder{}
	e.PutFd(int(w.Fd()))
	e.PutUint(99)
	client.SendRequest(obj, 3, e)
	if err := client.Flush(); err != nil {
		t.Fatal(err)
	}

	// Server receives the message and the fd.
	oob := make([]byte, 128)
	buf := make([]byte, 64)
	server.Sock.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, oobn, _, _, err := server.Sock.ReadMsgUnix(buf, oob)
	if err != nil {
		t.Fatal(err)
	}
	if n != headerSize+4 {
		t.Fatalf("server read %d bytes, want %d", n, headerSize+4)
	}
	cmsgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil || len(cmsgs) != 1 {
		t.Fatalf("control messages: %v %v", cmsgs, err)
	}
	fds, err := syscall.ParseUnixRights(&cmsgs[0])
	if err != nil || len(fds) != 1 {
		t.Fatalf("unix rights: %v %v", fds, err)
	}
	defer syscall.Close(fds[0])
	// Write through the received fd and read it back from our pipe.
	if _, err := syscall.Write(fds[0], []byte("ping")); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, 4)
	if _, err := r.Read(got); err != nil {
		t.Fatal(err)
	}
	if string(got) != "ping" {
		t.Errorf("read %q through the round-tripped fd", got)
	}
}

func TestPartialMessageReassembly(t *testing.T) {
	client, server := newPair(t)
	var got []string
	reg := client.Display.GetRegistry()
	reg.OnGlobal = func(_ uint32, iface string, _ uint32) { got = append(got, iface) }
	client.Flush()
	recv(server)

	// Build two global events and send them split at an awkward boundary.
	var stream []byte
	for i, iface := range []string{"alpha", "beta"} {
		e := &Encoder{}
		e.PutUint(uint32(i))
		e.PutString(iface)
		e.PutUint(1)
		size := headerSize + len(e.buf)
		stream = order.AppendUint32(stream, reg.ID())
		stream = order.AppendUint32(stream, uint32(size)<<16|0)
		stream = append(stream, e.buf...)
	}
	// First write ends mid-way through the second message's header.
	split := len(stream) - 13
	server.Sock.Write(stream[:split])
	// Give the client a chance to read the partial data.
	if _, err := client.Dispatch(); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "alpha" {
		t.Fatalf("after partial write got %v, want [alpha]", got)
	}
	server.Sock.Write(stream[split:])
	if _, err := client.Dispatch(); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1] != "beta" {
		t.Fatalf("after completing the write got %v, want [alpha beta]", got)
	}
}

// testObject is a minimal Object implementation for wire-level tests.
type testObject struct {
	Proxy
	iface    string
	dispatch func(opcode uint16, d *Decoder) error
}

func (o *testObject) Interface() string { return o.iface }
func (o *testObject) Dispatch(opcode uint16, d *Decoder) error {
	if o.dispatch == nil {
		return nil
	}
	return o.dispatch(opcode, d)
}
