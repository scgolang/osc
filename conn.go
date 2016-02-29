package osc

import (
	"errors"
	"net"
	"strings"
)

// Common errors.
var (
	ErrNilDispatcher  = errors.New("nil dispatcher")
	ErrPrematureClose = errors.New("server cannot be closed before calling Listen")
)

// Conn defines the methods
type Conn interface {
	net.Conn
	Serve(Dispatcher) error
	Send(Packet) (int64, error)
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
