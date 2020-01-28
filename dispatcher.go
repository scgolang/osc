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
type Dispatcher interface {
	Dispatch(bundle Bundle, exactMatch bool) error
	Invoke(msg Message, exactMatch bool) error
}

// PatternMatching is a dispatcher that implements OSC 1.0 pattern matching.
// See http://opensoundcontrol.org/spec-1_0 "OSC Message Dispatching and Pattern Matching"
type PatternMatching map[string]MessageHandler

// Dispatch invokes an OSC bundle's messages.
func (h PatternMatching) Dispatch(b Bundle, exactMatch bool) error {
	var (
		now = time.Now()
		tt  = b.Timetag.Time()
	)
	if tt.Before(now) {
		return h.immediately(b, exactMatch)
	}
	<-time.After(tt.Sub(now))
	return h.immediately(b, exactMatch)
}

// immediately invokes an OSC bundle immediately.
func (h PatternMatching) immediately(b Bundle, exactMatch bool) error {
	for _, p := range b.Packets {
		errs := []string{}
		if err := h.invoke(p, exactMatch); err != nil {
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
func (h PatternMatching) invoke(p Packet, exactMatch bool) error {
	switch x := p.(type) {
	case Message:
		return h.Invoke(x, exactMatch)
	case Bundle:
		return h.immediately(x, exactMatch)
	default:
		return errors.Errorf("unsupported type for dispatcher: %T", p)
	}
}

// Invoke invokes an OSC message.
func (h PatternMatching) Invoke(msg Message, exactMatch bool) error {
	for address, handler := range h {
		matched, err := msg.Match(address, exactMatch)
		if err != nil {
			return err
		}
		if matched {
			return handler.Handle(msg)
		}
	}
	return nil
}
