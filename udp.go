package osc

import "net"

// UDPConn is an OSC connection over UDP.
type UDPConn struct {
	*net.UDPConn
}

// DialUDP creates a new OSC connection over UDP.
func DialUDP(network string, laddr, raddr *net.UDPAddr) (*UDPConn, error) {
	conn, err := net.DialUDP(network, laddr, raddr)
	if err != nil {
		return nil, err
	}
	return &UDPConn{UDPConn: conn}, nil
}

// ListenUDP creates a new UDP server.
func ListenUDP(network string, laddr *net.UDPAddr) (*UDPConn, error) {
	conn, err := net.ListenUDP(network, laddr)
	if err != nil {
		return nil, err
	}
	return &UDPConn{UDPConn: conn}, nil
}

// ListenMulticastUDP listens for OSC messages
// addressed to the multicast group gaddr on the
// interface ifi.
// See https://golang.org/pkg/net/#ListenMulticastUDP.
func ListenMulticastUDP(network string, ifi *net.Interface, gaddr *net.UDPAddr) (*UDPConn, error) {
	conn, err := net.ListenMulticastUDP(network, ifi, gaddr)
	if err != nil {
		return nil, err
	}
	return &UDPConn{UDPConn: conn}, nil
}

// Serve starts dispatching OSC.
func (conn *UDPConn) Serve(dispatcher Dispatcher) error {
	if dispatcher == nil {
		return ErrNilDispatcher
	}

	for addr := range dispatcher {
		if err := validateAddress(addr); err != nil {
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

	_, senderAddress, err := conn.UDPConn.ReadFromUDP(data)
	if err != nil {
		return err
	}

	switch data[0] {
	case messageChar:
		msg, err := parseMessage(data, senderAddress)
		if err != nil {
			return err
		}
		// TODO: handle error.
		go dispatcher.DispatchMessage(msg)
	case bundleChar:
		bun, err := parseBundle(data, senderAddress)
		if err != nil {
			return err
		}
		// TODO: handle error.
		go dispatcher.DispatchBundle(bun)
	default:
		return ErrParse
	}

	return nil
}

// Send sends an OSC message over UDP.
func (conn *UDPConn) Send(p Packet) error {
	contents, err := p.Contents()
	if err != nil {
		return err
	}
	if _, err := conn.UDPConn.Write(contents); err != nil {
		return err
	}
	return nil
}
