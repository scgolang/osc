package osc

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

const (
	readBufSize = 16384
)

// Common errors.
var (
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
func NewServer(addr string, handlers map[string]HandlerFunc) (*Server, error) {
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
func (self *Server) LocalAddr() net.Addr {
	return self.conn.LocalAddr()
}

// Close stops the OSC server and closes the connection.
func (self *Server) Close() error {
	if self.conn == nil {
		return nil
	}
	return self.conn.Close()
}

// Listen retrieves incoming OSC packets and dispatches the retrieved OSC packets.
func (self *Server) Listen() error {
	if self.dispatcher == nil {
		return ErrNoDispatcher
	}

	self.Listening <- struct{}{}

	msg, err := self.readFromConnection()
	if err != nil {
		return err
	}
	self.dispatcher.dispatch(msg)

	return nil
}

// Send sends an OSC Bundle or an OSC Message.
func (self *Server) SendTo(addr net.Addr, packet Packet) (err error) {
	data, err := packet.ToByteArray()
	if err != nil {
		self.conn.Close()
		return err
	}

	written, err := self.conn.WriteTo(data, addr)
	if err != nil {
		fmt.Println("could not write packet")
		self.conn.Close()
		return err
	}
	if written != len(data) {
		errmsg := "only wrote %d bytes of osc packet with length %d"
		return fmt.Errorf(errmsg, written, len(data))
	}

	return nil
}

// readFromConnection retrieves OSC packets.
func (self *Server) readFromConnection() (packet Packet, err error) {
	data := make([]byte, readBufSize)

	_, senderAddress, err := self.conn.ReadFromUDP(data)
	if err != nil {
		return nil, err
	}

	switch data[0] {
	case messageChar:
		return parseMessage(data, senderAddress)
	case bundleChar:
		return parseBundle(data, senderAddress)
	default:
		return nil, ErrParse
	}

	return packet, nil
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
