package osc

import (
	"strings"
	"time"

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
type Dispatcher map[string]MessageHandler

// Dispatch invokes an OSC bundle's messages.
func (d Dispatcher) Dispatch(b Bundle) error {
	var (
		now = time.Now()
		tt  = b.Timetag.Time()
	)
	if tt.Before(now) {
		return d.immediately(b)
	}
	<-time.After(tt.Sub(now))
	return d.immediately(b)
}

// immediately invokes an OSC bundle immediately.
func (d Dispatcher) immediately(b Bundle) error {
	for _, p := range b.Packets {
		errs := []string{}
		if err := d.invoke(p); err != nil {
			errs = append(errs, err.Error())
		}
		if len(errs) > 0 {
			return errors.New(strings.Join(errs, " and "))
		}
		return nil
	}
	return nil
}

// invoke invokes an OSC packet, which could be a message or a bundle of messages.
func (d Dispatcher) invoke(p Packet) error {
	switch x := p.(type) {
	case Message:
		return d.Invoke(x)
	case Bundle:
		return d.immediately(x)
	default:
		return errors.Errorf("unsupported type for dispatcher: %T", p)
	}
}

// Invoke invokes an OSC message.
func (d Dispatcher) Invoke(msg Message) error {
	for address, handler := range d {
		matched, err := msg.Match(address)
		if err != nil {
			return err
		}
		if matched {
			return handler.Handle(msg)
		}
	}
	return nil
}
