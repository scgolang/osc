package osc

import "errors"

// Common errors.
var (
	ErrInvalidAddress = errors.New("invalid OSC address")
)

// Method is an OSC method
type Method func(msg *Message) error

// Dispatcher dispatches OSC packets.
type Dispatcher map[string]Method

// DispatchMessage dispatches OSC message.
func (d Dispatcher) DispatchMessage(msg *Message) error {
	for address, method := range d {
		matched, err := msg.Match(address)
		if err != nil {
			return err
		}
		if matched {
			return method(msg)
		}
	}
	return nil
}

// DispatchBundle dispatches an OSC bundle.
func (d Dispatcher) DispatchBundle(bun *Bundle) error {
	for address, method := range d {
		if err := bun.Invoke(address, method); err != nil {
			return err
		}
	}
	return nil
}
