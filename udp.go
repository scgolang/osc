package osc

import (
	"context"
	"net"

	"github.com/pkg/errors"
)

// udpConn includes exactly the methods we need from *net.UDPConn
type udpConn interface {
	net.Conn

	ReadFromUDP([]byte) (int, *net.UDPAddr, error)
	WriteTo([]byte, net.Addr) (int, error)
}

// UDPConn is an OSC connection over UDP.
type UDPConn struct {
	udpConn
	closeChan chan struct{}
	ctx       context.Context
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
	return &UDPConn{
		udpConn:   conn,
		closeChan: make(chan struct{}),
		ctx:       ctx,
	}, nil
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
	return &UDPConn{
		udpConn:   conn,
		closeChan: make(chan struct{}),
		ctx:       ctx,
	}, nil
}

// Context returns the context associated with the conn.
func (conn *UDPConn) Context() context.Context {
	return conn.ctx
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
func (conn *UDPConn) Serve(dispatcher Dispatcher) error {
	if dispatcher == nil {
		return ErrNilDispatcher
	}

	for addr := range dispatcher {
		if err := ValidateAddress(addr); err != nil {
			return err
		}
	}

	errChan := make(chan error)

	go func() {
		for {
			if err := conn.serve(dispatcher); err != nil {
				errChan <- err
			}
		}
	}()

	// If the connection is closed or the context is canceled then stop serving.
	select {
	case err := <-errChan:
		return errors.Wrap(err, "error serving udp")
	case <-conn.closeChan:
	case <-conn.ctx.Done():
		return conn.ctx.Err()
	}
	return nil
}

// serve retrieves OSC packets.
func (conn *UDPConn) serve(dispatcher Dispatcher) error {
	data := make([]byte, readBufSize)

	_, sender, err := conn.ReadFromUDP(data)
	if err != nil {
		return err
	}

	switch data[0] {
	case BundleTag[0]:
		bundle, err := ParseBundle(data, sender)
		if err != nil {
			return err
		}
		if err := dispatcher.Dispatch(bundle); err != nil {
			return errors.Wrap(err, "dispatch bundle")
		}
	case MessageChar:
		msg, err := ParseMessage(data, sender)
		if err != nil {
			return err
		}
		if err := dispatcher.Invoke(msg); err != nil {
			return errors.Wrap(err, "dispatch message")
		}
	default:
		return ErrParse
	}

	return nil
}

// SetContext sets the context associated with the conn.
func (conn *UDPConn) SetContext(ctx context.Context) {
	conn.ctx = ctx
}

// Close closes the udp conn.
func (conn *UDPConn) Close() error {
	close(conn.closeChan)
	return conn.udpConn.Close()
}
