package osc

import (
	"context"
	"net"

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

// initialize initializes the connection.
func (conn *UnixConn) initialize() (*UnixConn, error) {
	if err := conn.unixConn.SetWriteBuffer(bufSize); err != nil {
		return nil, errors.Wrap(err, "setting write buffer size")
	}
	return conn, nil
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
	// TODO
	return nil
}
