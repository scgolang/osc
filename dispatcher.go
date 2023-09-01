package osc

import (
	"github.com/pkg/errors"
)

// Common errors.
var (
	ErrInvalidAddress = errors.New("invalid OSC address")
)

// Method is an OSC method
type Method func(msg Message) error

// Handle handles an OSC message.
func (method Method) Handle(m Message) error {
	return method(m)
}

// MessageHandler is any type that can handle an OSC message.
type MessageHandler interface {
	Handle(Message) error
}

// Dispatcher dispatches OSC packets.
type Dispatcher interface {
	Dispatch(bundle Bundle, exactMatch bool) error
	Invoke(msg Message, exactMatch bool) error
}
