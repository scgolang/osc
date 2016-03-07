package osc

import (
	"bytes"
	"fmt"
	"net"
)

// TCPListener is an OSC server based on TCP.
type TCPListener struct {
	*net.TCPListener
}

// ListenTCP creates a new TCP server.
func ListenTCP(network string, laddr *net.TCPAddr) (*TCPListener, error) {
	listener, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, err
	}
	return &TCPListener{TCPListener: listener}, nil
}

// Serve starts dispatching OSC.
func (listener *TCPListener) Serve(dispatcher Dispatcher) error {
	if dispatcher == nil {
		return ErrNilDispatcher
	}

	for addr := range dispatcher {
		if err := validateAddress(addr); err != nil {
			return err
		}
	}

	for {
		if err := listener.serve(dispatcher); err != nil {
			return err
		}
		break
	}

	return nil
}

// serve retrieves OSC packets.
func (listener *TCPListener) serve(dispatcher Dispatcher) error {
	buf := make([]byte, 65536)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		for _, err := conn.Read(buf); true; _, err = conn.Read(buf) {
			if err != nil {
				return err
			}
			switch buf[0] {
			case messageChar:
				msg, err := parseMessage(buf)
				if err != nil {
					fmt.Printf("========> error parsing message %s\n", err)
					return err
				}
				go func() { _ = dispatcher.DispatchMessage(msg) }()
			case bundleChar:
				bundle, err := parseBundle(buf)
				if err != nil {
					return err
				}
				go func() { _ = dispatcher.DispatchBundle(bundle) }()
			default:
				return ErrParse
			}
		}
	}
}

// TCPConn is an OSC connection over TCP.
type TCPConn struct {
	*net.TCPConn
}

// DialTCP creates a new OSC connection over TCP.
func DialTCP(network string, laddr, raddr *net.TCPAddr) (*TCPConn, error) {
	conn, err := net.DialTCP(network, laddr, raddr)
	if err != nil {
		return nil, err
	}
	return &TCPConn{TCPConn: conn}, nil
}

// Send sends an OSC message over UDP.
func (conn *TCPConn) Send(p Packet) (int64, error) {
	buf := &bytes.Buffer{}
	if _, err := p.WriteTo(buf); err != nil {
		return 0, err
	}
	n, err := conn.Write(buf.Bytes())
	if err != nil {
		return 0, err
	}
	return int64(n), nil
}
