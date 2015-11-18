package osc

import (
	"errors"
	"net"
	"strings"
)

const (
	readBufSize = 4096
)

// Common errors.
var (
	errBundle         = errors.New("message is a bundle")
	ErrNilDispatcher  = errors.New("nil dispatcher")
	ErrPrematureClose = errors.New("server cannot be closed before calling Listen")
	networkTCP        = "tcp"
	networkUDP        = "udp"
)

// Conn defines the methods
type Conn interface {
	net.Conn
	Serve(Dispatcher) error
	Send(*Message) error
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
