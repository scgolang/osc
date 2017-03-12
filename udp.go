package osc

import (
	"context"
	"net"

	"github.com/pkg/errors"
)

// udpConn includes exactly the methods we need from *net.UDPConn
type udpConn interface {
	net.Conn
	netWriter

	ReadFromUDP([]byte) (int, *net.UDPAddr, error)
}

// UDPConn is an OSC connection over UDP.
type UDPConn struct {
	udpConn

	closeChan chan struct{}
	ctx       context.Context
	errChan   chan error
}

// DialUDP creates a new OSC connection over UDP.
func DialUDP(network string, laddr, raddr *net.UDPAddr) (*UDPConn, error) {
	return DialUDPContext(context.Background(), network, laddr, raddr)
}

// DialUDPContext returns a new OSC connection over UDP that can be canceled with the provided context.
func DialUDPContext(ctx context.Context, network string, laddr, raddr *net.UDPAddr) (*UDPConn, error) {
	conn, err := net.DialUDP(network, laddr, raddr)
	if err != nil {
		return nil, err
	}
	uc := &UDPConn{
		udpConn:   conn,
		closeChan: make(chan struct{}),
		ctx:       ctx,
		errChan:   make(chan error),
	}
	return uc.initialize()
}

// ListenUDP creates a new UDP server.
func ListenUDP(network string, laddr *net.UDPAddr) (*UDPConn, error) {
	return ListenUDPContext(context.Background(), network, laddr)
}

// ListenUDPContext creates a UDP listener that can be canceled with the provided context.
func ListenUDPContext(ctx context.Context, network string, laddr *net.UDPAddr) (*UDPConn, error) {
	conn, err := net.ListenUDP(network, laddr)
	if err != nil {
		return nil, err
	}
	uc := &UDPConn{
		udpConn:   conn,
		closeChan: make(chan struct{}),
		ctx:       ctx,
		errChan:   make(chan error),
	}
	return uc.initialize()
}

// Close closes the udp conn.
func (conn *UDPConn) Close() error {
	close(conn.closeChan)
	return conn.udpConn.Close()
}

// CloseChan returns a channel that is closed when the connection gets closed.
func (conn *UDPConn) CloseChan() <-chan struct{} {
	return conn.closeChan
}

// Context returns the context associated with the conn.
func (conn *UDPConn) Context() context.Context {
	return conn.ctx
}

// initialize initializes a UDP connection.
func (conn *UDPConn) initialize() (*UDPConn, error) {
	if err := conn.udpConn.SetWriteBuffer(bufSize); err != nil {
		return nil, errors.Wrap(err, "setting write buffer size")
	}
	return conn, nil

}

// read reads bytes and returns the net.Addr of the sender.
func (conn *UDPConn) read(data []byte) (int, net.Addr, error) {
	return conn.ReadFromUDP(data)
}

// Send sends an OSC message over UDP.
func (conn *UDPConn) Send(p Packet) error {
	_, err := conn.Write(p.Bytes())
	return err
}

// SendTo sends a packet to the given address.
func (conn *UDPConn) SendTo(addr net.Addr, p Packet) error {
	_, err := conn.WriteTo(p.Bytes(), addr)
	return err
}

// Serve starts dispatching OSC.
// Any errors returned from a dispatched method will be returned.
// Note that this means that errors returned from a dispatcher method will kill your server.
// If context.Canceled or context.DeadlineExceeded are encountered they will be returned directly.
func (conn *UDPConn) Serve(numWorkers int, dispatcher Dispatcher) error {
	return serve(conn, numWorkers, dispatcher)
}

// SetContext sets the context associated with the conn.
func (conn *UDPConn) SetContext(ctx context.Context) {
	conn.ctx = ctx
}
