package osc

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"strings"
)

const (
	readBufSize = 4096
)

// Common errors.
var (
	errBundle         = errors.New("message is a bundle")
	ErrNoDispatcher   = errors.New("no dispatcher defined")
	ErrPrematureClose = errors.New("server cannot be closed before calling Listen")
)

// Server is an OSC server.
type Server struct {
	Listening  chan struct{} // Listening is a channel used to indicate when the server is running.
	dispatcher oscDispatcher // Dispatcher that dispatches OSC packets/messages.
	conn       *net.UDPConn  // conn is a UDP connection object.
}

// NewServer returns a new OSC Server.
func NewServer(addr string, handlers map[string]Method) (*Server, error) {
	for addr, _ := range handlers {
		if err := validateAddress(addr); err != nil {
			return nil, err
		}
	}

	netAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", netAddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		Listening:  make(chan struct{}),
		dispatcher: oscDispatcher(handlers),
		conn:       conn,
	}, nil
}

// LocalAddr returns the local network address.
func (server *Server) LocalAddr() net.Addr {
	return server.conn.LocalAddr()
}

// Close stops the OSC server and closes the connection.
func (server *Server) Close() error {
	if server.conn == nil {
		return nil
	}
	return server.conn.Close()
}

// Listen retrieves incoming OSC packets and dispatches the retrieved OSC packets.
func (server *Server) Listen() error {
	if server.dispatcher == nil {
		return ErrNoDispatcher
	}

	server.Listening <- struct{}{}

	for {
		if err := server.serve(); err != nil {
			return err
		}
	}
}

// serve retrieves OSC packets.
func (server *Server) serve() error {
	data := make([]byte, readBufSize)

	_, senderAddress, err := server.conn.ReadFromUDP(data)
	if err != nil {
		return err
	}

	switch data[0] {
	case messageChar:
		msg, err := parseMessage(data, senderAddress)
		if err != nil {
			return err
		}
		return server.dispatcher.dispatchMessage(msg)
	case bundleChar:
		bun, err := parseBundle(data, senderAddress)
		if err != nil {
			return err
		}
		return server.dispatcher.dispatchBundle(bun)
	default:
		return ErrParse
	}

	return nil
}

// Send sends an OSC Bundle or an OSC Message.
func (server *Server) SendTo(addr net.Addr, r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		server.conn.Close()
		return err
	}
	if _, err := server.conn.WriteTo(data, addr); err != nil {
		return server.conn.Close()
	}
	return nil
}

var invalidAddressRunes = []rune{'*', '?', ',', '[', ']', '{', '}', '#', ' '}

func validateAddress(addr string) error {
	for _, chr := range invalidAddressRunes {
		if strings.ContainsRune(addr, chr) {
			return ErrInvalidAddress
		}
	}
	return nil
}
