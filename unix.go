package osc

import (
	"context"
	"net"
	"os"
	"path/filepath"

	ulid "github.com/imdario/go-ulid"
	"github.com/pkg/errors"
)

type unixConn interface {
	net.Conn
	netWriter

	ReadFromUnix([]byte) (int, *net.UnixAddr, error)
}

// UnixConn handles OSC over a unix socket.
type UnixConn struct {
	unixConn

	closeChan chan struct{}
	ctx       context.Context
	errChan   chan error
}

// DialUnix opens a unix socket for OSC communication.
func DialUnix(network string, laddr, raddr *net.UnixAddr) (*UnixConn, error) {
	return DialUnixContext(context.Background(), network, laddr, raddr)
}

// DialUnixContext creates a new UnixConn.
func DialUnixContext(ctx context.Context, network string, laddr, raddr *net.UnixAddr) (*UnixConn, error) {
	conn, err := net.DialUnix(network, laddr, raddr)
	if err != nil {
		return nil, err
	}
	uc := &UnixConn{
		unixConn:  conn,
		closeChan: make(chan struct{}),
		ctx:       ctx,
		errChan:   make(chan error),
	}
	return uc.initialize()
}

// ListenUnix creates a Unix listener that can be canceled with the provided context.
func ListenUnix(network string, laddr *net.UnixAddr) (*UnixConn, error) {
	return ListenUnixContext(context.Background(), network, laddr)
}

// ListenUnixContext creates a Unix listener that can be canceled with the provided context.
func ListenUnixContext(ctx context.Context, network string, laddr *net.UnixAddr) (*UnixConn, error) {
	conn, err := net.ListenUnixgram(network, laddr)
	if err != nil {
		return nil, err
	}
	uc := &UnixConn{
		unixConn:  conn,
		closeChan: make(chan struct{}),
		ctx:       ctx,
		errChan:   make(chan error),
	}
	return uc.initialize()
}

// Close closes the connection.
func (conn *UnixConn) Close() error {
	close(conn.closeChan)
	return conn.unixConn.Close()
}

// CloseChan returns a channel that is closed when the connection gets closed.
func (conn *UnixConn) CloseChan() <-chan struct{} {
	return conn.closeChan
}

// Context returns the context for the unix conn.
func (conn *UnixConn) Context() context.Context {
	return conn.ctx
}

// initialize initializes the connection.
func (conn *UnixConn) initialize() (*UnixConn, error) {
	if err := conn.unixConn.SetWriteBuffer(bufSize); err != nil {
		return nil, errors.Wrap(err, "setting write buffer size")
	}
	return conn, nil
}

func (conn *UnixConn) read(data []byte) (int, net.Addr, error) {
	return conn.ReadFromUnix(data)
}

// Send sends a Packet.
func (conn *UnixConn) Send(p Packet) error {
	_, err := conn.Write(p.Bytes())
	return err
}

// SendTo sends a Packet to the provided net.Addr.
func (conn *UnixConn) SendTo(addr net.Addr, p Packet) error {
	_, err := conn.WriteTo(p.Bytes(), addr)
	return err
}

// Serve starts dispatching OSC.
// Any errors returned from a dispatched method will be returned.
// Note that this means that errors returned from a dispatcher method will kill your server.
// If context.Canceled or context.DeadlineExceeded are encountered they will be returned directly.
func (conn *UnixConn) Serve(numWorkers int, dispatcher Dispatcher) error {
	return serve(conn, numWorkers, dispatcher)
}

// TempSocket creates an absolute path to a temporary socket file.
func TempSocket() string {
	return filepath.Join(os.TempDir(), ulid.New().String()) + ".sock"
}
