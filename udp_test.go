package osc

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestInvalidAddress(t *testing.T) {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server, err := ListenUDP("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	if err := server.Serve(Dispatcher{
		"/[": Method(func(msg Message) error {
			return nil
		}),
	}); err != ErrInvalidAddress {
		t.Fatal("expected invalid address error")
	}
}

func TestDialUDP(t *testing.T) {
	if _, err := DialUDP("asdfiauosweif", nil, nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDialUDPSetWriteBufferError(t *testing.T) {
	uc := &UDPConn{udpConn: errConn{}}
	_, err := uc.initialize()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := `setting write buffer size: derp`, err.Error(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestListenUDP(t *testing.T) {
	if _, err := ListenUDP("asdfiauosweif", nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDialUDPContext(t *testing.T) {
	raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DialUDPContext(context.Background(), "udp", laddr, raddr)
	if err != nil {
		t.Fatal(err)
	}
	ctxTimeout, cancel := context.WithTimeout(c.Context(), 20*time.Millisecond)
	defer cancel()
	c.SetContext(ctxTimeout)
	if c.Context() != ctxTimeout {
		t.Fatalf("expected %+v to be %+v", ctxTimeout, c.Context())
	}
	if err := c.Serve(Dispatcher{}); err != context.DeadlineExceeded {
		t.Fatalf("expected context.DeadlineExceeded, got %+v", err)
	}
}

// testUDPServer creates a server listening on an ephemeral port,
// initializes a client connection to that server, returns the server connection,
// the client connection, and a channel that emits the error returned from the server's Serve method.
// For clients that are interested in closing the server with an OSC
// message, a method is automatically added to the provided dispatcher
// at the "/server/close" address that closes the server.
func testUDPServer(t *testing.T, dispatcher Dispatcher) (*UDPConn, *UDPConn, chan error) {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server, err := ListenUDP("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	if dispatcher == nil {
		dispatcher = Dispatcher{}
	}
	dispatcher["/server/close"] = Method(func(msg Message) error {
		return server.Close()
	})
	errChan := make(chan error)

	go func() {
		errChan <- server.Serve(dispatcher)
	}()

	raddr, err := net.ResolveUDPAddr("udp", server.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn, err := DialUDP("udp", nil, raddr)
	if err != nil {
		t.Fatal(err)
	}
	return server, conn, errChan
}

func TestUDPConnSend_OK(t *testing.T) {
	_, conn, errChan := testUDPServer(t, nil)
	if err := conn.Send(Message{Address: "/server/close"}); err != nil {
		t.Fatal(err)
	}
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
}

// errConn is an implementation of the udpConn interface that returns errors from all it's methods.
type errConn struct {
	udpConn
}

func (e errConn) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	return 0, nil, errors.New("oops")
}

func (e errConn) SetWriteBuffer(bytes int) error {
	return errors.New("derp")
}

func TestUDPConnServe_ContextTimeout(t *testing.T) {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	server, err := ListenUDPContext(timeoutCtx, "udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	errChan := make(chan error)
	go func() {
		errChan <- server.Serve(Dispatcher{})
	}()
	select {
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout")
	case err := <-errChan:
		if err != context.DeadlineExceeded {
			t.Fatalf("expected context.DeadlineExceeded, got %+v", err)
		}
	}
}

func TestUDPConnServe_ReadError(t *testing.T) {
	errChan := make(chan error)

	// Setup the server.
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serverConn, err := ListenUDP("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	server := &UDPConn{
		udpConn: errConn{udpConn: serverConn},
		ctx:     context.Background(),
	}
	go func() {
		errChan <- server.Serve(Dispatcher{
			"/close": Method(func(msg Message) error {
				return server.Close()
			}),
		})
	}()

	// Setup the client.
	raddr, err := net.ResolveUDPAddr("udp", server.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn, err := DialUDP("udp", nil, raddr)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Send(Message{Address: "/close"}); err != nil {
		t.Fatal(err)
	}
	if err := <-errChan; err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUDPConnServe_NilDispatcher(t *testing.T) {
	// Setup the server.
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server, err := ListenUDP("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	if err := server.Serve(nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUDPConnServe_BadInboundAddr(t *testing.T) {
	for i, packet := range []Packet{
		Message{Address: "/["},
		Message{Address: "["},
		badPacket{},
	} {
		// Send a message with a bad address.
		_, conn, errChan := testUDPServer(t, Dispatcher{
			"/foo": Method(func(msg Message) error {
				return nil
			}),
		})
		if err := conn.Send(packet); err != nil {
			t.Fatal(err)
		}
		if err := <-errChan; err == nil {
			t.Fatalf("(packet %d) expected error, got nil", i)
		}
	}
}

func TestUDPConnSendTo(t *testing.T) {
	_, conn, errChan := testUDPServer(t, nil)
	laddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn2, err := ListenUDP("udp", laddr2)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn2.SendTo(conn.RemoteAddr(), badPacket{}); err != nil {
		t.Fatal(err)
	}
	if err := <-errChan; err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUDPConnSendBundle(t *testing.T) {
	b := Bundle{
		Timetag: FromTime(time.Now()),
		Packets: []Packet{
			Message{Address: "/server/close"},
		},
	}
	_, conn, errChan := testUDPServer(t, nil)
	if err := conn.Send(b); err != nil {
		t.Fatal(err)
	}
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
}

func TestUDPConnSendBundle_BadTypetag(t *testing.T) {
	_, conn, errChan := testUDPServer(t, nil)
	if err := conn.Send(badBundle{}); err != nil {
		t.Fatal(err)
	}
	err := <-errChan
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	expected, got := `error serving udp: read packets: read packet: parse message from packet: parse message: read argument 0: typetag "Q": invalid type tag`, err.Error()
	if expected != got {
		t.Fatal(err)
	}
}

func TestUDPConnSendBundle_DispatchError(t *testing.T) {
	b := Bundle{
		Timetag: FromTime(time.Now()),
		Packets: []Packet{
			Message{Address: "/foo"},
		},
	}
	_, conn, errChan := testUDPServer(t, Dispatcher{
		"/foo": Method(func(msg Message) error {
			return errors.New("oops")
		}),
	})
	if err := conn.Send(b); err != nil {
		t.Fatal(err)
	}
	err := <-errChan
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	expected, got := `error serving udp: dispatch bundle: oops`, err.Error()
	if expected != got {
		t.Fatal(err)
	}
}

// badPacket is a Packet that returns an OSC message with typetag 'Q'
type badPacket struct{}

func (bp badPacket) Bytes() []byte {
	return bytes.Join(
		[][]byte{
			{'/', 'f', 'o', 'o', 0, 0, 0, 0},
			{TypetagPrefix, 'Q', 0, 0},
		},
		[]byte{},
	)
}

func (bp badPacket) Equal(other Packet) bool {
	return false
}

type badBundle struct{}

func (bb badBundle) Bytes() []byte {
	msg := badPacket{}.Bytes()
	return bytes.Join(
		[][]byte{
			ToBytes(BundleTag),
			FromTime(time.Now()).Bytes(),
			Int(len(msg)).Bytes(),
			msg,
		},
		[]byte{},
	)
}

func (bb badBundle) Equal(other Packet) bool {
	return false
}
