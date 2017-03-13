package osc

import (
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func tmpListener(t *testing.T, dispatcher Dispatcher) (*UnixConn, chan error) {
	addr, err := net.ResolveUnixAddr("unixgram", TempSocket())
	if err != nil {
		t.Fatal(err)
	}
	conn, err := ListenUnix("unixgram", addr)
	if err != nil {
		t.Fatal(err)
	}
	errChan := make(chan error)
	go func() {
		if err := conn.Serve(1, dispatcher); err != nil {
			errChan <- err
		}
		close(errChan)
	}()
	return conn, errChan
}

func TestUnixSend(t *testing.T) {
	fooch := make(chan struct{})

	server, errChan := tmpListener(t, Dispatcher{
		"/foo": Method(func(m Message) error {
			close(fooch)
			return nil
		}),
	})
	addr, err := net.ResolveUnixAddr("unixgram", server.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn, err := DialUnix("unixgram", nil, addr)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Send(Message{Address: "/foo"}); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-errChan:
		t.Fatal(err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	case <-fooch:
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
}

func TestDialUnixBadNetwork(t *testing.T) {
	addr, err := net.ResolveUnixAddr("unixgram", TempSocket())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DialUnix("foo", nil, addr); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListenBadNetwork(t *testing.T) {
	addr, err := net.ResolveUnixAddr("unixgram", TempSocket())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ListenUnix("foo", addr); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type errUnixConn struct {
	unixConn
}

func (e errUnixConn) Close() error {
	return errors.New("bork")
}

func (e errUnixConn) ReadFromUnix(b []byte) (int, *net.UnixAddr, error) {
	return 0, nil, errors.New("oops")
}

func (e errUnixConn) SetWriteBuffer(bytes int) error {
	return errors.New("derp")
}

func TestDialUnixSetWriteBufferError(t *testing.T) {
	uc := &UnixConn{unixConn: errUnixConn{}}
	_, err := uc.initialize()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := `setting write buffer size: derp`, err.Error(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestUnixSendTo(t *testing.T) {
	fooch := make(chan struct{})

	server, errChan := tmpListener(t, Dispatcher{
		"/foo": Method(func(m Message) error {
			close(fooch)
			return nil
		}),
	})
	laddr, err := net.ResolveUnixAddr("unixgram", TempSocket())
	if err != nil {
		t.Fatal(err)
	}
	conn, err := ListenUnix("unixgram", laddr)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.SendTo(server.LocalAddr(), Message{Address: "/foo"}); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-errChan:
		t.Fatal(err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	case <-fooch:
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
}
