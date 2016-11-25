package osc

import (
	"io"
	"net"

	"github.com/pkg/errors"
)

// udpConn includes exactly the methods we need from *net.UDPConn
type udpConn interface {
	io.WriteCloser

	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	ReadFromUDP([]byte) (int, *net.UDPAddr, error)
	WriteTo([]byte, net.Addr) (int, error)
}

// UDPConn is an OSC connection over UDP.
type UDPConn struct {
	udpConn
}

// DialUDP creates a new OSC connection over UDP.
func DialUDP(network string, laddr, raddr *net.UDPAddr) (*UDPConn, error) {
	conn, err := net.DialUDP(network, laddr, raddr)
	if err != nil {
		return nil, err
	}
	return &UDPConn{udpConn: conn}, nil
}

// ListenUDP creates a new UDP server.
func ListenUDP(network string, laddr *net.UDPAddr) (*UDPConn, error) {
	conn, err := net.ListenUDP(network, laddr)
	if err != nil {
		return nil, err
	}
	return &UDPConn{udpConn: conn}, nil
}

// Serve starts dispatching OSC.
func (conn *UDPConn) Serve(dispatcher Dispatcher) error {
	if dispatcher == nil {
		return ErrNilDispatcher
	}

	for addr := range dispatcher {
		if err := ValidateAddress(addr); err != nil {
			return err
		}
	}

	for {
		if err := conn.serve(dispatcher); err != nil {
			return err
		}
		break
	}

	return nil
}

// serve retrieves OSC packets.
func (conn *UDPConn) serve(dispatcher Dispatcher) error {
	data := make([]byte, readBufSize)

	_, senderAddress, err := conn.ReadFromUDP(data)
	if err != nil {
		return err
	}

	switch data[0] {
	// TODO: handle bundle
	case MessageChar:
		msg, err := ParseMessage(data, senderAddress)
		if err != nil {
			return err
		}
		if err := dispatcher.Dispatch(msg); err != nil {
			return errors.Wrap(err, "dispatch message")
		}
	default:
		return ErrParse
	}

	return nil
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
